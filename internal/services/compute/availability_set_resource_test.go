package compute_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type AvailabilitySetResource struct{}

func TestAccAvailabilitySet_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_availability_set", "test")
	r := AvailabilitySetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("platform_update_domain_count").HasValue("5"),
				check.That(data.ResourceName).Key("platform_fault_domain_count").HasValue("3"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAvailabilitySet_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_availability_set", "test")
	r := AvailabilitySetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("platform_update_domain_count").HasValue("5"),
				check.That(data.ResourceName).Key("platform_fault_domain_count").HasValue("3"),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_availability_set"),
		},
	})
}

func TestAccAvailabilitySet_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_availability_set", "test")
	r := AvailabilitySetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccAvailabilitySet_withTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_availability_set", "test")
	r := AvailabilitySetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withTags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("2"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("Production"),
				check.That(data.ResourceName).Key("tags.cost_center").HasValue("MSFT"),
			),
		},
		{
			Config: r.withUpdatedTags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("staging"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAvailabilitySet_withDomainCounts(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_availability_set", "test")
	r := AvailabilitySetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withDomainCounts(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("platform_update_domain_count").HasValue("3"),
				check.That(data.ResourceName).Key("platform_fault_domain_count").HasValue("3"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAvailabilitySet_unmanaged(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_availability_set", "test")
	r := AvailabilitySetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.unmanaged(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("managed").HasValue("false"),
			),
		},
		data.ImportStep(),
	})
}

func (AvailabilitySetResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.AvailabilitySetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.AvailabilitySetsClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Availability Set %q", id.String())
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (AvailabilitySetResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.AvailabilitySetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Compute.AvailabilitySetsClient.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.WasNotFound(resp.Response) {
			return nil, fmt.Errorf("deleting on availSetClient: %+v", err)
		}
	}

	return pointer.FromBool(true), nil
}

func (AvailabilitySetResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                = "acctestavset-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r AvailabilitySetResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_availability_set" "import" {
  name                = azurestack_availability_set.test.name
  location            = azurestack_availability_set.test.location
  resource_group_name = azurestack_availability_set.test.resource_group_name
}
`, r.basic(data))
}

func (AvailabilitySetResource) withTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                = "acctestavset-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (AvailabilitySetResource) withUpdatedTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                = "acctestavset-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (AvailabilitySetResource) withDomainCounts(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                         = "acctestavset-%d"
  location                     = azurestack_resource_group.test.location
  resource_group_name          = azurestack_resource_group.test.name
  platform_update_domain_count = 3
  platform_fault_domain_count  = 3
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (AvailabilitySetResource) unmanaged(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                         = "acctestavset-%d"
  location                     = azurestack_resource_group.test.location
  resource_group_name          = azurestack_resource_group.test.name
  platform_update_domain_count = 3
  platform_fault_domain_count  = 3
  managed                      = false
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
