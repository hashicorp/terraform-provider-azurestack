package managementgroup_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/preview/resources/mgmt/2018-03-01-preview/managementgroups"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/managementgroup/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type ManagementGroupSubscriptionAssociation struct{}

func TestAccManagementGroupSubscriptionAssociation_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group_subscription_association", "test")

	r := ManagementGroupSubscriptionAssociation{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccManagementGroupSubscriptionAssociation_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group_subscription_association", "test")

	r := ManagementGroupSubscriptionAssociation{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (r ManagementGroupSubscriptionAssociation) basic() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "test" {
  subscription_id = %q
}

resource "azurerm_management_group" "test" {
}

resource "azurerm_management_group_subscription_association" "test" {
  management_group_id = azurerm_management_group.test.id
  subscription_id     = data.azurerm_subscription.test.id
}

`, os.Getenv("ARM_SUBSCRIPTION_ID_ALT"))
}

func (r ManagementGroupSubscriptionAssociation) requiresImport(_ acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_management_group_subscription_association" "import" {
  management_group_id = azurerm_management_group_subscription_association.test.management_group_id
  subscription_id     = azurerm_management_group_subscription_association.test.subscription_id
}

`, r.basic())
}

func (r ManagementGroupSubscriptionAssociation) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ManagementGroupSubscriptionAssociationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.ManagementGroups.GroupsClient.Get(ctx, id.ManagementGroup, "children", utils.Bool(false), "", "no-cache")
	if err != nil {
		return nil, fmt.Errorf("retrieving Management Group to check for Subscription Association: %+v", err)
	}

	if resp.Properties == nil || resp.Properties.Children == nil {
		return utils.Bool(false), nil
	}

	present := false
	for _, v := range *resp.Children {
		if v.Type == managementgroups.Type1Subscriptions && v.Name != nil && *v.Name == id.SubscriptionId {
			present = true
		}
	}

	return utils.Bool(present), nil
}
