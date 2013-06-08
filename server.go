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

	emailFlag           = flag.String("e", "smtp://localhost", "Email server address")
	emailFromFlag       = flag.String("f", "hookworm@localhost", "Email from address")
	emailRcptsFlag      = flag.String("r", "", "Email recipients (comma-delimited)")
	watchedBranchesFlag = flag.String("b", "", "Watched branches (comma-delimited regexes)")
	watchedPathsFlag    = flag.String("p", "", "Watched paths (comma-delimited regexes)")

	useSyslogFlag = flag.Bool("S", false, "Send all received events to syslog")

	printVersionFlag = flag.Bool("v", false, "Print version and exit")
	debugFlag        = flag.Bool("d", false, "Show debug output")

	logTimeFmt = "2/Jan/2006:15:04:05 -0700" // "%d/%b/%Y:%H:%M:%S %z"
)

type Server struct {
	pipeline Handler
	debug    bool
}

func ServerMain() {
	flag.Usage = func() {
		fmt.Printf("Usage: %v [options]\n", progName)
		flag.PrintDefaults()
	}

	flag.Parse()
	if *printVersionFlag {
		printVersion()
		return
	} else {
		log.Println("Starting", progVersion())
	}

	cfg := &HandlerConfig{
		Debug:           *debugFlag,
		EmailUri:        *emailFlag,
		EmailFromAddr:   *emailFromFlag,
		EmailRcpts:      commaSplit(*emailRcptsFlag),
		UseSyslog:       *useSyslogFlag,
		WatchedBranches: commaSplit(*watchedBranchesFlag),
		WatchedPaths:    commaSplit(*watchedPathsFlag),
	}

	if *debugFlag {
		log.Printf("Using handler config: %+v\n", cfg)
	}
	server := NewServer(cfg)
	if server == nil {
		log.Fatal("No server?  No worky!")
	}

	http.Handle("/", server)
	log.Printf("Listening on %v\n", *addrFlag)
	log.Fatal(http.ListenAndServe(*addrFlag, nil))
}

func NewServer(cfg *HandlerConfig) *Server {
	return &Server{
		pipeline: NewHandlerPipeline(cfg),
		debug:    cfg.Debug,
	}
}

func (me *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := http.StatusBadRequest

	defer func() {
		w.WriteHeader(status)
		fmt.Fprintf(os.Stderr, "%s - - [%s] \"%s %s %s\" %d -\n",
			r.RemoteAddr, time.Now().UTC().Format(logTimeFmt),
			r.Method, r.URL.Path, r.Proto, status)
	}()

	rawPayload := r.FormValue("payload")
	if len(rawPayload) < 1 {
		log.Println("Empty payload!")
		return
	}

	payload := &Payload{}
	err := json.Unmarshal([]byte(rawPayload), payload)
	if err != nil {
		log.Println("Failed to unmarshal payload: ", err)
		if me.debug {
			log.Println("Raw payload: ", rawPayload)
		}
		return
	}

	status = http.StatusNoContent

	if me.pipeline == nil {
		return
	}

	err = me.pipeline.HandlePayload(payload)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
}
