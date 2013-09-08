package hookworm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

var (
	addrFlag        = flag.String("a", ":9988", "Server address")
	wormTimeoutFlag = flag.Int("T", 30, "Timeout for handler executables (in seconds)")
	workingDirFlag  = flag.String("D", "", "Working directory (scratch pad)")
	wormDirFlag     = flag.String("W", "", "Worm directory that contains handler executables")
	staticDirFlag   = flag.String("S", "", "Public static directory (default $PWD/public)")
	pidFileFlag     = flag.String("P", "", "PID file (only written if flag given)")
	debugFlag       = flag.Bool("d", false, "Show debug output")

	githubPathFlag = flag.String("github.path", "/github", "Path to handle Github payloads")
	travisPathFlag = flag.String("travis.path", "/travis", "Path to handle Travis payloads")

	printRevisionFlag       = flag.Bool("rev", false, "Print revision and exit")
	printVersionFlag        = flag.Bool("version", false, "Print version and exit")
	printVersionRevTagsFlag = flag.Bool("version+", false, "Print version, revision, and build tags")

	logTimeFmt   = "2/Jan/2006:15:04:05 -0700" // "%d/%b/%Y:%H:%M:%S %z"
	testFormHTML = template.Must(template.New("test_form").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Hookworm test page</title>
    <link rel="shortcut icon" href="../favicon.ico">
    <style type="text/css">
      body { font-family: sans-serif; }
    </style>
  </head>
  <body>
    <article>
      <h1>Hookworm test page</h1>
      <pre>{{.ProgVersion}}</pre>
      <hr>
      {{if .Debug}}
      <section id="debug_links">
        <h2>debugging </h2>
        <ul>
          <li><a href="vars">vars</a></li>
          <li><a href="pprof/">pprof</a></li>
        </ul>
      </section>
      {{end}}
      <section id="github_test">
        <h2>github test</h2>
        <form name="github" action="{{.GithubPath}}" method="post">
          <textarea name="payload" cols="80" rows="20"
                    placeholder="github payload JSON here"></textarea>
          <input type="submit" value="POST" />
        </form>
      </section>
      <section id="travis_test">
        <h2>travis test</h2>
        <form name="travis" action="{{.TravisPath}}" method="post">
          <textarea name="payload" cols="80" rows="20"
                    placeholder="travis payload JSON here"></textarea>
          <input type="submit" value="POST" />
        </form>
      </section>
    </article>
  </body>
</html>
`))
	hookwormIndex = `
   oo           ___        ___       ___  __   ___
   |"     |__| |__  \ /     |  |__| |__  |__) |__
   |      |  | |___  |      |  |  | |___ |  \ |___
 --'
--------------------------------------------------
`
	hookwormFaviconBytes []byte
)

const (
	boomExplosionsJSON = `{"error":"BOOM EXPLOSIONS"}`
	ctypeText          = "text/plain; charset=utf-8"
	ctypeJSON          = "application/json; charset=utf-8"
	ctypeHTML          = "text/html; charset=utf-8"
	ctypeIcon          = "image/vnd.microsoft.icon"

	hookwormFaviconBase64 = `
AAABAAEAEBAQAAAAAAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAAAAAAA
AAB7/wAAgushAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAABEREQAAEQABEREREAAREREREREQABERERAAERAAARERAAAR
EAAAERAAABEQAAAAAAAAERAAAAAAAAAREAAAAAAAABERAAAAAAAAEREQAAAAAAARERAAAAAAAAERAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAA`
)

type testFormContext struct {
	GithubPath  string
	TravisPath  string
	ProgVersion string
	Debug       bool
}

func init() {
	hookwormFaviconBytes, _ = base64.StdEncoding.DecodeString(hookwormFaviconBase64)
}

// Server implements ServeHTTP, parsing payloads and handing them off to the
// handler pipeline
type Server struct {
	pipeline   Handler
	cfg        *HandlerConfig
	debug      bool
	githubPath string
	travisPath string
	fileServer http.Handler
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
	if err := os.Setenv("HOOKWORM_WORKING_DIR", workingDir); err != nil {
		log.Printf("ERROR: %v\n", err)
		return 1
	}

	defer os.RemoveAll(workingDir)

	staticDir, err := getStaticDir(*staticDirFlag)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return 1
	}

	log.Println("Using static directory", staticDir)
	if err := os.Setenv("HOOKWORM_STATIC_DIR", staticDir); err != nil {
		log.Printf("ERROR: %v\n", err)
		return 1
	}

	cfg := &HandlerConfig{
		Debug:         *debugFlag,
		GithubPath:    *githubPathFlag,
		ServerAddress: *addrFlag,
		ServerPidFile: *pidFileFlag,
		StaticDir:     staticDir,
		TravisPath:    *travisPathFlag,
		WorkingDir:    workingDir,
		WormDir:       *wormDirFlag,
		WormTimeout:   *wormTimeoutFlag,
		WormFlags:     wormFlags,
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
		cfg:        cfg,
		debug:      cfg.Debug,
		githubPath: cfg.GithubPath,
		travisPath: cfg.TravisPath,
		fileServer: http.FileServer(http.Dir(cfg.StaticDir)),
	}
	return server, nil
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := http.StatusBadRequest
	body := ""
	contentType := ctypeJSON

	defer func() {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		fmt.Fprintf(w, "%s", body)
		fmt.Fprintf(os.Stderr, "%s - - [%s] \"%s %s %s\" %d -\n",
			r.RemoteAddr, time.Now().UTC().Format(logTimeFmt),
			r.Method, r.URL.Path, r.Proto, status)
	}()

	switch r.URL.Path {
	case srv.githubPath:
		log.Printf("Handling github request at %s", r.URL.Path)
		if r.Method != "POST" {
			status = http.StatusMethodNotAllowed
			return
		}
		status, body = srv.handleGithubPayload(w, r)
		return
	case srv.travisPath:
		log.Printf("Handling travis request at %s", r.URL.Path)
		if r.Method != "POST" {
			status = http.StatusMethodNotAllowed
			return
		}
		status, body = srv.handleTravisPayload(w, r)
		return
	case "/blank":
		log.Printf("Handling blank request at %s", r.URL.Path)
		status = http.StatusNoContent
		contentType = ctypeText
		return
	case "/config":
		log.Printf("Handling config request at %s", r.URL.Path)
		status, body = srv.handleConfig(w, r)
		return
	case "/favicon.ico":
		log.Printf("Handling favicon request at %s", r.URL.Path)
		status = http.StatusOK
		contentType = ctypeIcon
		body = string(hookwormFaviconBytes)
		return
	case "/", "/index", "/index.txt":
		log.Printf("Handling index page request at %s", r.URL.Path)
		status = http.StatusOK
		contentType = ctypeText
		body = fmt.Sprintf("%s\n%s\n", hookwormIndex, progVersion())
		return
	case "/debug/test":
		if srv.debug {
			log.Printf("Handling test page request at %s", r.URL.Path)
			status, body, contentType = srv.handleTestPage(w, r)
		} else {
			log.Printf("Debug not enabled, so returning 404 for %s", r.URL.Path)
			status = http.StatusNotFound
			contentType = ctypeText
			body = "Nothing here.\n"
		}
		return
	default:
		log.Printf("Forwarding request to file server: %s", r.URL.Path)
		srv.fileServer.ServeHTTP(w, r)
		return
	}
}

func (srv *Server) handleTestPage(w http.ResponseWriter, r *http.Request) (int, string, string) {
	status := http.StatusOK
	body := ""
	contentType := ctypeText

	var bodyBuf bytes.Buffer

	err := testFormHTML.Execute(&bodyBuf, &testFormContext{
		GithubPath:  strings.TrimLeft(srv.githubPath, "/"),
		TravisPath:  strings.TrimLeft(srv.travisPath, "/"),
		ProgVersion: progVersion(),
		Debug:       srv.debug,
	})
	if err != nil {
		status = http.StatusInternalServerError
		body = fmt.Sprintf("%+v", err)
		contentType = ctypeText
	} else {
		body = string(bodyBuf.Bytes())
		contentType = ctypeHTML
	}

	return status, body, contentType
}

func (srv *Server) handleConfig(w http.ResponseWriter, r *http.Request) (int, string) {
	bodyJSON, err := json.MarshalIndent(srv.cfg, "", "  ")
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("%v", err)
	}

	return http.StatusOK, string(bodyJSON) + "\n"
}

func (srv *Server) handleGithubPayload(w http.ResponseWriter, r *http.Request) (int, string) {
	payload, err := srv.extractPayload(r)
	if err != nil {
		log.Printf("Error extracting payload: %v\n", err)
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusBadRequest, string(errJSON)
		}
		return http.StatusBadRequest, boomExplosionsJSON
	}

	if srv.pipeline == nil {
		if srv.debug {
			log.Println("No pipeline present, so doing nothing.")
		}
		return http.StatusNoContent, ""
	}

	if srv.debug {
		log.Printf("Sending payload down pipeline: %+v", payload)
	}

	_, err = srv.pipeline.HandleGithubPayload(payload)
	if err != nil {
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	return http.StatusNoContent, ""
}

func (srv *Server) handleTravisPayload(w http.ResponseWriter, r *http.Request) (int, string) {
	payload, err := srv.extractPayload(r)
	if err != nil {
		log.Printf("Error extracting payload: %v\n", err)

		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	if srv.pipeline == nil {
		if srv.debug {
			log.Println("No pipeline present, so doing nothing.")
		}
		return http.StatusNoContent, ""
	}

	if srv.debug {
		log.Printf("Sending payload down pipeline: %+v", payload)
	}

	_, err = srv.pipeline.HandleTravisPayload(payload)
	if err != nil {
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	return http.StatusNoContent, ""
}

func (srv *Server) extractPayload(r *http.Request) (string, error) {
	rawPayload := ""
	ctype := abbrCtype(r.Header.Get("Content-Type"))

	switch ctype {
	case "application/json", "text/javascript", "text/plain":
		rawPayloadBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", err
		}
		rawPayload = string(rawPayloadBytes)
	case "application/x-www-form-urlencoded":
		rawPayload = r.FormValue("payload")
	}

	if len(rawPayload) < 1 {
		log.Println("Empty payload!")
		return "", fmt.Errorf("empty payload")
	}

	if srv.debug {
		log.Println("Raw payload: ", rawPayload)
	}
	return rawPayload, nil
}

func abbrCtype(ctype string) string {
	s := strings.Split(ctype, ";")[0]
	return strings.ToLower(strings.TrimSpace(s))
}
