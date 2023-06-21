// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package compute

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func expandIDsToSubResources(input []interface{}) *[]compute.SubResource {
	ids := make([]compute.SubResource, 0)

	for _, v := range input {
		ids = append(ids, compute.SubResource{
			ID: utils.String(v.(string)),
		})
	}

	return &ids
}

func flattenSubResourcesToIDs(input *[]compute.SubResource) []interface{} {
	ids := make([]interface{}, 0)
	if input == nil {
		return ids
	}

	for _, v := range *input {
		if v.ID == nil {
			continue
		}

		ids = append(ids, *v.ID)
	}

	return ids
}
