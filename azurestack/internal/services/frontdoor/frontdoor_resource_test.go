package frontdoor_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/frontdoor/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type FrontDoorResource struct {
}

func TestAccFrontDoor_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor", "test")
	r := FrontDoorResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("backend_pool_health_probe.0.enabled").HasValue("true"),
				check.That(data.ResourceName).Key("backend_pool_health_probe.0.probe_method").HasValue("GET"),
			),
		},
		{
			Config: r.basicDisabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("backend_pool_health_probe.0.enabled").HasValue("false"),
				check.That(data.ResourceName).Key("backend_pool_health_probe.0.probe_method").HasValue("HEAD"),
			),
		},
		data.ImportStep("explicit_resource_order"),
	})
}

// remove in 3.0
func TestAccFrontDoor_global(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor", "test")
	r := FrontDoorResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.global(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("location").HasValue("global"),
			),
			ExpectNonEmptyPlan: true,
		},
		data.ImportStep("explicit_resource_order"),
	})
}

func TestAccFrontDoor_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor", "test")
	r := FrontDoorResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccFrontDoor_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor", "test")
	r := FrontDoorResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("explicit_resource_order"),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("explicit_resource_order"),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("explicit_resource_order"),
	})
}

func TestAccFrontDoor_multiplePools(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor", "test")
	r := FrontDoorResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multiplePools(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("backend_pool.#").HasValue("2"),
				check.That(data.ResourceName).Key("backend_pool_health_probe.#").HasValue("2"),
				check.That(data.ResourceName).Key("backend_pool_load_balancing.#").HasValue("2"),
				check.That(data.ResourceName).Key("routing_rule.#").HasValue("2"),
			),
		},
		// I am ignoring the returned values here because the ImportStep does not respect my state file reording code,
		// and the API is currently returning these value intermentently out of order causing false positive failures
		// of the test case.
		data.ImportStep("explicit_resource_order", "routing_rule", "backend_pool_load_balancing", "backend_pool_health_probe", "backend_pool"),
	})
}

func TestAccFrontDoor_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor", "test")
	r := FrontDoorResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("explicit_resource_order"),
	})
}

func TestAccFrontDoor_waf(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor", "test")
	r := FrontDoorResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.waf(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("explicit_resource_order"),
	})
}

func TestAccFrontDoor_EnableDisableCache(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor", "test")
	r := FrontDoorResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.EnableCache(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_enabled").HasValue("true"),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_use_dynamic_compression").HasValue("false"),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_query_parameter_strip_directive").HasValue("StripAll"),
			),
		},
		{
			Config: r.DisableCache(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_enabled").HasValue("false"),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_use_dynamic_compression").HasValue("false"),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_query_parameter_strip_directive").HasValue("StripAll"),
			),
		},
		{
			Config: r.EnableCache(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_enabled").HasValue("true"),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_use_dynamic_compression").HasValue("false"),
				check.That(data.ResourceName).Key("routing_rule.0.forwarding_configuration.0.cache_query_parameter_strip_directive").HasValue("StripAll"),
			),
		},
		data.ImportStep("explicit_resource_order"),
	})
}

func (FrontDoorResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.FrontDoorIDInsensitively(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Frontdoor.FrontDoorsClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Front Door %q (Resource Group %q): %v", id.Name, id.ResourceGroup, err)
	}

	return utils.Bool(resp.Properties != nil), nil
}

func (FrontDoorResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%d"
  location = "%s"
}

locals {
  backend_name        = "backend-bing"
  endpoint_name       = "frontend-endpoint"
  health_probe_name   = "health-probe"
  load_balancing_name = "load-balancing-setting"
}

resource "azurerm_frontdoor" "test" {
  name                                         = "acctest-FD-%d"
  resource_group_name                          = azurerm_resource_group.test.name
  enforce_backend_pools_certificate_name_check = false

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name = local.health_probe_name
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name      = local.endpoint_name
    host_name = "acctest-FD-%d.azurefd.net"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (FrontDoorResource) basicDisabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%d"
  location = "%s"
}

locals {
  backend_name        = "backend-bing"
  endpoint_name       = "frontend-endpoint"
  health_probe_name   = "health-probe"
  load_balancing_name = "load-balancing-setting"
}

