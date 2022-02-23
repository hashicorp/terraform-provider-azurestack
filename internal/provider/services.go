package provider

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/authorization"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/dns"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/resource"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/subscription"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/sdk"
)

//go:generate go run ../tools/generator-services/main.go -path=../../

func SupportedUntypedServices() []sdk.UntypedServiceRegistration {
	return []sdk.UntypedServiceRegistration{
		authorization.Registration{},
		compute.Registration{},
		dns.Registration{},
		loadbalancer.Registration{},
		network.Registration{},
		resource.Registration{},
		storage.Registration{},
		subscription.Registration{},
	}
}

func SupportedTypedServices() []sdk.TypedServiceRegistration {
	return []sdk.TypedServiceRegistration{
		dns.Registration{},
		resource.Registration{},
	}
}
