package authentication

import (
	"github.com/Azure/go-autorest/autorest/adal"
)

// Config is the configuration structure used to instantiate a
// new Azure Stack management client.
type Config struct {
	ClientID       string
	SubscriptionID string
	TenantID       string
	ARMEndpoint    string
	Environment    string

	// Service Principal Auth
	ClientSecret string

	// Bearer Auth
	AccessToken  *adal.Token
	IsCloudShell bool
	UseMsi       bool
	MsiEndpoint  string
}
