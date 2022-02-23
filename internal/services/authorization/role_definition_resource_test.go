package authorization_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type RoleDefinitionResource struct{}

func TestAccAzureRMRoleDefinition_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")
	r := RoleDefinitionResource{}

	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Errorf("got an error when generating UUID, %s", err)
	}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(id, data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMRoleDefinition_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Errorf("got an error when generating UUID, %s", err)
	}

	r := RoleDefinitionResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(id, data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(func(data acceptance.TestData) string {
			return r.requiresImport(id, data)
		}),
	})
}

func TestAccAzureRMRoleDefinition_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")
	r := RoleDefinitionResource{}
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Errorf("got an error when generating UUID, %s", err)
	}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(id, data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("role_definition_id", "scope"),
	})
}

func TestAccAzureRMRoleDefinition_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")
	r := RoleDefinitionResource{}
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Errorf("got an error when generating UUID, %s", err)
	}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(id, data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updated(id, data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMRoleDefinition_updateEmptyId(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")

	r := RoleDefinitionResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.emptyId(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateEmptyId(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMRoleDefinition_emptyName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")
	r := RoleDefinitionResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.emptyId(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

// func TestAccAzureRMRoleDefinition_managementGroup(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")
// 	r := RoleDefinitionResource{}
// 	id, err := uuid.GenerateUUID()
// 	if err != nil {
// 		t.Errorf("got an error when generating UUID, %s", err)
// 	}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.managementGroup(id, data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 			),
// 		},
// 		data.ImportStep("scope"),
// 	})
// }

func TestAccAzureRMRoleDefinition_assignToSmallerScope(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")
	r := RoleDefinitionResource{}
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Errorf("got an error when generating UUID, %s", err)
	}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.assignToSmallerScope(id, data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMRoleDefinition_noAssignableScope(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_role_definition", "test")
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Errorf("got an error when generating UUID, %s", err)
	}

	r := RoleDefinitionResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.noAssignableScope(id, data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (RoleDefinitionResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	scope := state.Attributes["scope"]
	roleDefinitionId := state.Attributes["role_definition_id"]

	resp, err := client.Authorization.RoleDefinitionsClient.Get(ctx, scope, roleDefinitionId)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("Bad: Get on roleDefinitionsClient: %+v", err)
	}

	return utils.Bool(true), nil
}

func (RoleDefinitionResource) basic(id string, data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "primary" {
}

resource "azurestack_role_definition" "test" {
  role_definition_id = "%s"
  name               = "acctestrd-%d"
  scope              = data.azurestack_subscription.primary.id

  permissions {
    actions     = ["*"]
    not_actions = []
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}
`, id, data.RandomInteger)
}

func (r RoleDefinitionResource) requiresImport(id string, data acceptance.TestData) string {
	template := r.basic(id, data)
	return fmt.Sprintf(`
%s

resource "azurestack_role_definition" "import" {
  role_definition_id = azurestack_role_definition.test.role_definition_id
  name               = azurestack_role_definition.test.name
  scope              = azurestack_role_definition.test.scope

  permissions {
    actions     = ["*"]
    not_actions = []
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}
`, template)
}

func (RoleDefinitionResource) complete(id string, data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "primary" {
}

resource "azurestack_role_definition" "test" {
  role_definition_id = "%s"
  name               = "acctestrd-%d"
  scope              = data.azurestack_subscription.primary.id
  description        = "Acceptance Test Role Definition"

  permissions {
    actions     = ["*"]
    not_actions = ["Microsoft.Authorization/*/read"]
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}
`, id, data.RandomInteger)
}

func (RoleDefinitionResource) updated(id string, data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "primary" {
}

resource "azurestack_role_definition" "test" {
  role_definition_id = "%s"
  name               = "acctestrd-%d-updated"
  scope              = data.azurestack_subscription.primary.id
  description        = "Acceptance Test Role Definition"

  permissions {
    actions     = ["*"]
    not_actions = ["Microsoft.Authorization/*/read"]
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}
`, id, data.RandomInteger)
}

func (RoleDefinitionResource) emptyId(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "primary" {
}

resource "azurestack_role_definition" "test" {
  name  = "acctestrd-%d"
  scope = data.azurestack_subscription.primary.id

  permissions {
    actions     = ["*"]
    not_actions = []
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}
`, data.RandomInteger)
}

func (RoleDefinitionResource) updateEmptyId(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "primary" {
}

resource "azurestack_role_definition" "test" {
  name  = "acctestrd-%d"
  scope = data.azurestack_subscription.primary.id

  permissions {
    actions     = ["*"]
    not_actions = ["Microsoft.Authorization/*/read"]
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}
`, data.RandomInteger)
}

// nolint: unused
func (RoleDefinitionResource) managementGroup(id string, data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "primary" {
}

resource "azurestack_management_group" "test" {
}

resource "azurestack_role_definition" "test" {
  role_definition_id = "%s"
  name               = "acctestrd-%d"
  scope              = azurestack_management_group.test.id

  permissions {
    actions     = ["*"]
    not_actions = []
  }

  assignable_scopes = [
    azurestack_management_group.test.id,
    data.azurestack_subscription.primary.id,
  ]
}
`, id, data.RandomInteger)
}

func (RoleDefinitionResource) assignToSmallerScope(id string, data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "primary" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestrg-%d"
  location = "%s"
}

resource "azurestack_role_definition" "test" {
  role_definition_id = "%s"
  name               = "acctestrd-%d"
  scope              = data.azurestack_subscription.primary.id

  permissions {
    actions     = ["*"]
    not_actions = []
  }

  assignable_scopes = [
    azurestack_resource_group.test.id
  ]
}
`, data.RandomInteger, data.Locations.Primary, id, data.RandomInteger)
}

func (RoleDefinitionResource) noAssignableScope(id string, data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "primary" {
}

resource "azurestack_role_definition" "test" {
  role_definition_id = "%s"
  name               = "acctestrd-%d"
  scope              = data.azurestack_subscription.primary.id

  permissions {
    actions     = ["*"]
    not_actions = []
  }
}
`, id, data.RandomInteger)
}
