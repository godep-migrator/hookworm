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
	addrFlag        = flag.String("a", ":9988", "Server address")
	wormTimeoutFlag = flag.Int("T", 30, "Timeout for handler executables (in seconds)")
	workingDirFlag  = flag.String("D", "", "Working directory (scratch pad)")
	wormDirFlag     = flag.String("W", "", "Worm directory that contains handler executables")
	pidFileFlag     = flag.String("P", "", "PID file (only written if flag given)")
	debugFlag       = flag.Bool("d", false, "Show debug output")

	githubPathFlag = flag.String("github.path", "/github", "Path to handle Github payloads")
	travisPathFlag = flag.String("travis.path", "/travis", "Path to handle Travis payloads")

	printRevisionFlag       = flag.Bool("rev", false, "Print revision and exit")
	printVersionFlag        = flag.Bool("version", false, "Print version and exit")
	printVersionRevTagsFlag = flag.Bool("version+", false, "Print version, revision, and build tags")

	// TODO remove these once the python rogue handler is ready
	emailFlag           = flag.String("e", "smtp://localhost:25", "Email server address")
	emailFromFlag       = flag.String("f", "hookworm@localhost", "Email from address")
	emailRcptsFlag      = flag.String("r", "", "Email recipients (comma-delimited)")
	watchedBranchesFlag = flag.String("b", "", "Watched branches (comma-delimited regexes)")
	watchedPathsFlag    = flag.String("p", "", "Watched paths (comma-delimited regexes)")
	// END TODO

	logTimeFmt = "2/Jan/2006:15:04:05 -0700" // "%d/%b/%Y:%H:%M:%S %z"
)

const (
	boomExplosionsJSON = `{"error":"BOOM EXPLOSIENS"}`
	testFormHTML       = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Hookworm test page</title>
    <style type="text/css">
      body {
        font-family: sans-serif;
      }
    </style>
  </head>
  <body>
    <article>
      <h1>Hookworm test page</h1>
      <section id="github_test">
        <h2>github test</h2>
        <form name="github" action="/github" method="post">
          <textarea name="payload" cols="80" rows="20"
		    placeholder="github payload JSON here"></textarea>
          <input type="submit" value="POST" />
        </form>
      </section>
      <section id="travis_test">
        <h2>travis test</h2>
        <form name="travis" action="/travis" method="post">
          <textarea name="payload" cols="80" rows="20"
		    placeholder="travis payload JSON here"></textarea>
          <input type="submit" value="POST" />
        </form>
      </section>
    </article>
  </body>
</html>
`
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
		fmt.Printf("Usage: %v [options] [key=value...]\n", progName)
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

	if *printVersionRevTagsFlag {
		printVersionRevTags()
		return 0
	}

	log.Println("Starting", progVersion())

	wormFlags := newWormFlagMap()
	for i := 0; i < flag.NArg(); i++ {
		wormFlags.Set(flag.Arg(i))
	}

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
		WatchedBranches: commaSplit(*watchedBranchesFlag),
		WatchedPaths:    commaSplit(*watchedPathsFlag),
		WorkingDir:      workingDir,
		WormDir:         *wormDirFlag,
		WormTimeout:     *wormTimeoutFlag,
		WormFlags:       wormFlags,
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
	body := ""
	contentType := "application/json; charset=utf-8"

	defer func() {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		fmt.Fprintf(w, "%s", body)
		fmt.Fprintf(os.Stderr, "%s - - [%s] \"%s %s %s\" %d -\n",
			r.RemoteAddr, time.Now().UTC().Format(logTimeFmt),
			r.Method, r.URL.Path, r.Proto, status)
	}()

	switch r.URL.Path {
	case me.githubPath:
		log.Printf("Handling github request at %s", r.URL.Path)
		if r.Method != "POST" {
			status = http.StatusMethodNotAllowed
			return
		}
		status, body = me.handleGithubPayload(w, r)
		return
	case me.travisPath:
		log.Printf("Handling travis request at %s", r.URL.Path)
		if r.Method != "POST" {
			status = http.StatusMethodNotAllowed
			return
		}
		status, body = me.handleTravisPayload(w, r)
		return
	case "/blank":
		log.Printf("Handling blank request at %s", r.URL.Path)
		status = http.StatusNoContent
		contentType = "text/plain; charset=utf-8"
		return
	case "/":
		log.Printf("Handling test page request at %s", r.URL.Path)
		status = http.StatusOK
		body = testFormHTML
		contentType = "text/html; charset=utf-8"
		return
	default:
		log.Printf("Handling 404 at %s", r.URL.Path)
		status = http.StatusNotFound
		contentType = "text/plain; charset=utf-8"
		body = "Nothing here.\n"
	}
}

func (me *Server) handleGithubPayload(w http.ResponseWriter, r *http.Request) (int, string) {
	payload := &GithubPayload{}
	err := me.extractPayload(payload, r)
	if err != nil {
		log.Printf("Error extracting payload: %v\n", err)
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusBadRequest, string(errJSON)
		}
		return http.StatusBadRequest, boomExplosionsJSON
	}

	if !payload.IsValid() {
		if me.debug {
			log.Println("Invalid payload!")
		}
		errJSON, err := json.Marshal(fmt.Errorf("Invalid payload!"))
		if err != nil {
			return http.StatusBadRequest, string(errJSON)
		}
		return http.StatusBadRequest, boomExplosionsJSON
	}

	if me.pipeline == nil {
		if me.debug {
			log.Println("No pipeline present, so doing nothing.")
		}
		return http.StatusNoContent, ""
	}

	if me.debug {
		log.Printf("Sending payload down pipeline: %+v", payload)
	}

	err = me.pipeline.HandleGithubPayload(payload)
	if err != nil {
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	return http.StatusNoContent, ""
}

func (me *Server) handleTravisPayload(w http.ResponseWriter, r *http.Request) (int, string) {
	payload := &TravisPayload{}
	err := me.extractPayload(payload, r)
	if err != nil {
		log.Printf("Error extracting payload: %v\n", err)

		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	if me.pipeline == nil {
		if me.debug {
			log.Println("No pipeline present, so doing nothing.")
		}
		return http.StatusNoContent, ""
	}

	if me.debug {
		log.Printf("Sending payload down pipeline: %+v", payload)
	}

	err = me.pipeline.HandleTravisPayload(payload)
	if err != nil {
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	return http.StatusNoContent, ""
}

func (me *Server) extractPayload(payload Payload, r *http.Request) error {
	rawPayload := r.FormValue("payload")
	if len(rawPayload) < 1 {
		log.Println("Empty payload!")
		return fmt.Errorf("Empty payload!")
	}

	if me.debug {
		log.Println("Raw payload: ", rawPayload)
	}

	err := json.Unmarshal([]byte(rawPayload), payload)
	if err != nil {
		log.Println("Failed to unmarshal payload: ", err)
		return err
	}

	return nil
}