resource "azurerm_frontdoor" "test" {
  name                                         = "acctest-FD-%d"
  resource_group_name                          = azurerm_resource_group.test.name
  enforce_backend_pools_certificate_name_check = false

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name         = local.health_probe_name
    enabled      = false
    probe_method = "HEAD"
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name      = local.endpoint_name
    host_name = "acctest-FD-%d.azurefd.net"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

// remove in 3.0
func (FrontDoorResource) global(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%d"
  location = "%s"
}

locals {
  backend_name        = "backend-bing"
  endpoint_name       = "frontend-endpoint"
  health_probe_name   = "health-probe"
  load_balancing_name = "load-balancing-setting"
}

resource "azurerm_frontdoor" "test" {
  name                                         = "acctest-FD-%d"
  resource_group_name                          = azurerm_resource_group.test.name
  location                                     = "%s"
  enforce_backend_pools_certificate_name_check = false

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name = local.health_probe_name
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name      = local.endpoint_name
    host_name = "acctest-FD-%d.azurefd.net"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r FrontDoorResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor" "import" {
  name                                         = azurerm_frontdoor.test.name
  resource_group_name                          = azurerm_frontdoor.test.resource_group_name
  enforce_backend_pools_certificate_name_check = azurerm_frontdoor.test.enforce_backend_pools_certificate_name_check

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name = local.health_probe_name
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name      = local.endpoint_name
    host_name = "acctest-FD-%d.azurefd.net"
  }
}
`, r.basic(data), data.RandomInteger)
}

func (FrontDoorResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%d"
  location = "%s"
}

locals {
  backend_name        = "backend-bing"
  endpoint_name       = "frontend-endpoint"
  health_probe_name   = "health-probe"
  load_balancing_name = "load-balancing-setting"
}

resource "azurerm_frontdoor" "test" {
  name                                         = "acctest-FD-%d"
  resource_group_name                          = azurerm_resource_group.test.name
  enforce_backend_pools_certificate_name_check = false
  backend_pools_send_receive_timeout_seconds   = 45

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name = local.health_probe_name
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name      = local.endpoint_name
    host_name = "acctest-FD-%d.azurefd.net"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (FrontDoorResource) waf(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%d"
  location = "%s"
}

locals {
  backend_name        = "backend-bing"
  endpoint_name       = "frontend-endpoint"
  health_probe_name   = "health-probe"
  load_balancing_name = "load-balancing-setting"
}

resource "azurerm_frontdoor_firewall_policy" "test" {
  name                = "acctestwafp%d"
  resource_group_name = azurerm_resource_group.test.name
  mode                = "Prevention"
}

resource "azurerm_frontdoor" "test" {
  name                                         = "acctest-FD-%d"
  resource_group_name                          = azurerm_resource_group.test.name
  enforce_backend_pools_certificate_name_check = false

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name = local.health_probe_name
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name                                    = local.endpoint_name
    host_name                               = "acctest-FD-%d.azurefd.net"
    web_application_firewall_policy_link_id = azurerm_frontdoor_firewall_policy.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (FrontDoorResource) DisableCache(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%d"
  location = "%s"
}

locals {
  backend_name        = "backend-bing"
  endpoint_name       = "frontend-endpoint"
  health_probe_name   = "health-probe"
  load_balancing_name = "load-balancing-setting"
}

resource "azurerm_frontdoor" "test" {
  name                                         = "acctest-FD-%d"
  resource_group_name                          = azurerm_resource_group.test.name
  enforce_backend_pools_certificate_name_check = false

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name = local.health_probe_name
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name      = local.endpoint_name
    host_name = "acctest-FD-%d.azurefd.net"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (FrontDoorResource) EnableCache(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%d"
  location = "%s"
}

locals {
  backend_name        = "backend-bing"
  endpoint_name       = "frontend-endpoint"
  health_probe_name   = "health-probe"
  load_balancing_name = "load-balancing-setting"
}

resource "azurerm_frontdoor" "test" {
  name                                         = "acctest-FD-%d"
  resource_group_name                          = azurerm_resource_group.test.name
  enforce_backend_pools_certificate_name_check = false

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]

    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
      cache_enabled       = true
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name = local.health_probe_name
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name      = local.endpoint_name
    host_name = "acctest-FD-%d.azurefd.net"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (FrontDoorResource) multiplePools(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%[1]d"
  location = "%s"
}

resource "azurerm_frontdoor" "test" {
  name                                         = "acctest-FD-%[1]d"
  resource_group_name                          = azurerm_resource_group.test.name
  enforce_backend_pools_certificate_name_check = false

  frontend_endpoint {
    name      = "acctest-FD-%[1]d-default-FE"
    host_name = "acctest-FD-%[1]d.azurefd.net"
  }

  # --- Pool 1

  routing_rule {
    name               = "acctest-FD-%[1]d-bing-RR"
    accepted_protocols = ["Https"]
    patterns_to_match  = ["/poolBing/*"]
    frontend_endpoints = ["acctest-FD-%[1]d-default-FE"]

    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "acctest-FD-%[1]d-pool-bing"
      cache_enabled       = true
    }
  }

  backend_pool_load_balancing {
    name                            = "acctest-FD-%[1]d-bing-LB"
    additional_latency_milliseconds = 0
    sample_size                     = 4
    successful_samples_required     = 2
  }

  backend_pool_health_probe {
    name         = "acctest-FD-%[1]d-bing-HP"
    protocol     = "Https"
    enabled      = true
    probe_method = "HEAD"
  }

  backend_pool {
    name                = "acctest-FD-%[1]d-pool-bing"
    load_balancing_name = "acctest-FD-%[1]d-bing-LB"
    health_probe_name   = "acctest-FD-%[1]d-bing-HP"

    backend {
      host_header = "bing.com"
      address     = "bing.com"
      http_port   = 80
      https_port  = 443
      weight      = 75
      enabled     = true
    }
  }

  # --- Pool 2

  routing_rule {
    name               = "acctest-FD-%[1]d-google-RR"
    accepted_protocols = ["Https"]
    patterns_to_match  = ["/poolGoogle/*"]
    frontend_endpoints = ["acctest-FD-%[1]d-default-FE"]

    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "acctest-FD-%[1]d-pool-google"
      cache_enabled       = true
    }
  }

  backend_pool_load_balancing {
    name                            = "acctest-FD-%[1]d-google-LB"
    additional_latency_milliseconds = 0
    sample_size                     = 4
    successful_samples_required     = 2
  }

  backend_pool_health_probe {
    name     = "acctest-FD-%[1]d-google-HP"
    protocol = "Https"
  }

  backend_pool {
    name                = "acctest-FD-%[1]d-pool-google"
    load_balancing_name = "acctest-FD-%[1]d-google-LB"
    health_probe_name   = "acctest-FD-%[1]d-google-HP"

    backend {
      host_header = "google.com"
      address     = "google.com"
      http_port   = 80
      https_port  = 443
      weight      = 75
      enabled     = true
    }
  }
}
`, data.RandomInteger, data.Locations.Primary)
}
