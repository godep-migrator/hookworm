package hookworm

import (
	"log"
	"log/syslog"
)

type HandlerConfig struct {
	Debug           bool
	EmailUri        string
	EmailFromAddr   string
	EmailRcpts      []string
	UseSyslog       bool
	PolicedBranches []string
	ServerPidFile   string
	ServerAddress   string
}

type Handler interface {
	HandlePayload(*Payload) error
	SetNextHandler(Handler)
	NextHandler() Handler
}

func NewHandlerPipeline(cfg *HandlerConfig) Handler {
	var err error
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

	if len(cfg.PolicedBranches) > 0 {
		elHandler.SetNextHandler(NewRogueCommitHandler(cfg))
		if cfg.Debug {
			log.Printf("Added rogue commit handler for policed branches %+v\n",
				cfg.PolicedBranches)
		}
	} else {
		if cfg.Debug {
			log.Println("No rogue commit handler added")
		}
	}

	return (Handler)(elHandler)
}
