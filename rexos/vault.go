package rexos

// Vault stores all secrets which are necessary
type Vault struct {
	ServiceClientSecret string `json:"ServiceClientSecret"`
	JwtSigningKey       string `json:"JwtSigningKey"`
}
