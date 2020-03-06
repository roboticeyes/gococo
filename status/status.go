package status

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Status structure with code and message presentable to the user
type Status struct {
	Code           int         `json:"code" example:"400"`
	Message        string      `json:"message" example:"status bad request"`
	InternalStatus RexOSStatus `json:"-"`
}

// RexOSStatus structure as received from RexOS services
type RexOSStatus struct {
	Message   string `json:"message" example:"status bad request"`
	Timestamp string `json:"timestamp" example:"2019-12-02T14:12:53.346+0000"`
	Path      string `json:"path" example:"/resources/1/subResource"`
	Type      string `json:"type,omitempty" example:"OPTIMISTIC_LOCKING_FAILURE"`
	Code      int    `json:"status" example:"409"`
	Error     string `json:"error" example:"Conflict"`
}

// NewStatus creates a new object by the given information
func NewStatus(body []byte, code int, message string) *Status {
	status := &Status{
		Code:    code,
		Message: message,
	}
	if body != nil {
		status.InternalStatus = parseRexOSStatus(body)
	}
	return status
}

// NewHTTPStatus encapsulates a proper http error response
func NewHTTPStatus(ctx *gin.Context, status int, err error) {
	er := Status{
		Code:    status,
		Message: err.Error(),
	}
	ctx.JSON(status, er)
}

// Send sends the status back as a JSON response
func (s *Status) Send(ctx *gin.Context) {
	ctx.JSON(s.Code, s)
}

// Implements the error interface
func (s Status) Error() string {
	if s.Message != "" {
		return s.Message
	}
	return http.StatusText(s.Code)
}

func parseRexOSStatus(jsonData []byte) RexOSStatus {
	var rexOSStatus RexOSStatus
	json.Unmarshal(jsonData, &rexOSStatus)
	return rexOSStatus
}
