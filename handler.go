package hookworm

import (
	"log"
	"log/syslog"
	"os"
)

// HandlerConfig contains the bag of configuration poo used by all handlers
type HandlerConfig struct {
	Debug           bool     `json:"debug"`
	EmailUri        string   `json:"email_uri"`
	EmailFromAddr   string   `json:"email_from_addr"`
	EmailRcpts      []string `json:"email_recipients"`
	UseSyslog       bool     `json:"syslog"`
	WatchedBranches []string `json:"watched_branches"`
	WatchedPaths    []string `json:"watched_paths"`
	ServerPidFile   string   `json:"server_pid_file"`
	ServerAddress   string   `json:"server_address"`
	WormDir         string   `json:"worm_dir"`
	WorkingDir      string   `json:"working_dir"`
}

// Handler is the interface each pipeline handler must fulfill
type Handler interface {
	HandleGithubPayload(*GithubPayload) error
	SetNextHandler(Handler)
	NextHandler() Handler
}

// NewHandlerPipeline constructs a linked-list-like pipeline of handlers,
// each responsible for passing control to the next if deemed appropriate.
func NewHandlerPipeline(cfg *HandlerConfig) Handler {

	var (
		err        error
		pipeline   Handler
		collection []string
		directory  *os.File
	)

	pipeline = NewFakeHandler()

	if directory, err = os.Open(cfg.WormDir); err != nil {
		log.Println(err)
		log.Println("The worm dir was not able to be opened.")
		log.Println("This should be the abs path to the worm dir:" + cfg.WormDir)
	}

	if collection, err = directory.Readdirnames(-1); err != nil {
		log.Println(err)
		log.Println("Could not read the file names from the directory.")
	}

	for _, name := range collection {
		n := NewShellHandler(name, cfg)
		n.SetNextHandler(pipeline.NextHandler())
		pipeline.SetNextHandler(n)
	}

	elHandler := &EventLogHandler{debug: cfg.Debug}

	if cfg.UseSyslog {
		elHandler.sysLogger, err = syslog.NewLogger(syslog.LOG_INFO, log.LstdFlags)
		if err != nil {
			log.Panicln("Failed to initialize syslogger!", err)
		}
		if cfg.Debug {
			log.Println("Added syslog logger to event handler")
		}
	} else {
		if cfg.Debug {
			log.Println("No syslog logger added to event handler")
		}
	}
	// end part that will likely stay the same

	/*
		TODO move the rogue commit handler to a script
		that ships with the repository in the default "worm dir"
	*/
	if len(cfg.WatchedBranches) > 0 {
		elHandler.SetNextHandler(NewRogueCommitHandler(cfg))
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

	return (Handler)(elHandler)
}
