module github.com/terraform-providers/terraform-provider-azurestack

go 1.16

require (
	github.com/Azure/azure-sdk-for-go v51.2.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.17
	github.com/davecgh/go-spew v1.1.1
	github.com/dnaeon/go-vcr v1.0.1 // indirect
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/go-azure-helpers v0.13.0
	github.com/hashicorp/terraform-plugin-sdk v1.16.0
	github.com/terraform-providers/terraform-provider-azurerm v1.44.1-0.20200911233120-57b2bfc9d42c
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
)

replace github.com/terraform-providers/terraform-provider-azurerm => github.com/openshift/terraform-provider-azurerm v1.40.1-0.20210224232508-7509319df0f4 // Pin to openshift fork with tag v2.48.0-openshift
