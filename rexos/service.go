// Package rexos is the connection layer to store the data in the rexOS.
package rexos

import (
	"github.com/roboticeyes/gococo/event"
)

var log = event.Log

// Service is the connection to rexOS
type Service struct {
	config RexConfig
	client *Client // this is the client which is used to perform the rexOS calls
}

// NewService returns a new rexos service which is implementing the RexOSAccessor interface
func NewService(cfg RexConfig) *Service {

	return &Service{
		client: NewClient(cfg),
		config: cfg,
	}
}
