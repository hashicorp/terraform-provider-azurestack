package identity

import (
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

const (
	none           = "None"
	systemAssigned = "SystemAssigned"
	userAssigned   = "UserAssigned"
)

// TODO: support SystemAssigned, UserAssigned
// const systemAssignedUserAssigned = "SystemAssigned, UserAssigned"

type ExpandedConfig struct {
	// Type is the type of User Assigned Identity, either `None`, `SystemAssigned`, `UserAssigned`
	// or `SystemAssigned, UserAssigned`
	Type                    string
	PrincipalId             *string
	TenantId                *string
	UserAssignedIdentityIds *[]string
}

type Identity interface {
	Expand(input []interface{}) (*ExpandedConfig, error)
	Flatten(input *ExpandedConfig) []interface{}
	Schema() *pluginsdk.Schema
}
