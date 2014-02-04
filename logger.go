package hookworm

import (
	"log"
)

type hookwormLogger struct {
	*log.Logger
}

func (l *hookwormLogger) Debugf(format string, v ...interface{}) {
	if !debug {
		return
	}

	l.Printf("DEBUG: "+format, v...)
}
