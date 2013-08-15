package hookworm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

type wormFlagMap struct {
	values map[string]interface{}
}

func newWormFlagMap() *wormFlagMap {
	return &wormFlagMap{
		values: make(map[string]interface{}),
	}
}

func (me *wormFlagMap) String() string {
	s := ""
	for k, v := range me.values {
		s += fmt.Sprintf("%s=%v;", k, v)
	}
	return s
}

func (me *wormFlagMap) Set(value string) error {
	pairs := strings.Split(value, ";")

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			k := parts[0]
			v := parts[1]
			switch strings.ToLower(v) {
			case "true", "yes", "on":
				me.values[k] = true
			case "false", "no", "off":
				me.values[k] = false
			default:
				me.values[k] = v
			}
		} else {
			me.values[parts[0]] = true
		}
	}

	return nil
}

func (me *wormFlagMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(me.values)
}

func (me *wormFlagMap) UnmarshalJSON(raw []byte) error {
	return json.Unmarshal(raw, &me.values)
}

// HandlerConfig contains the bag of configuration poo used by all handlers
type HandlerConfig struct {
	Debug         bool         `json:"debug"`
	GithubPath    string       `json:"github_path"`
	ServerAddress string       `json:"server_address"`
	ServerPidFile string       `json:"server_pid_file"`
	TravisPath    string       `json:"travis_path"`
	WorkingDir    string       `json:"working_dir"`
	WormDir       string       `json:"worm_dir"`
	WormTimeout   int          `json:"worm_timeout"`
	WormFlags     *wormFlagMap `json:"worm_flags"`

	// TODO remove these once the python rogue handler is ready
	EmailFromAddr   string   `json:"email_from_addr"`
	EmailRcpts      []string `json:"email_recipients"`
	EmailUri        string   `json:"email_uri"`
	WatchedBranches []string `json:"watched_branches"`
	WatchedPaths    []string `json:"watched_paths"`
	// END TODO
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
