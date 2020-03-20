package rexos

// ProjectFile is a simple structure for carry binary meta-data
type ProjectFile struct {
	Name    string `json:"name"`
	Project string `json:"project"`
	Type    string `json:"type"`
}
