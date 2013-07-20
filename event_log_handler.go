package hookworm

import (
	"encoding/json"
	"log"
)

type EventLogHandler struct {
	sysLogger   *log.Logger
	nextHandler Handler
	debug       bool
}

func (me *EventLogHandler) HandleGithubPayload(payload *GithubPayload) error {
	asJsonBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Println("Failed to re-serialize payload:", err)
		return err
	}
	asJson := string(asJsonBytes)

	log.Printf("Pull request merge? %v\n", payload.IsPullRequestMerge())
	if me.debug {
		log.Printf("payload=%+v\n", payload)
		log.Printf("payload json=%v\n", asJson)
	}

	if me.sysLogger != nil {
		me.sysLogger.Println(asJson)
	}

	if nxt := me.NextHandler(); nxt != nil {
		return nxt.HandleGithubPayload(payload)
	}

	return nil
}

func (me *EventLogHandler) SetNextHandler(handler Handler) {
	me.nextHandler = handler
}

func (me *EventLogHandler) NextHandler() Handler {
	return me.nextHandler
}
