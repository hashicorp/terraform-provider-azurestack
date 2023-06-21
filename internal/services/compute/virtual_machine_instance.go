// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package compute

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
)

// virtualMachineShouldBeStarted determines if the Virtual Machine should be started after
// the Virtual Machine has been shut down for maintenance. This means that Virtual Machines
// which are already stopped can be updated but will not be started
func virtualMachineShouldBeStarted(instanceView compute.VirtualMachineInstanceView) bool {
	if instanceView.Statuses != nil {
		for _, status := range *instanceView.Statuses {
			if status.Code == nil {
				continue
			}

			// could also be the provisioning state which we're not bothered with here
			state := strings.ToLower(*status.Code)
			if !strings.HasPrefix(state, "powerstate/") {
				continue
			}

			state = strings.TrimPrefix(state, "powerstate/")
			if strings.EqualFold(state, "running") {
				return true
			}
		}
	}

	return false
}
