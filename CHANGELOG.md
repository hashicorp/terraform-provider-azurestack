## 0.9.0 (Unreleased)
## 0.8.1 (July 09, 2019)

BUG FIXES:

* `azurestack_storage_account` - handling the casing of the Storage Account ID changing in Azure Stack 1905 ([#99](https://github.com/terraform-providers/terraform-provider-azurestack/issues/99))

## 0.8.0 (June 24, 2019)

FEATURES:

* **New Data Source:** `azurestack_platform_image` ([#85](https://github.com/terraform-providers/terraform-provider-azurestack/issues/85))
* **New Resource:** `azurestack_managed_disk` ([#85](https://github.com/terraform-providers/terraform-provider-azurestack/issues/85))

IMPROVEMENTS:

* dependencies: Changing to the `2019-03-01` profile ([#84](https://github.com/terraform-providers/terraform-provider-azurestack/issues/84))
* dependencies: upgrading to `v30.0.0` of `github.com/Azure/azure-sdk-for-go` ([#88](https://github.com/terraform-providers/terraform-provider-azurestack/issues/88))
* `azurestack_virtual_machine` - support for manage disks ([#85](https://github.com/terraform-providers/terraform-provider-azurestack/issues/85))
* `azurestack_virtual_machine_scale_set` - add support for managed disks ([#93](https://github.com/terraform-providers/terraform-provider-azurestack/issues/93))

## 0.7.0 (May 23, 2019)

* dependencies: upgrading to `v29.0.0` of `github.com/Azure/azure-sdk-for-go` ([#83](https://github.com/terraform-providers/terraform-provider-azurestack/issues/83))
* dependencies: upgrading to `v11.7.0` of `github.com/Azure/go-autorest` ([#83](https://github.com/terraform-providers/terraform-provider-azurestack/issues/83))
* dependencies: upgrading to `v0.12.0` of `github.com/hashicorp/terraform` ([#86](https://github.com/terraform-providers/terraform-provider-azurestack/issues/86))

## 0.6.0 (April 19, 2019)

NOTES:

* This release includes a Terraform SDK upgrade with compatibility for Terraform v0.12. The provider remains backwards compatible with Terraform v0.11 and there should not be any significant behavioural changes. ([#75](https://github.com/terraform-providers/terraform-provider-azurestack/issues/75))

## 0.5.0 (April 05, 2019)

IMPROVEMENTS:

* dependencies: switching to Go Modules ([#70](https://github.com/terraform-providers/terraform-provider-azurestack/issues/70))
* dependencies: upgrading to Go 1.11 ([#53](https://github.com/terraform-providers/terraform-provider-azurestack/issues/53))
* dependencies: upgrading to version 21.3.0 of github.com/Azure/azure-sdk-for-go ([#53](https://github.com/terraform-providers/terraform-provider-azurestack/issues/53))
* dependencies: upgrading to terraform 0.11.13 ([#78](https://github.com/terraform-providers/terraform-provider-azurestack/issues/78))
* authentication: switching to use the new authentication package ([#54](https://github.com/terraform-providers/terraform-provider-azurestack/issues/54))
* authentication: support for Client Certificate authentication ([#56](https://github.com/terraform-providers/terraform-provider-azurestack/issues/56))
* authentication: support for CLI authentication ([#57](https://github.com/terraform-providers/terraform-provider-azurestack/issues/57))

BUG FIXES:

* `azurestack_virtual_network_gateway` - will no longer panic when the API/SDK return an empty `bgp_settings` property ([#71](https://github.com/terraform-providers/terraform-provider-azurestack/issues/71))

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
