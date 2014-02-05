package hookworm

import (
	"log"
)

type hookwormLogger struct {
	*log.Logger
	debug bool
}

func (l *hookwormLogger) Debugf(format string, v ...interface{}) {
	if !l.debug {
		return
	}

	l.Printf("DEBUG: "+format, v...)
}
