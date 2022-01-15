package provider

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/resource"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/sdk"
)

//go:generate go run ../tools/generator-services/main.go -path=../../

func SupportedTypedServices() []sdk.TypedServiceRegistration {
	return []sdk.TypedServiceRegistration{
		resource.Registration{},
	}
}

func SupportedUntypedServices() []sdk.UntypedServiceRegistration {
	return []sdk.UntypedServiceRegistration{
		resource.Registration{},
	}
}
