package hookworm

import (
	"log"
	"os"
	"path"
)

// HandlerConfig contains the bag of configuration poo used by all handlers
type HandlerConfig struct {
	Debug           bool     `json:"debug"`
	EmailFromAddr   string   `json:"email_from_addr"`
	EmailRcpts      []string `json:"email_recipients"`
	EmailUri        string   `json:"email_uri"`
	GithubPath      string   `json:"github_path"`
	ServerAddress   string   `json:"server_address"`
	ServerPidFile   string   `json:"server_pid_file"`
	TravisPath      string   `json:"travis_path"`
	UseSyslog       bool     `json:"syslog"`
	WatchedBranches []string `json:"watched_branches"`
	WatchedPaths    []string `json:"watched_paths"`
	WorkingDir      string   `json:"working_dir"`
	WormDir         string   `json:"worm_dir"`
	WormTimeout     int      `json:"worm_timeout"`
}

// Handler is the interface each pipeline handler must fulfill
type Handler interface {
	HandleGithubPayload(*GithubPayload) error
	HandleTravisPayload(*TravisPayload) error
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

	/*
		TODO move the rogue commit handler to a script
		that ships with the repository in the default "worm dir"
	*/
	if len(cfg.WatchedBranches) > 0 {
		pipeline.SetNextHandler(NewRogueCommitHandler(cfg))
		if cfg.Debug {
			log.Printf("Added rogue commit handler "+
				"for watched branches %+v, watched paths %+v\n",
				cfg.WatchedBranches, cfg.WatchedPaths)
		}
	} else {
		if cfg.Debug {
			log.Println("No rogue commit handler added")
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
		log.Printf("The worm dir was not able to be opened: %v", err)
		log.Printf("This should be the abs path to the worm dir: %v", cfg.WormDir)
		return err
	}

	if collection, err = directory.Readdirnames(-1); err != nil {
		log.Printf("Could not read the file names from the directory: %v", err)
		return err
	}

	for _, name := range collection {
		fullpath := path.Join(cfg.WormDir, name)
		sh, err := newShellHandler(fullpath, cfg)

		if err != nil {
			log.Printf("Failed to build shell handler for %v, skipping.: %v",
				fullpath, err)
			continue
		}

		if cfg.Debug {
			log.Printf("Adding shell handler for %v\n", fullpath)
		}

		sh.SetNextHandler(pipeline.NextHandler())
		pipeline.SetNextHandler(sh)
	}

	return nil
}
