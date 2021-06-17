package eventgrid_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/eventgrid/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type EventGridEventSubscriptionResource struct {
}

func TestAccEventGridEventSubscription_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("event_delivery_schema").HasValue("EventGridSchema"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccEventGridEventSubscription_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_eventgrid_event_subscription"),
		},
	})
}

func TestAccEventGridEventSubscription_eventHubID(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.eventHubID(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("event_delivery_schema").HasValue("CloudEventSchemaV1_0"),
				check.That(data.ResourceName).Key("eventhub_endpoint_id").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccEventGridEventSubscription_serviceBusQueueID(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.serviceBusQueueID(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("event_delivery_schema").HasValue("CloudEventSchemaV1_0"),
				check.That(data.ResourceName).Key("service_bus_queue_endpoint_id").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccEventGridEventSubscription_serviceBusTopicID(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.serviceBusTopicID(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("event_delivery_schema").HasValue("CloudEventSchemaV1_0"),
				check.That(data.ResourceName).Key("service_bus_topic_endpoint_id").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccEventGridEventSubscription_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("event_delivery_schema").HasValue("EventGridSchema"),
				check.That(data.ResourceName).Key("storage_queue_endpoint.#").HasValue("1"),
				check.That(data.ResourceName).Key("storage_blob_dead_letter_destination.#").HasValue("1"),
				check.That(data.ResourceName).Key("included_event_types.0").HasValue("Microsoft.Resources.ResourceWriteSuccess"),
				check.That(data.ResourceName).Key("retry_policy.0.max_delivery_attempts").HasValue("11"),
				check.That(data.ResourceName).Key("retry_policy.0.event_time_to_live").HasValue("11"),
				check.That(data.ResourceName).Key("labels.0").HasValue("test"),
				check.That(data.ResourceName).Key("labels.2").HasValue("test2"),
			),
		},
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("included_event_types.0").HasValue("Microsoft.Storage.BlobCreated"),
				check.That(data.ResourceName).Key("included_event_types.1").HasValue("Microsoft.Storage.BlobDeleted"),
				check.That(data.ResourceName).Key("subject_filter.0.subject_ends_with").HasValue(".jpg"),
				check.That(data.ResourceName).Key("subject_filter.0.subject_begins_with").HasValue("test/test"),
				check.That(data.ResourceName).Key("retry_policy.0.max_delivery_attempts").HasValue("10"),
				check.That(data.ResourceName).Key("retry_policy.0.event_time_to_live").HasValue("12"),
				check.That(data.ResourceName).Key("labels.0").HasValue("test4"),
				check.That(data.ResourceName).Key("labels.2").HasValue("test6"),
			),
		},
	})
}

