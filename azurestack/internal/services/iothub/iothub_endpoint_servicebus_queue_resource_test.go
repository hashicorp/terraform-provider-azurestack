package iothub_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type IotHubEndpointServiceBusQueueResource struct {
}

func TestAccIotHubEndpointServiceBusQueue_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_endpoint_servicebus_queue", "test")
	r := IotHubEndpointServiceBusQueueResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccIotHubEndpointServiceBusQueue_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_endpoint_servicebus_queue", "test")
	r := IotHubEndpointServiceBusQueueResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_iothub_endpoint_servicebus_queue"),
		},
	})
}

func (IotHubEndpointServiceBusQueueResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-iothub-%[1]d"
  location = "%[2]s"
}

resource "azurerm_servicebus_namespace" "test" {
  name                = "acctest-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Standard"
}

resource "azurerm_servicebus_queue" "test" {
  name                = "acctest-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  namespace_name      = azurerm_servicebus_namespace.test.name

  enable_partitioning = true
}

resource "azurerm_servicebus_queue_authorization_rule" "test" {
  name                = "acctest-%[1]d"
  namespace_name      = azurerm_servicebus_namespace.test.name
  queue_name          = azurerm_servicebus_queue.test.name
  resource_group_name = azurerm_resource_group.test.name

  listen = false
  send   = true
  manage = false
}

resource "azurerm_iothub" "test" {
  name                = "acctestIoTHub-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  sku {
    name     = "B1"
    capacity = "1"
  }

  tags = {
    purpose = "testing"
  }
}

resource "azurerm_iothub_endpoint_servicebus_queue" "test" {
  resource_group_name = azurerm_resource_group.test.name
  iothub_name         = azurerm_iothub.test.name
  name                = "acctest"

  connection_string = azurerm_servicebus_queue_authorization_rule.test.primary_connection_string
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r IotHubEndpointServiceBusQueueResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_iothub_endpoint_servicebus_queue" "import" {
  resource_group_name = azurerm_resource_group.test.name
  iothub_name         = azurerm_iothub.test.name
  name                = "acctest"

  connection_string = azurerm_servicebus_queue_authorization_rule.test.primary_connection_string
}
`, r.basic(data))
}

func (t IotHubEndpointServiceBusQueueResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := azure.ParseAzureResourceID(state.ID)
	if err != nil {
		return nil, err
	}
	resourceGroup := id.ResourceGroup
	iothubName := id.Path["IotHubs"]
	endpointName := id.Path["Endpoints"]

	iothub, err := clients.IoTHub.ResourceClient.Get(ctx, resourceGroup, iothubName)
	if err != nil || iothub.Properties == nil || iothub.Properties.Routing == nil || iothub.Properties.Routing.Endpoints == nil {
		return nil, fmt.Errorf("reading IotHuB Endpoint Service Bus (%s): %+v", id, err)
	}

	if endpoints := iothub.Properties.Routing.Endpoints.ServiceBusQueues; endpoints != nil {
		for _, endpoint := range *endpoints {
			if existingEndpointName := endpoint.Name; existingEndpointName != nil {
				if strings.EqualFold(*existingEndpointName, endpointName) {
					return utils.Bool(true), nil
				}
			}
		}
	}

	return utils.Bool(false), nil
}
