package azurestack

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/authentication"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func buildConfigForSweepers() (*ArmClient, error) {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	tenantID := os.Getenv("AZURE_TENANT_ID")

	if subscriptionID == "" || clientID == "" || clientSecret == "" || tenantID == "" {
		return nil, fmt.Errorf("AZURE_SUBSCRIPTION_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID and AZURE_LOCATION must be set for acceptance tests")
	}

	config := &authentication.Config{
		SubscriptionID: subscriptionID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		TenantID:       tenantID,
	}

	return getArmClient(config)
}

func shouldSweepAcceptanceTestResource(name string, resourceLocation string, region string) bool {
	loweredName := strings.ToLower(name)

	if !strings.HasPrefix(loweredName, "acctest") {
		log.Printf("Ignoring Resource %q as it doesn't start with `acctest`", name)
		return false
	}

	normalisedResourceLocation := azureRMNormalizeLocation(resourceLocation)
	normalisedRegion := azureRMNormalizeLocation(region)

	if normalisedResourceLocation != normalisedRegion {
		log.Printf("Region %q isn't %q - skipping", normalisedResourceLocation, normalisedRegion)
		return false
	}

	return true
}
