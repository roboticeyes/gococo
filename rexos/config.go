package rexos

// Config defines all the required settings for the REXos
type Config struct {
	API            ResourcesConfig     `json:"ResourceUrls"`
	ServiceClient  ServiceClientConfig `json:"ServiceClient"`
	BasePathExtern string              `json:"BasePathExtern"`
}

// ResourcesConfig groups all resource URLs which are required by this composite service
type ResourcesConfig struct {
	Project      string `json:"Project"`
	ProjectFile  string `json:"ProjectFile"`
	RexReference string `json:"RexReference"`
	Inspection   string `json:"Inspection"`
	Activity     string `json:"Activity"`
	User         string `json:"User"`
}

// ServiceClientConfig specifies the credentials for the service user
type ServiceClientConfig struct {
	AccessTokenURL string `json:"AccessTokenURL"`
	ID             string `json:"ClientID"`
}
