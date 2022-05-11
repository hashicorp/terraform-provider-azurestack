package clients

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

type ResourceManagerAccount struct {
	AuthenticatedAsAServicePrincipal bool
	ClientId                         string
	Environment                      azure.Environment
	ObjectId                         string
	SkipResourceProviderRegistration bool
	SubscriptionId                   string
	TenantId                         string
}

func NewResourceManagerAccount(ctx context.Context, config authentication.Config, env azure.Environment, skipResourceProviderRegistration bool) (*ResourceManagerAccount, error) {
	objectId := ""

	// TODO remove this when we confirm that MSI no longer returns nil with getAuthenticatedObjectID
	// todo comment out for now as it is not stack env aware, add in a env param for it to use so it doens't look it up?
	splitEndpoint := strings.Split(env.ActiveDirectoryEndpoint, "/")
	splitEndpointlastIndex := len(splitEndpoint) - 1
	if splitEndpoint[splitEndpointlastIndex] == "adfs" || splitEndpoint[splitEndpointlastIndex] == "adfs/" {
		config.TenantID = "adfs"
	}

	if !strings.EqualFold(config.TenantID, "adfs") {
		if getAuthenticatedObjectID := config.GetAuthenticatedObjectID; getAuthenticatedObjectID != nil {
			v, err := getAuthenticatedObjectID(ctx)
			if err != nil {
				return nil, fmt.Errorf("getting authenticated object ID: %v", err)
			}
			objectId = *v
		}
	}

	account := ResourceManagerAccount{
		AuthenticatedAsAServicePrincipal: config.AuthenticatedAsAServicePrincipal,
		ClientId:                         config.ClientID,
		Environment:                      env,
		ObjectId:                         objectId,
		TenantId:                         config.TenantID,
		SkipResourceProviderRegistration: skipResourceProviderRegistration,
		SubscriptionId:                   config.SubscriptionID,
	}
	return &account, nil
}
