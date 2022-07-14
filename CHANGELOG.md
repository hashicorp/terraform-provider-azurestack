

## 1.0.0 (Unreleased)

NOTES:

* **Major Version**: Version `1.0` of the Azure Stack Provider is a major version - some behaviours have changed and some deprecated fields/resources have been removed - please refer to [the 1.0 upgrade guide for more information](https://registry.terraform.io/providers/hashicorp/azurestack/latest/docs/guides/1.0-upgrade-guide).
* **Provider Block:** The Azure Stack Provider now requires that a `features` block is specified within the Provider block, which can be used to alter the behaviour of certain resources - [more information on the `features` block can be found in the documentation](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#features).
* **Terraform 0.10/0.11:** Version `1.0` of the Azure Stack Provider no longer supports Terraform `0.10` or `0.11` - you must upgrade to Terraform `0.12` to use version `1.0` of the Azure Stack Provider.
* When upgrading to `v1.0` of the AzureStack Provider, we recommend upgrading to the latest version of Terraform Core ([which can be found here](https://www.terraform.io/downloads)) - the next major release of the AzureRM Provider (`v2.0`) will require Terraform `1.0` or later.

FEATURES:

* **Custom Timeouts:** - all resources within the Azure Stack Provider now allow configuring custom timeouts - please [see Terraform's Timeout documentation](https://www.terraform.io/docs/configuration/resources.html#operation-timeouts) and the documentation in each data source resource for more information.
* **Requires Import:** The Azure Stack Provider now checks for the presence of an existing resource prior to creating it - which means that if you try and create a resource which already exists (without importing it) you'll be prompted to import this into the state.
* **Import:** The Azure Stack Provider now checks import IDs for the correct resource ID formatting and reports what segments are either missing or incorrect.
* **New Service**: `keyvault` [GH-151]
* **New Data Source**: `azurestack_dns_zone` [GH-149]
* **New Data Source**: `azurestack_image` [GH-143]
* **New Data Source**: `azurestack_key_vault` [GH-151]
* **New Data Source**: `azurestack_key_vault_key` [GH-151]
* **New Data Source**: `azurestack_key_vault_secret` [GH-151]
* **New Data Source**: `azurestack_key_vault_access_policy` [GH-151]
* **New Data Source**: `azurestack_resources` [GH-170]
* **New Data Source**: `azurestack_storage_container` [GH-157]
* **New Resource**: `azurestack_dns_aaaa_record` [GH-149]
* **New Resource**: `azurestack_dns_cname_record` [GH-149]
* **New Resource**: `azurestack_dns_mx_record` [GH-149]
* **New Resource**: `azurestack_dns_ns_record` [GH-149]
* **New Resource**: `azurestack_dns_ptr_record` [GH-149]
* **New Resource**: `azurestack_dns_srv_record` [GH-149]
* **New Resource**: `azurestack_dns_txt_record` [GH-149]
* **New Resource**: `azurestack_dns_zone` [GH-149]
* **New Resource**: `azurestack_image` [GH-143]
* **New Resource**: `azurestack_key_vault` [GH-151]
* **New Resource**: `azurestack_key_vault_key` [GH-151]
* **New Resource**: `azurestack_key_vault_secret` [GH-151]
* **New Resource**: `azurestack_key_vault_access_policy` [GH-151]
* **New Resource**: `azurestack_linux_virtual_machine_scale_set` [GH-182]
* **New Resource**: `azurestack_linux_virtual_machine` [GH-148]
* **New Resource**: `azurestack_network_interface_backend_address_pool_association` [GH-155]
* **New Resource**: `azurestack_virtual_network_peering` [GH-147]
* **New Resource**: `azurestack_windows_virtual_machine_scale_set` [GH-182]
* **New Resource**: `azurestack_windows_virtual_machine` [GH-148]


BREAKING CHANGES:

* all `tags` properties are no longer `computed`
* Data Source: `azurestack_subnet` - the `ip_configurations` attribute has been removed [GH-167]
* `azurestack_network_interface` - the `load_balancer_backend_address_pools_ids`, `load_balancer_inbound_nat_rules_ids`, `internal_fqdn`,  and `internal_dns_name_label` properties have been removed [GH-164]
* `azurestack_storage_blob` - the `resource_group_name` property has been removed [GH-163]
* `azurestack_storage_container` - the `resource_group_name` property has been removed [GH-157Z]
* `azurestack_virtual_network_gateway_connection` - the `type` property no longer supports `Vnet2Vnet` [GH-173]

ENHANCEMENTS:

* dependencies: `azure-sdk-for-go` profile has been upgraded to `v59.2.0` 
* dependencies: the `azure-sdk-for-go` profile has been updated to `2020-09-01`
* provider: added a new feature flag within the `resource_group` block for `prevent_deletion_if_contains_resources`, for configuring whether Terraform should prevent the deletion of a Resource Group which still contains items
* provider: added a new feature flag `force_delete` within the `virtual_machine_scale_set` block to configure whether the VMSS should be force deleted during deletion [GH-182]
* provider: added a new feature flag `roll_instances_when_required` within the `virtual_machine_scale_set` block to configure whether the VMSS should be rolled when required [GH-182]
* provider: added a new feature flag `scale_to_zero_before_deletion` within the `virtual_machine_scale_set` block to configure whether the VMSS should be scaled to zero during deletion [GH-182]

* Data Source: `azurestack_storage_account` - support for the `enable_https_traffic_only` attributes [GH-169]
* `azurestack_loadbalancer` - support for the `sku` property [GH-152]
* `azurestack_managed_disk` - support for the `encryption` block and `hyper_v_generation` property [GH-175]
* `azurestack_resource_group` - Terraform now checks during the deletion of a Resource Group if there's any items remaining and will raise an error if so by default (to avoid deleting items unintentionally). This behaviour can be controlled using the `prevent_deletion_if_contains_resources` feature-flag within the `resource_group` block within the `features` block
* `azurestack_storage_account` - support for the `enable_https_traffic_only` property [GH-169]
* `azurestack_storage_blob` - support for the `has_immutability_policy`, `content_type`, `source_content`, `content_md5`, and `metadata` properties [GH-163]
* `azurestack_storage_container` - now exports the `cache_control` and `has_legal_hold` attributes [GH-157]
* `azurestack_storage_container` - the `properties` property has been renamed to `metadata` [GH-157]
* `azurestack_virtual_network_gateway_connection` - the `ike_encryption` property now supports `GCMAES128` and `GCMAES256` [GH-173]
* `azurestack_virtual_network_gateway_connection` - the `pfs_group` property now supports `PFS14` and `PFSMM` [GH-173]

---

For information on changes prior to the v1.0.0 release, please see [the v0.x changelog](https://github.com/hashicorp/terraform-provider-azurestack/blob/main/CHANGELOG-v0.md).
