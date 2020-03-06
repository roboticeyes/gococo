package rexos

// Config carries all important configurations for the REXos system
type Config struct {
	// AccessTokenURL is the absolute path for requesting the token
	AccessTokenURL string

	// ClientID is the service user ID
	ClientID string

	// ClientSecret is the password for the service user
	ClientSecret string

	// JwtSigningKey is the signing key to validate the REXos token
	JwtSigningKey string

	// BasePathExtern defines the path prefix for external access
	BasePathExtern string
}
