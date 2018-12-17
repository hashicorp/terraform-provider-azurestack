## 0.5.0 (Unreleased)

IMPROVEMENTS:

* dependencies: upgrading to Go 1.11 [GH-53]
* dependencies: upgrading to version 21.3.0 of github.com/Azure/azure-sdk-for-go [GH-53]
* authentication: switching to use the new authentication package [GH-54]
* authentication: support for Client Certificate authentication [GH-56]

## 0.4.0 (September 18, 2018)

* **New Resource:** `azurestack_route_table` ([#26](https://github.com/terraform-providers/terraform-provider-azurestack/issues/26))
* **New Resource:** `azurestack_route` ([#27](https://github.com/terraform-providers/terraform-provider-azurestack/issues/27))
* **New Resource:** `azurestack_template_deployment` ([#33](https://github.com/terraform-providers/terraform-provider-azurestack/issues/33))
* **New Resource:** `azurestack_virtual_machine_scale_set` ([#15](https://github.com/terraform-providers/terraform-provider-azurestack/issues/15))
* **New Resource:** `azurestack_virtual_network_gateway` ([#31](https://github.com/terraform-providers/terraform-provider-azurestack/issues/31))
* **New Resource:** `azurestack_virtual_network_gateway_connection` ([#35](https://github.com/terraform-providers/terraform-provider-azurestack/issues/35))
* **New Data Source:** `azurestack_public_ip` ([#34](https://github.com/terraform-providers/terraform-provider-azurestack/issues/34))
* **New Data Source:** `azurestack_route_table` ([#26](https://github.com/terraform-providers/terraform-provider-azurestack/issues/26))
* **New Data Source:** `azurestack_subnet` ([#34](https://github.com/terraform-providers/terraform-provider-azurestack/issues/34))
* **New Data Source:** `azurestack_virtual_network_gateway` ([#31](https://github.com/terraform-providers/terraform-provider-azurestack/issues/31))

IMPROVEMENTS:

* dependencies: upgrading to `v20.1.0` of `github.com/Azure/azure-sdk-for-go` ([#38](https://github.com/terraform-providers/terraform-provider-azurestack/issues/38))
* dependencies: upgrading to `v10.15.4` of `github.com/Azure/go-autorest` ([#38](https://github.com/terraform-providers/terraform-provider-azurestack/issues/38))

BUG FIXES:

* `azurestack_public_ip` - now correctly reading and importing the `idle_timeout_in_minutes` property ([#42](https://github.com/terraform-providers/terraform-provider-azurestack/issues/42))

## 0.3.0 (August 13, 2018)

* **New Resource:** `azurestack_lb` ([#16](https://github.com/terraform-providers/terraform-provider-azurestack/issues/16))
* **New Resource:** `azurestack_lb_backend_address_pool` ([#21](https://github.com/terraform-providers/terraform-provider-azurestack/issues/21))
* **New Resource:** `azurestack_lb_nat_rule` ([#22](https://github.com/terraform-providers/terraform-provider-azurestack/issues/22))
* **New Resource:** `azurestack_lb_nat_pool` ([#24](https://github.com/terraform-providers/terraform-provider-azurestack/issues/24))
* **New Resource:** `azurestack_lb_probe` ([#23](https://github.com/terraform-providers/terraform-provider-azurestack/issues/23))
* **New Resource:** `azurestack_lb_rule` ([#25](https://github.com/terraform-providers/terraform-provider-azurestack/issues/25))

## 0.2.0 (July 26, 2018)

* **New Resource:** `azurestack_local_network_gateway` ([#13](https://github.com/terraform-providers/terraform-provider-azurestack/issues/13))
* **New Data Source:** `azurestack_client_config` ([#9](https://github.com/terraform-providers/terraform-provider-azurestack/issues/9))

## 0.1.0 (June 19, 2018) 

* Initial Release
