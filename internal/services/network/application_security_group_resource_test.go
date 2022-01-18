package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type ApplicationSecurityGroupResource struct{}

func TestAccApplicationSecurityGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_application_security_group", "test")
	r := ApplicationSecurityGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func TestAccApplicationSecurityGroup_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_application_security_group", "test")
	r := ApplicationSecurityGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_application_security_group"),
		},
	})
}

func TestAccApplicationSecurityGroup_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_application_security_group", "test")
	r := ApplicationSecurityGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.Hello").HasValue("World"),
			),
		},
	})
}

func TestAccApplicationSecurityGroup_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_application_security_group", "test")
	r := ApplicationSecurityGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.Hello").HasValue("World"),
			),
		},
	})
}

func (t ApplicationSecurityGroupResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ApplicationSecurityGroupID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.ApplicationSecurityGroupsClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %+v", *id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (ApplicationSecurityGroupResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_application_security_group" "test" {
  name                = "acctest-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r ApplicationSecurityGroupResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_application_security_group" "import" {
  name                = azurestack_application_security_group.test.name
  location            = azurestack_application_security_group.test.location
  resource_group_name = azurestack_application_security_group.test.resource_group_name
}
`, r.basic(data))
}

func (ApplicationSecurityGroupResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_application_security_group" "test" {
  name                = "acctest-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  tags = {
    Hello = "World"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
