package hookworm

import (
	"log"
	"log/syslog"
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
	var err error

	//var pipeline Handler

	/*
		TODO rework pipeline construction to use ShellHandler for most biz logic
		if worm dir does not exist, then
			yell about it and exit
		else, for each file in worm dir
			if file is executable, then
				execute it with a single positional param of "configure"
				passing the JSON-serialized HandlerConfig on STDIN
					if configure call exits 0, then
						create a ShellHandler instance with the executable name
						add the ShellHandler instance to the pipeline
						log it
	*/

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
