package rexos

var (
	// GlobalVault contains all the important secrets
	GlobalVault Vault
)

func init() {
	GlobalVault = Vault{
		ServiceClientSecret: "NotSoSecret4",
		JwtSigningKey:       "dev01",
	}
}

// Vault stores all secrets which are necessary
type Vault struct {
	ServiceClientSecret string `json:"ServiceClientSecret"`
	JwtSigningKey       string `json:"JwtSigningKey"`
}
