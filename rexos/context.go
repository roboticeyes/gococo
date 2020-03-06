package rexos

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// ContextDataKey is the identifier for getting the context data
	ContextDataKey = "data"
)

// XForwarded header information
type XForwarded struct {
	Host   string
	Port   string
	Prefix string
	Proto  string
}

// ContextData is used as payload for the REX context in the interface
// The caller of the functions which takes a context must make sure that
// both data values are filled. Please use `context.WithValue` to add
// this information
type ContextData struct {
	AccessToken string
	UserID      string
	XForwarded  XForwarded
}

// GetRexContext parses the GIN context and extracts the necessary token, while
// adding the token to a new context for the rexOS calls.
func GetRexContext(c *gin.Context) (context.Context, context.CancelFunc) {
	var contextData ContextData
	contextData.AccessToken = c.GetHeader(AuthorizationKey)
	userID, exists := c.Get(KeyUserID)
	if !exists {
		userID = ""
	}
	contextData.UserID = userID.(string)
	contextData.XForwarded.Host = c.Request.Header.Get("X-Forwarded-Host")
	contextData.XForwarded.Port = c.Request.Header.Get("X-Forwarded-Port")
	contextData.XForwarded.Prefix = c.Request.Header.Get("X-Forwarded-Prefix")
	contextData.XForwarded.Proto = c.Request.Header.Get("X-Forwarded-Proto")

	ctx := context.WithValue(context.Background(), ContextDataKey, contextData)
	return context.WithTimeout(ctx, time.Second*2)
}

// GetUserIDFromContext retruns the user id from the context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	contextData := ctx.Value(ContextDataKey)
	if contextData == nil {
		return "", fmt.Errorf("Context does not contain any data")
	}
	return contextData.(ContextData).UserID, nil
}

// GetAccessTokenFromContext returns the accesstoken from the context
func GetAccessTokenFromContext(ctx context.Context) (string, error) {
	contextData := ctx.Value(ContextDataKey)
	if contextData == nil {
		return "", fmt.Errorf("Context does not contain any data")
	}
	return contextData.(ContextData).AccessToken, nil
}

// GetXForwarded returns X-Forwarded header
func GetXForwarded(ctx context.Context) (XForwarded, error) {
	contextData := ctx.Value(ContextDataKey)
	if contextData == nil {
		return XForwarded{}, fmt.Errorf("Context does not contain any data")
	}
	return contextData.(ContextData).XForwarded, nil
}
