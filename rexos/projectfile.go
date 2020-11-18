package rexos

import "github.com/roboticeyes/gococo/math"

// ProjectFile is a simple structure for carry binary meta-data
type ProjectFile struct {
	Name               string                       `json:"name"`
	Project            string                       `json:"project"`
	DataTransformation math.TransformationWithScale `json:"dataTransformation"`
	Type               string                       `json:"type"`
}