func TestAccEventGridEventSubscription_filter(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.filter(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("included_event_types.0").HasValue("Microsoft.Storage.BlobCreated"),
				check.That(data.ResourceName).Key("included_event_types.1").HasValue("Microsoft.Storage.BlobDeleted"),
				check.That(data.ResourceName).Key("subject_filter.0.subject_ends_with").HasValue(".jpg"),
				check.That(data.ResourceName).Key("subject_filter.0.subject_begins_with").HasValue("test/test"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccEventGridEventSubscription_advancedFilter(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.advancedFilter(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("advanced_filter.0.bool_equals.0.key").HasValue("subject"),
				check.That(data.ResourceName).Key("advanced_filter.0.bool_equals.0.value").HasValue("true"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_greater_than.0.key").HasValue("data.metadataVersion"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_greater_than.0.value").HasValue("1"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_greater_than_or_equals.0.key").HasValue("data.contentLength"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_greater_than_or_equals.0.value").HasValue("42"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_less_than.0.key").HasValue("data.contentLength"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_less_than.0.value").HasValue("42.1"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_less_than_or_equals.0.key").HasValue("data.metadataVersion"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_less_than_or_equals.0.value").HasValue("2"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_in.0.key").HasValue("data.contentLength"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_in.0.values.0").HasValue("0"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_not_in.0.key").HasValue("data.contentLength"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_not_in.0.values.0").HasValue("5"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_begins_with.0.key").HasValue("subject"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_begins_with.0.values.0").HasValue("foo"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_ends_with.0.key").HasValue("subject"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_ends_with.0.values.0").HasValue("bar"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_contains.0.key").HasValue("data.contentType"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_contains.0.values.0").HasValue("application"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_in.0.key").HasValue("data.blobType"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_in.0.values.0").HasValue("Block"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_not_in.0.key").HasValue("data.blobType"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_not_in.0.values.0").HasValue("Page"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccEventGridEventSubscription_advancedFilterMaxItems(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_eventgrid_event_subscription", "test")
	r := EventGridEventSubscriptionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.advancedFilterMaxItems(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("advanced_filter.0.bool_equals.0.key").HasValue("subject"),
				check.That(data.ResourceName).Key("advanced_filter.0.bool_equals.0.value").HasValue("true"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_greater_than.0.key").HasValue("data.metadataVersion"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_greater_than.0.value").HasValue("2"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_greater_than_or_equals.0.key").HasValue("data.contentLength"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_greater_than_or_equals.0.value").HasValue("3"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_less_than.0.key").HasValue("data.contentLength"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_less_than.0.value").HasValue("4"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_less_than_or_equals.0.key").HasValue("data.metadataVersion"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_less_than_or_equals.0.value").HasValue("5"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_in.0.key").HasValue("data.contentLength"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_in.0.values.0").HasValue("6"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_in.0.values.1").HasValue("7"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_in.0.values.2").HasValue("8"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_not_in.0.key").HasValue("data.contentLength"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_not_in.0.values.0").HasValue("9"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_not_in.0.values.1").HasValue("10"),
				check.That(data.ResourceName).Key("advanced_filter.0.number_not_in.0.values.2").HasValue("11"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_begins_with.0.key").HasValue("subject"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_begins_with.0.values.0").HasValue("12"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_begins_with.0.values.1").HasValue("13"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_begins_with.0.values.2").HasValue("14"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_ends_with.0.key").HasValue("subject"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_ends_with.0.values.0").HasValue("15"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_ends_with.0.values.1").HasValue("16"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_ends_with.0.values.2").HasValue("17"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_contains.0.key").HasValue("data.contentType"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_contains.0.values.0").HasValue("18"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_contains.0.values.1").HasValue("19"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_contains.0.values.2").HasValue("20"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_in.0.key").HasValue("data.blobType"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_in.0.values.0").HasValue("21"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_in.0.values.1").HasValue("22"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_in.0.values.2").HasValue("23"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_not_in.0.key").HasValue("data.blobType"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_not_in.0.values.0").HasValue("24"),
				check.That(data.ResourceName).Key("advanced_filter.0.string_not_in.0.values.1").HasValue("25"),
			),
		},
		data.ImportStep(),
	})
}

func (EventGridEventSubscriptionResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.EventSubscriptionID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.EventGrid.EventSubscriptionsClient.Get(ctx, id.Scope, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving EventGrid Event Subscription %q (scope: %q): %+v", id.Name, id.Scope, err)
	}

	return utils.Bool(resp.EventSubscriptionProperties != nil), nil
}

func (EventGridEventSubscriptionResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-eg-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurerm_storage_queue" "test" {
  name                 = "mysamplequeue-%d"
  storage_account_name = azurerm_storage_account.test.name
}

resource "azurerm_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurerm_storage_account.test.name
  container_access_type = "private"
}

resource "azurerm_storage_blob" "test" {
  name = "herpderp1.vhd"

  storage_account_name   = azurerm_storage_account.test.name
  storage_container_name = azurerm_storage_container.test.name

  type = "Page"
  size = 5120
}

resource "azurerm_eventgrid_event_subscription" "test" {
  name  = "acctesteg-%d"
  scope = azurerm_resource_group.test.id

  storage_queue_endpoint {
    storage_account_id = azurerm_storage_account.test.id
    queue_name         = azurerm_storage_queue.test.name
  }

  storage_blob_dead_letter_destination {
    storage_account_id          = azurerm_storage_account.test.id
    storage_blob_container_name = azurerm_storage_container.test.name
  }

  retry_policy {
    event_time_to_live    = 11
    max_delivery_attempts = 11
  }

  labels = ["test", "test1", "test2"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, data.RandomInteger)
}

func (EventGridEventSubscriptionResource) requiresImport(data acceptance.TestData) string {
	template := EventGridEventSubscriptionResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_eventgrid_event_subscription" "import" {
  name  = azurerm_eventgrid_event_subscription.test.name
  scope = azurerm_eventgrid_event_subscription.test.scope
}
`, template)
}

func (EventGridEventSubscriptionResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-eg-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurerm_storage_queue" "test" {
  name                 = "mysamplequeue-%d"
  storage_account_name = azurerm_storage_account.test.name
}

resource "azurerm_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurerm_storage_account.test.name
  container_access_type = "private"
}

resource "azurerm_storage_blob" "test" {
  name = "herpderp1.vhd"

  storage_account_name   = azurerm_storage_account.test.name
  storage_container_name = azurerm_storage_container.test.name

  type = "Page"
  size = 5120
}

resource "azurerm_eventgrid_event_subscription" "test" {
  name  = "acctest-eg-%d"
  scope = azurerm_resource_group.test.id

  storage_queue_endpoint {
    storage_account_id = azurerm_storage_account.test.id
    queue_name         = azurerm_storage_queue.test.name
  }

  storage_blob_dead_letter_destination {
    storage_account_id          = azurerm_storage_account.test.id
    storage_blob_container_name = azurerm_storage_container.test.name
  }

  retry_policy {
    event_time_to_live    = 12
    max_delivery_attempts = 10
  }

  subject_filter {
    subject_begins_with = "test/test"
    subject_ends_with   = ".jpg"
  }

  included_event_types = ["Microsoft.Storage.BlobCreated", "Microsoft.Storage.BlobDeleted"]
  labels               = ["test4", "test5", "test6"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, data.RandomInteger)
}

func (EventGridEventSubscriptionResource) eventHubID(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-eg-%d"
  location = "%s"
}

resource "azurerm_eventhub_namespace" "test" {
  name                = "acctesteventhubnamespace-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Basic"
}

resource "azurerm_eventhub" "test" {
  name                = "acctesteventhub-%d"
  namespace_name      = azurerm_eventhub_namespace.test.name
  resource_group_name = azurerm_resource_group.test.name
  partition_count     = 2
  message_retention   = 1
}

resource "azurerm_eventgrid_event_subscription" "test" {
  name                  = "acctest-eg-%d"
  scope                 = azurerm_resource_group.test.id
  event_delivery_schema = "CloudEventSchemaV1_0"

  eventhub_endpoint_id = azurerm_eventhub.test.id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (EventGridEventSubscriptionResource) serviceBusQueueID(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-eg-%d"
  location = "%s"
}

resource "azurerm_servicebus_namespace" "example" {
  name                = "acctestservicebusnamespace-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Basic"
}
resource "azurerm_servicebus_queue" "test" {
  name                = "acctestservicebusqueue-%d"
  resource_group_name = azurerm_resource_group.test.name
  namespace_name      = azurerm_servicebus_namespace.example.name
  enable_partitioning = true
}
resource "azurerm_eventgrid_event_subscription" "test" {
  name                          = "acctest-eg-%d"
  scope                         = azurerm_resource_group.test.id
  event_delivery_schema         = "CloudEventSchemaV1_0"
  service_bus_queue_endpoint_id = azurerm_servicebus_queue.test.id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (EventGridEventSubscriptionResource) serviceBusTopicID(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-eg-%d"
  location = "%s"
}
resource "azurerm_servicebus_namespace" "example" {
  name                = "acctestservicebusnamespace-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Standard"
}
resource "azurerm_servicebus_topic" "test" {
  name                = "acctestservicebustopic-%d"
  resource_group_name = azurerm_resource_group.test.name
  namespace_name      = azurerm_servicebus_namespace.example.name
  enable_partitioning = true
}
resource "azurerm_eventgrid_event_subscription" "test" {
  name                          = "acctest-eg-%d"
  scope                         = azurerm_resource_group.test.id
  event_delivery_schema         = "CloudEventSchemaV1_0"
  service_bus_topic_endpoint_id = azurerm_servicebus_topic.test.id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (EventGridEventSubscriptionResource) filter(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-eg-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurerm_storage_queue" "test" {
  name                 = "mysamplequeue-%d"
  storage_account_name = azurerm_storage_account.test.name
}

resource "azurerm_eventgrid_event_subscription" "test" {
  name  = "acctest-eg-%d"
  scope = azurerm_resource_group.test.id

  storage_queue_endpoint {
    storage_account_id = azurerm_storage_account.test.id
    queue_name         = azurerm_storage_queue.test.name
  }

  included_event_types = ["Microsoft.Storage.BlobCreated", "Microsoft.Storage.BlobDeleted"]

  subject_filter {
    subject_begins_with = "test/test"
    subject_ends_with   = ".jpg"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, data.RandomInteger)
}

func (EventGridEventSubscriptionResource) advancedFilter(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-eg-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurerm_storage_queue" "test" {
  name                 = "mysamplequeue-%d"
  storage_account_name = azurerm_storage_account.test.name
}

resource "azurerm_eventgrid_event_subscription" "test" {
  name  = "acctesteg-%d"
  scope = azurerm_storage_account.test.id

  storage_queue_endpoint {
    storage_account_id = azurerm_storage_account.test.id
    queue_name         = azurerm_storage_queue.test.name
  }

  advanced_filter {
    bool_equals {
      key   = "subject"
      value = true
    }
    number_greater_than {
      key   = "data.metadataVersion"
      value = 1
    }
    number_greater_than_or_equals {
      key   = "data.contentLength"
      value = 42.0
    }
    number_less_than {
      key   = "data.contentLength"
      value = 42.1
    }
    number_less_than_or_equals {
      key   = "data.metadataVersion"
      value = 2
    }
    number_in {
      key    = "data.contentLength"
      values = [0, 1, 1, 2, 3]
    }
    number_not_in {
      key    = "data.contentLength"
      values = [5, 8, 13, 21, 34]
    }
    string_begins_with {
      key    = "subject"
      values = ["foo"]
    }
    string_ends_with {
      key    = "subject"
      values = ["bar"]
    }
    string_contains {
      key    = "data.contentType"
      values = ["application", "octet-stream"]
    }
    string_in {
      key    = "data.blobType"
      values = ["Block"]
    }
    string_not_in {
      key    = "data.blobType"
      values = ["Page"]
    }
  }

}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, data.RandomInteger)
}

func (EventGridEventSubscriptionResource) advancedFilterMaxItems(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-eg-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurerm_storage_queue" "test" {
  name                 = "mysamplequeue-%d"
  storage_account_name = azurerm_storage_account.test.name
}

resource "azurerm_eventgrid_event_subscription" "test" {
  name  = "acctesteg-%d"
  scope = azurerm_storage_account.test.id

  storage_queue_endpoint {
    storage_account_id = azurerm_storage_account.test.id
    queue_name         = azurerm_storage_queue.test.name
  }

  advanced_filter {
    bool_equals {
      key   = "subject"
      value = true
    }
    number_greater_than {
      key   = "data.metadataVersion"
      value = 2
    }
    number_greater_than_or_equals {
      key   = "data.contentLength"
      value = 3
    }
    number_less_than {
      key   = "data.contentLength"
      value = 4
    }
    number_less_than_or_equals {
      key   = "data.metadataVersion"
      value = 5
    }
    number_in {
      key    = "data.contentLength"
      values = [6, 7, 8]
    }
    number_not_in {
      key    = "data.contentLength"
      values = [9, 10, 11]
    }
    string_begins_with {
      key    = "subject"
      values = ["12", "13", "14"]
    }
    string_ends_with {
      key    = "subject"
      values = ["15", "16", "17"]
    }
    string_contains {
      key    = "data.contentType"
      values = ["18", "19", "20"]
    }
    string_in {
      key    = "data.blobType"
      values = ["21", "22", "23"]
    }
    string_not_in {
      key    = "data.blobType"
      values = ["24", "25"]
    }
  }

}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, data.RandomInteger)
}
