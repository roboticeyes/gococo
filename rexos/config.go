package rexos

var (
	// GlobalConfig is the global config container which contains all the REXos config data
	GlobalConfig RexConfig
)

// RexConfig defines all the required settings for the REXos
type RexConfig struct {
	Resources      Resources     `json:"ResourceUrls"`
	ServiceClient  ServiceClient `json:"ServiceClient"`
	BasePathExtern string        `json:"BasePathExtern"`
}

// Resources groups all resource URLs which are required by this composite service
type Resources struct {
	Project      string `json:"Project"`
	ProjectFile  string `json:"ProjectFile"`
	RexReference string `json:"RexReference"`
	Inspection   string `json:"Inspection"`
	Activity     string `json:"Activity"`
	User         string `json:"User"`
}

// ServiceClient specifies the credentials for the service user
type ServiceClient struct {
	AccessTokenURL string `json:"AccessTokenURL"`
	ID             string `json:"ClientID"`
}
