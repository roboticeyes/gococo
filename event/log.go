package event

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// Log is a global logrus instance configured to log compliant to our journal system
var (
	Log *log.Logger
)

// Fields type, used to pass to `WithFields`. Forwarded from logrus library
type Fields = log.Fields

func init() {
	Log = &log.Logger{
		Out:          os.Stderr,
		Formatter:    &log.TextFormatter{DisableColors: false, FullTimestamp: true},
		Hooks:        make(log.LevelHooks),
		Level:        log.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}

// ConfigureLogging configures Log to write event logs compliant to our journal system
func ConfigureLogging(debug bool) {
	Log.SetLevel(log.DebugLevel)
	if !debug {
		Log.SetLevel(log.InfoLevel)
	}
}
