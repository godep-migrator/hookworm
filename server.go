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
	"strconv"
	"strings"
	"text/template"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
)

var (
	addr              = os.Getenv("HOOKWORM_ADDR")
	wormTimeoutString = os.Getenv("HOOKWORM_HANDLER_TIMEOUT")
	wormTimeout       = uint64(30)
	workingDir        = os.Getenv("HOOKWORM_WORKING_DIR")
	wormDir           = os.Getenv("HOOKWORM_WORM_DIR")
	staticDir         = os.Getenv("HOOKWORM_STATIC_DIR")
	pidFile           = os.Getenv("HOOKWORM_PID_FILE")
	debugString       = os.Getenv("HOOKWORM_DEBUG")
	debug             = false

	envWormFlags = os.Getenv("HOOKWORM_WORM_FLAGS")

	githubPath = os.Getenv("HOOKWORM_GITHUB_PATH")
	travisPath = os.Getenv("HOOKWORM_TRAVIS_PATH")
)

var (
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
	var err error

	hookwormFaviconBytes, _ = base64.StdEncoding.DecodeString(hookwormFaviconBase64)

	if len(wormTimeoutString) > 0 {
		wormTimeout, err = strconv.ParseUint(wormTimeoutString, 10, 64)
		if err != nil {
			log.Fatalf("Invalid worm timeout string given: %q %v", wormTimeoutString, err)
		}
	}

	if len(debugString) > 0 {
		debug, err = strconv.ParseBool(debugString)
		if err != nil {
			log.Fatalf("Invalid debug string given: %q %v", debugString, err)
		}
	}

	if githubPath == "" {
		githubPath = "/github"
	}

	if travisPath == "" {
		travisPath = "/travis"
	}

	if addr == "" {
		addr = ":9988"
	}

	flag.StringVar(&addr, "a", addr, "Server address [HOOKWORM_ADDR]")
	flag.Uint64Var(&wormTimeout, "T", wormTimeout, "Timeout for handler executables (in seconds) [HOOKWORM_HANDLER_TIMEOUT]")
	flag.StringVar(&workingDir, "D", workingDir, "Working directory (scratch pad) [HOOKWORM_WORKING_DIR]")
	flag.StringVar(&wormDir, "W", wormDir, "Worm directory that contains handler executables [HOOKWORM_WORM_DIR]")
	flag.StringVar(&staticDir, "S", staticDir, "Public static directory (default $PWD/public) [HOOKWORM_STATIC_DIR]")
	flag.StringVar(&pidFile, "P", pidFile, "PID file (only written if flag given) [HOOKWORM_PID_FILE]")
	flag.BoolVar(&debug, "d", debug, "Show debug output [HOOKWORM_DEBUG]")

	flag.StringVar(&githubPath, "github.path", githubPath, "Path to handle Github payloads [HOOKWORM_GITHUB_PATH]")
	flag.StringVar(&travisPath, "travis.path", travisPath, "Path to handle Travis payloads [HOOKWORM_TRAVIS_PATH]")
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

	envWormFlagParts := strings.Split(envWormFlags, ";")
	for _, flagPart := range envWormFlagParts {
		wormFlags.Set(strings.TrimSpace(flagPart))
	}

	workingDir, err := getWorkingDir(workingDir)
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

	staticDir, err := getStaticDir(staticDir)
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
		Debug:         debug,
		GithubPath:    githubPath,
		ServerAddress: addr,
		ServerPidFile: pidFile,
		StaticDir:     staticDir,
		TravisPath:    travisPath,
		WorkingDir:    workingDir,
		WormDir:       wormDir,
		WormTimeout:   int(wormTimeout),
		WormFlags:     wormFlags,
		Version:       progVersion(),
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

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, server))
	return 0 // <-- never reached, but necessary to appease compiler
}

// NewServer builds a martini.ClassicMartini instance given a HandlerConfig
func NewServer(cfg *HandlerConfig) (*martini.ClassicMartini, error) {
	pipeline, err := NewHandlerPipeline(cfg)
	if err != nil {
		return nil, err
	}

	m := martini.Classic()

	m.Use(martini.Static(cfg.StaticDir))
	m.Use(render.Renderer())

	m.MapTo(pipeline, (*Handler)(nil))
	m.Map(cfg)

	m.Post(cfg.GithubPath, handleGithubPayload)
	m.Post(cfg.TravisPath, handleTravisPayload)
	m.Get("/blank", func() int {
		return http.StatusNoContent
	})
	m.Get("/config", handleConfig)
	m.Get("/favicon.ico", func() (int, string) {
		return http.StatusOK, string(hookwormFaviconBytes)
	})
	m.Get("/", handleIndex)
	m.Get("/index", handleIndex)
	m.Get("/index.txt", handleIndex)
	if cfg.Debug {
		m.Get("/debug/test", handleTestPage)
	}

	return m, nil
}

func handleIndex() string {
	return fmt.Sprintf("%s\n%s\n", hookwormIndex, progVersion())
}

func handleTestPage(cfg *HandlerConfig, w http.ResponseWriter) (int, string) {
	status := http.StatusOK
	body := ""

	var bodyBuf bytes.Buffer

	err := testFormHTML.Execute(&bodyBuf, &testFormContext{
		GithubPath:  strings.TrimLeft(cfg.GithubPath, "/"),
		TravisPath:  strings.TrimLeft(cfg.TravisPath, "/"),
		ProgVersion: progVersion(),
		Debug:       cfg.Debug,
	})
	if err != nil {
		status = http.StatusInternalServerError
		body = fmt.Sprintf("<!DOCTYPE html><html><head></head><body><h1>%+v</h1></body></html>", err)
	} else {
		body = string(bodyBuf.Bytes())
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	return status, body
}

func handleConfig(cfg *HandlerConfig, r render.Render) {
	r.JSON(http.StatusOK, cfg)
}

func handleGithubPayload(pipeline Handler, cfg *HandlerConfig, r *http.Request) (int, string) {
	payload, err := extractPayload(cfg, r)
	if err != nil {
		log.Printf("Error extracting payload: %v\n", err)
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusBadRequest, string(errJSON)
		}
		return http.StatusBadRequest, boomExplosionsJSON
	}

	if pipeline == nil {
		if cfg.Debug {
			log.Println("No pipeline present, so doing nothing.")
		}
		return http.StatusNoContent, ""
	}

	if cfg.Debug {
		log.Printf("Sending payload down pipeline: %+v", payload)
	}

	_, err = pipeline.HandleGithubPayload(payload)
	if err != nil {
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	return http.StatusNoContent, ""
}

func handleTravisPayload(pipeline Handler, cfg *HandlerConfig, r *http.Request) (int, string) {
	payload, err := extractPayload(cfg, r)
	if err != nil {
		log.Printf("Error extracting payload: %v\n", err)

		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	if pipeline == nil {
		if cfg.Debug {
			log.Println("No pipeline present, so doing nothing.")
		}
		return http.StatusNoContent, ""
	}

	if cfg.Debug {
		log.Printf("Sending payload down pipeline: %+v", payload)
	}

	_, err = pipeline.HandleTravisPayload(payload)
	if err != nil {
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	return http.StatusNoContent, ""
}

func extractPayload(cfg *HandlerConfig, r *http.Request) (string, error) {
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

	if cfg.Debug {
		log.Println("Raw payload: ", rawPayload)
	}
	return rawPayload, nil
}

func abbrCtype(ctype string) string {
	s := strings.Split(ctype, ";")[0]
	return strings.ToLower(strings.TrimSpace(s))
}
