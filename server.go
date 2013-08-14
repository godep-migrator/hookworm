package hookworm

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	addrFlag = flag.String("a", ":9988", "Server address")

	emailFlag           = flag.String("e", "smtp://localhost:25", "Email server address")
	emailFromFlag       = flag.String("f", "hookworm@localhost", "Email from address")
	emailRcptsFlag      = flag.String("r", "", "Email recipients (comma-delimited)")
	watchedBranchesFlag = flag.String("b", "", "Watched branches (comma-delimited regexes)")
	watchedPathsFlag    = flag.String("p", "", "Watched paths (comma-delimited regexes)")
	wormDirFlag         = flag.String("W", "", "Worm directory that contains handler executables")
	wormTimeoutFlag     = flag.Int("T", 30, "Timeout for handler executables (in seconds)")
	workingDirFlag      = flag.String("D", "", "Working directory (scratch pad)")

	useSyslogFlag = flag.Bool("S", false, "Send all received events to syslog")

	githubPathFlag = flag.String("github.path", "/github", "Path to handle Github payloads")
	travisPathFlag = flag.String("travis.path", "/travis", "Path to handle Travis payloads")

	pidFileFlag       = flag.String("P", "", "PID file (only written if flag given)")
	debugFlag         = flag.Bool("d", false, "Show debug output")
	printRevisionFlag = flag.Bool("rev", false, "Print revision and exit")
	printVersionFlag  = flag.Bool("version", false, "Print version and exit")

	logTimeFmt = "2/Jan/2006:15:04:05 -0700" // "%d/%b/%Y:%H:%M:%S %z"
)

// Server implements ServeHTTP, parsing payloads and handing them off to the
// handler pipeline
type Server struct {
	pipeline   Handler
	debug      bool
	githubPath string
	travisPath string
}

// ServerMain is the `main` entry point used by the `hookworm-server`
// executable
func ServerMain() int {
	flag.Usage = func() {
		fmt.Printf("Usage: %v [options]\n", progName)
		flag.PrintDefaults()
	}

	flag.Parse()
	if *printVersionFlag {
		printVersion()
		return 0
	}

	if *printRevisionFlag {
		printRevision()
		return 0
	}

	log.Println("Starting", progVersion())

	workingDir, err := getWorkingDir(*workingDirFlag)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return 1
	}

	log.Println("Using working directory", workingDir)

	defer os.RemoveAll(workingDir)

	cfg := &HandlerConfig{
		Debug:           *debugFlag,
		EmailFromAddr:   *emailFromFlag,
		EmailRcpts:      commaSplit(*emailRcptsFlag),
		EmailUri:        *emailFlag,
		GithubPath:      *githubPathFlag,
		ServerAddress:   *addrFlag,
		ServerPidFile:   *pidFileFlag,
		TravisPath:      *travisPathFlag,
		UseSyslog:       *useSyslogFlag,
		WatchedBranches: commaSplit(*watchedBranchesFlag),
		WatchedPaths:    commaSplit(*watchedPathsFlag),
		WorkingDir:      workingDir,
		WormDir:         *wormDirFlag,
		WormTimeout:     *wormTimeoutFlag,
	}

	if cfg.Debug {
		log.Printf("Using handler config: %+v\n", cfg)
	}

	if err := os.Chdir(cfg.WorkingDir); err != nil {
		log.Fatalf("Failed to move into working directory %v\n", cfg.WorkingDir)
	}

	server, err := NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", server)
	log.Printf("Listening on %v\n", cfg.ServerAddress)

	if len(cfg.ServerPidFile) > 0 {
		pidFile, err := os.Create(cfg.ServerPidFile)
		if err != nil {
			log.Fatal("Failed to open PID file:", err)
		}
		fmt.Fprintf(pidFile, "%d\n", os.Getpid())
		err = pidFile.Close()
		if err != nil {
			log.Fatal("Failed to close PID file:", err)
		}
	}

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, nil))
	return 0 // <-- never reached, but necessary to appease compiler
}

// NewServer builds a Server instance given a HandlerConfig
func NewServer(cfg *HandlerConfig) (*Server, error) {
	pipeline, err := NewHandlerPipeline(cfg)
	if err != nil {
		return nil, err
	}

	server := &Server{
		pipeline:   pipeline,
		debug:      cfg.Debug,
		githubPath: cfg.GithubPath,
		travisPath: cfg.TravisPath,
	}
	return server, nil
}

func (me *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := http.StatusBadRequest

	defer func() {
		w.WriteHeader(status)
		fmt.Fprintf(os.Stderr, "%s - - [%s] \"%s %s %s\" %d -\n",
			r.RemoteAddr, time.Now().UTC().Format(logTimeFmt),
			r.Method, r.URL.Path, r.Proto, status)
	}()

	if r.Method != "POST" {
		status = http.StatusMethodNotAllowed
		return
	}

	switch r.URL.Path {
	case me.githubPath:
		status = me.handleGithubPayload(w, r)
	case me.travisPath:
		status = me.handleTravisPayload(w, r)
	default:
		status = http.StatusNotFound
	}
}

func (me *Server) handleGithubPayload(w http.ResponseWriter, r *http.Request) int {
	status := http.StatusNoContent

	rawPayload := r.FormValue("payload")
	if len(rawPayload) < 1 {
		log.Println("Empty payload!")
		return http.StatusBadRequest
	}

	payload := &GithubPayload{}
	if me.debug {
		log.Println("Raw payload: ", rawPayload)
	}

	err := json.Unmarshal([]byte(rawPayload), payload)
	if err != nil {
		log.Println("Failed to unmarshal payload: ", err)
		return http.StatusBadRequest
	}

	/*
		TODO right here is "parse time" for payloads
	*/

	if !payload.IsValid() {
		if me.debug {
			log.Println("Invalid payload!")
		}
		return http.StatusBadRequest
	}

	if me.pipeline == nil {
		if me.debug {
			log.Println("No pipeline present, so doing nothing.")
		}
		return status
	}

	if me.debug {
		log.Printf("Sending payload down pipeline: %+v", payload)
	}

	err = me.pipeline.HandleGithubPayload(payload)
	if err != nil {
		status = http.StatusInternalServerError
	}

	return status
}

func (me *Server) handleTravisPayload(w http.ResponseWriter, r *http.Request) int {
	rawPayload := r.FormValue("payload")
	if len(rawPayload) < 1 {
		log.Println("Empty payload!")
		return http.StatusBadRequest
	}

	payload := &TravisPayload{}
	if me.debug {
		log.Println("Raw payload: ", rawPayload)
	}

	err := json.Unmarshal([]byte(rawPayload), payload)
	if err != nil {
		log.Println("Failed to unmarshal payload: ", err)
		return http.StatusBadRequest
	}

	if me.pipeline == nil {
		if me.debug {
			log.Println("No pipeline present, so doing nothing.")
		}
		return http.StatusNoContent
	}

	if me.debug {
		log.Printf("Sending payload down pipeline: %+v", payload)
	}

	err = me.pipeline.HandleTravisPayload(payload)
	if err != nil {
		return http.StatusInternalServerError
	}

	return http.StatusNotFound
}
