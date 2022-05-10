package authorization_test

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type ClientConfigDataSource struct{}

func TestAccClientConfigDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_client_config", "current")
	clientId := os.Getenv("ARM_CLIENT_ID")
	tenantId := os.Getenv("ARM_TENANT_ID")
	subscriptionId := os.Getenv("ARM_SUBSCRIPTION_ID")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: ClientConfigDataSource{}.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("client_id").HasValue(clientId),
				check.That(data.ResourceName).Key("tenant_id").HasValue(tenantId),
				check.That(data.ResourceName).Key("subscription_id").HasValue(subscriptionId),
				testAccCheckRegexSIDs("data.azurestack_client_config.current"),
			),
		},
	})
}

func testAccCheckRegexSIDs(resourceName string) pluginsdk.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		// TODO: Remove the validation until the bug of object_id in ADFS environment is fixed
		if strings.EqualFold(rs.Primary.Attributes["tenant_id"], "adfs") && rs.Primary.Attributes["object_id"] == "" {
			log.Println("[WARN] Validation passed when tenant_id is adfs and object_id is empty")
			// Will be passed until the bug of object id is fixed in ADFS environment
			return nil
		}

		objectIdRegex := regexp.MustCompile("^[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$")
		adfsID := regexp.MustCompile(`^S-\d-(\d+-){1,14}\d+$`)

		if !objectIdRegex.MatchString(rs.Primary.Attributes["object_id"]) {
			// For ADFS validation
			if !adfsID.MatchString(rs.Primary.Attributes["object_id"]) {
				return fmt.Errorf("object_id didn't match %v or %v, got %v", objectIdRegex, adfsID, rs.Primary.Attributes["object_id"])
			}
		}

		return nil
	}
}

func (d ClientConfigDataSource) basic() string {
	return `
data "azurestack_client_config" "current" {}
`
}
