package hookworm

import (
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

type serverSetupContext struct {
	addr                string
	wormTimeoutString   string
	wormTimeout         uint64
	workingDir          string
	wormDir             string
	staticDir           string
	pidFile             string
	debugString         string
	debug               bool
	envWormFlags        string
	githubPath          string
	travisPath          string
	printRevision       bool
	printVersion        bool
	printVersionRevTags bool
	fl                  *flag.FlagSet
	args                []string
	env                 []string
	noop                bool
}

var (
	logger = &hookwormLogger{log.New(os.Stderr, "[hookworm] ", log.LstdFlags), false}
)

// ServerMain is the `main` entry point used by the `hookworm-server`
// executable
func ServerMain(c *serverSetupContext) int {
	var err error
	if c == nil {
		c = &serverSetupContext{
			addr:              os.Getenv("HOOKWORM_ADDR"),
			wormTimeoutString: os.Getenv("HOOKWORM_HANDLER_TIMEOUT"),
			wormTimeout:       uint64(30),
			workingDir:        os.Getenv("HOOKWORM_WORKING_DIR"),
			wormDir:           os.Getenv("HOOKWORM_WORM_DIR"),
			staticDir:         os.Getenv("HOOKWORM_STATIC_DIR"),
			pidFile:           os.Getenv("HOOKWORM_PID_FILE"),
			debugString:       os.Getenv("HOOKWORM_DEBUG"),
			envWormFlags:      os.Getenv("HOOKWORM_WORM_FLAGS"),
			githubPath:        os.Getenv("HOOKWORM_GITHUB_PATH"),
			travisPath:        os.Getenv("HOOKWORM_TRAVIS_PATH"),
			fl:                flag.NewFlagSet("hookworm", flag.ExitOnError),
			args:              os.Args[1:],
			env:               os.Environ(),
		}
	}

	serverSetup(c)

	c.fl.Usage = func() {
		fmt.Printf("Usage: %v [options] [key=value...]\n", progName)
		c.fl.PrintDefaults()
	}

	c.fl.Parse(c.args)
	if c.printVersion {
		printVersion()
		return 0
	}

	if c.printRevision {
		printRevision()
		return 0
	}

	if c.printVersionRevTags {
		printVersionRevTags()
		return 0
	}

	logger.debug = c.debug

	logger.Println("Starting", progVersion())

	wormFlags := newWormFlagMap()
	for i := 0; i < c.fl.NArg(); i++ {
		wormFlags.Set(c.fl.Arg(i))
	}

	envWormFlagParts := strings.Split(c.envWormFlags, ";")
	for _, flagPart := range envWormFlagParts {
		wormFlags.Set(strings.TrimSpace(flagPart))
	}

	for _, pair := range c.env {
		if !strings.HasPrefix(pair, "HOOKWORM_WORM_FLAG_") {
			continue
		}
		wormFlags.Set(strings.Replace(pair, "HOOKWORM_WORM_FLAG_", "", 1))
	}

	c.workingDir, err = getWorkingDir(c.workingDir)
	if err != nil {
		logger.Printf("ERROR: %v\n", err)
		return 1
	}

	logger.Println("Using working directory", c.workingDir)
	if err := os.Setenv("HOOKWORM_WORKING_DIR", c.workingDir); err != nil {
		logger.Printf("ERROR: %v\n", err)
		return 1
	}

	defer os.RemoveAll(c.workingDir)

	staticDir, err := getStaticDir(c.staticDir)
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
		Debug:         c.debug,
		GithubPath:    c.githubPath,
		ServerAddress: c.addr,
		ServerPidFile: c.pidFile,
		StaticDir:     c.staticDir,
		TravisPath:    c.travisPath,
		WorkingDir:    c.workingDir,
		WormDir:       c.wormDir,
		WormTimeout:   int(c.wormTimeout),
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

	if c.noop {
		return 0
	}

	logger.Fatal(http.ListenAndServe(cfg.ServerAddress, server))
	return 0 // <-- never reached, but necessary to appease compiler
}

func serverSetup(c *serverSetupContext) {
	var (
		err error
		fl  = c.fl
	)

	if len(c.wormTimeoutString) > 0 {
		c.wormTimeout, err = strconv.ParseUint(c.wormTimeoutString, 10, 64)
		if err != nil {
			logger.Fatalf("Invalid worm timeout string given: %q %v", c.wormTimeoutString, err)
		}
	}

	if len(c.debugString) > 0 {
		c.debug, err = strconv.ParseBool(c.debugString)
		if err != nil {
			logger.Fatalf("Invalid debug string given: %q %v", c.debugString, err)
		}
	}

	if c.githubPath == "" {
		c.githubPath = "/github"
	}

	if c.travisPath == "" {
		c.travisPath = "/travis"
	}

	if c.addr == "" {
		c.addr = ":9988"
	}

	fl.BoolVar(&c.printRevision, "rev", c.printRevision, "Print revision and exit")
	fl.BoolVar(&c.printVersion, "version", c.printVersion, "Print version and exit")
	fl.BoolVar(&c.printVersionRevTags, "version+", c.printVersionRevTags, "Print version, revision, and build tags")

	fl.StringVar(&c.addr, "a", c.addr, "Server address [HOOKWORM_ADDR]")
	fl.Uint64Var(&c.wormTimeout, "T", c.wormTimeout, "Timeout for handler executables (in seconds) [HOOKWORM_HANDLER_TIMEOUT]")
	fl.StringVar(&c.workingDir, "D", c.workingDir, "Working directory (scratch pad) [HOOKWORM_WORKING_DIR]")
	fl.StringVar(&c.wormDir, "W", c.wormDir, "Worm directory that contains handler executables [HOOKWORM_WORM_DIR]")
	fl.StringVar(&c.staticDir, "S", c.staticDir, "Public static directory (default $PWD/public) [HOOKWORM_STATIC_DIR]")
	fl.StringVar(&c.pidFile, "P", c.pidFile, "PID file (only written if flag given) [HOOKWORM_PID_FILE]")
	fl.BoolVar(&c.debug, "d", c.debug, "Show debug output [HOOKWORM_DEBUG]")

	fl.StringVar(&c.githubPath, "github.path", c.githubPath, "Path to handle Github payloads [HOOKWORM_GITHUB_PATH]")
	fl.StringVar(&c.travisPath, "travis.path", c.travisPath, "Path to handle Travis payloads [HOOKWORM_TRAVIS_PATH]")
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
