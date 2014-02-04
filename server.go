package hookworm

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	printRevisionFlag       = flag.Bool("rev", false, "Print revision and exit")
	printVersionFlag        = flag.Bool("version", false, "Print version and exit")
	printVersionRevTagsFlag = flag.Bool("version+", false, "Print version, revision, and build tags")

	logger = &hookwormLogger{log.New(os.Stderr, "[hookworm] ", log.LstdFlags)}
)

func init() {
	var err error

	hookwormFaviconBytes, _ = base64.StdEncoding.DecodeString(hookwormFaviconBase64)

	if len(wormTimeoutString) > 0 {
		wormTimeout, err = strconv.ParseUint(wormTimeoutString, 10, 64)
		if err != nil {
			logger.Fatalf("Invalid worm timeout string given: %q %v", wormTimeoutString, err)
		}
	}

	if len(debugString) > 0 {
		debug, err = strconv.ParseBool(debugString)
		if err != nil {
			logger.Fatalf("Invalid debug string given: %q %v", debugString, err)
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

	logger.Println("Starting", progVersion())

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
		logger.Printf("ERROR: %v\n", err)
		return 1
	}

	logger.Println("Using working directory", workingDir)
	if err := os.Setenv("HOOKWORM_WORKING_DIR", workingDir); err != nil {
		logger.Printf("ERROR: %v\n", err)
		return 1
	}

	defer os.RemoveAll(workingDir)

	staticDir, err := getStaticDir(staticDir)
	if err != nil {
		logger.Printf("ERROR: %v\n", err)
		return 1
	}

	logger.Println("Using static directory", staticDir)
	if err := os.Setenv("HOOKWORM_STATIC_DIR", staticDir); err != nil {
		logger.Printf("ERROR: %v\n", err)
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

	logger.Debugf("Using handler config: %+v\n", cfg)

	if err := os.Chdir(cfg.WorkingDir); err != nil {
		logger.Fatalf("Failed to move into working directory %v\n", cfg.WorkingDir)
	}

	server, err := NewServer(cfg)

	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("Listening on %v\n", cfg.ServerAddress)

	if len(cfg.ServerPidFile) > 0 {
		pidFile, err := os.Create(cfg.ServerPidFile)
		if err != nil {
			logger.Fatal("Failed to open PID file:", err)
		}
		fmt.Fprintf(pidFile, "%d\n", os.Getpid())
		err = pidFile.Close()
		if err != nil {
			logger.Fatal("Failed to close PID file:", err)
		}
	}

	logger.Fatal(http.ListenAndServe(cfg.ServerAddress, server))
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
	m.Map(logger)

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
