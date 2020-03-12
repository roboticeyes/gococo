package event

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	// Log is a global logrus instance configured to log compliant to our journal system
	Log        *log.Logger
	timeFormat = "15:04:05.000" //time.RFC3339 //optional

	formatter = &log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "time",
			log.FieldKeyLevel: "log_level",
			log.FieldKeyMsg:   "message",
			log.FieldKeyFunc:  "logger_name",
		},
	}
)

// Fields type, used to pass to `WithFields`. Forwarded from logrus library
type Fields = log.Fields

// Entry type, used to construct logs with fields. Forwarded from logrus library
type Entry = log.Entry

//- time: datum und uhrzeit
//- message: eigentliche log meldung
//- log_level: info, warnung, error, debug etc.
//- logger_name: falls vorhanden, in java oft der name der klasse
//- username: REX username wenn JWT vorhanden
//- remote_addr: client IP (meist IP des reverse proxies oder eines anderen services)
//- http_x_forwarded_for: HTTP header X-Forwarded-For (tatsächliche client IP)
//- http_method: GET, POST etc.
//- request_path: URL path inklusive query string (z.b. /rex-gateway/api/v2/rexReferences/search/findByUrn?urn=robotic-eyes:rex-reference:1845)

func init() {
	Log = &log.Logger{
		Out:          os.Stderr,
		Formatter:    formatter,
		Hooks:        make(log.LevelHooks),
		Level:        log.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}

// ConfigureLogging configures Log to write event logs compliant to our journal system
func ConfigureLogging(debug, headless bool) {
	Log.SetLevel(log.InfoLevel)
	if debug {
		Log.SetLevel(log.DebugLevel)
	}
	if !headless {
		Log.Formatter = &log.TextFormatter{DisableColors: false, FullTimestamp: true, TimestampFormat: timeFormat}
	}
}

// Logger is the logrus logger handler
func Logger(logger log.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}
		userID, exists := c.Get("UserID")
		if !exists {
			userID = ""
		}

		entry := logger.WithFields(log.Fields{
			"http_method":          c.Request.Method,
			"request_path":         path,
			"http_x_forwarded_for": clientIP,
		})

		if userID != "" {
			entry.WithField("username", userID)
		}

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s %s %d %s (%dms)", c.Request.Method, path, statusCode, clientUserAgent, latency)
			if statusCode > 499 {
				entry.Error(msg)
			} else if statusCode > 399 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
