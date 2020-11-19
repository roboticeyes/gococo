package rexos

import "github.com/roboticeyes/gococo/math"

// Reference is a REXos reference
type Reference struct {
	// Read Only: passing this when creating a rootReference is not necessary
	RootReference       bool                         `json:"rootReference"`
	Key                 string                       `json:"key"`
	Name                string                       `json:"name"`
	Type                string                       `json:"type"`
	ProjectSelfLink     string                       `json:"project,omitempty"`
	ProjectFileSelfLink string                       `json:"projectFile,omitempty"`
	ParentReference     string                       `json:"parentReference,omitempty"`
	ProjectRoot         string                       `json:"projectRoot,omitempty"`
	LocalTransformation math.TransformationWithScale `json:"localTransformation"`
	Positioned          bool                         `json:"positioned"`
	Description         string                       `json:"description,omitempty"`
	Visible             bool                         `json:"visible"`
	Category            string                       `json:"category,omitempty"`
	DataResource        string                       `json:"dataResource,omitempty"`
}
