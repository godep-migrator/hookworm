package hookworm

import (
	"os"
	"path"
	"sort"
	"strings"
)

// HandlerConfig contains the bag of configuration poo used by all handlers
type HandlerConfig struct {
	Debug         bool         `json:"debug"`
	GithubPath    string       `json:"github_path"`
	ServerAddress string       `json:"server_address"`
	ServerPidFile string       `json:"server_pid_file"`
	StaticDir     string       `json:"static_dir"`
	TravisPath    string       `json:"travis_path"`
	WorkingDir    string       `json:"working_dir"`
	WormDir       string       `json:"worm_dir"`
	WormTimeout   int          `json:"worm_timeout"`
	WormFlags     *wormFlagMap `json:"worm_flags"`
	Version       string       `json:"version"`
}

// Handler is the interface each pipeline handler must fulfill
type Handler interface {
	HandleGithubPayload(string) (string, error)
	HandleTravisPayload(string) (string, error)
	SetNextHandler(Handler)
	NextHandler() Handler
}

// NewHandlerPipeline constructs a linked-list-like pipeline of handlers,
// each responsible for passing control to the next if deemed appropriate.
func NewHandlerPipeline(cfg *HandlerConfig) (Handler, error) {
	var (
		err      error
		pipeline Handler
	)

	pipeline = newTopHandler()

	if len(cfg.WormDir) > 0 {
		err = loadShellHandlersFromWormDir(pipeline, cfg)
		if err != nil {
			return nil, err
		}
	}

	return pipeline, nil
}

func loadShellHandlersFromWormDir(pipeline Handler, cfg *HandlerConfig) error {
	var (
		err        error
		collection []string
		directory  *os.File
	)

	if directory, err = os.Open(cfg.WormDir); err != nil {
		logger.Printf("The worm dir was not able to be opened: %v", err)
		logger.Printf("This should be the abs path to the worm dir: %v", cfg.WormDir)
		return err
	}

	if collection, err = directory.Readdirnames(-1); err != nil {
		logger.Printf("Could not read the file names from the directory: %v", err)
		return err
	}

	sort.Strings(collection)

	curHandler := pipeline

	for _, name := range collection {
		if strings.HasPrefix(name, ".") {
			logger.Printf("Ignoring hidden file %q\n", name)
			continue
		}

		fullpath := path.Join(cfg.WormDir, name)
		sh, err := newShellHandler(fullpath, cfg)

		if err != nil {
			logger.Printf("Failed to build shell handler for %v, skipping.: %v\n",
				fullpath, err)
			continue
		}

		logger.Debugf("Adding shell handler for %v\n", fullpath)

		curHandler.SetNextHandler(sh)
		curHandler = sh
	}

	logger.Debugf("Current pipeline: %#v\n", pipeline)

	for nh := pipeline.NextHandler(); nh != nil; nh = nh.NextHandler() {
		logger.Debugf("   ---> %#v\n", nh)
	}

	return nil
}
