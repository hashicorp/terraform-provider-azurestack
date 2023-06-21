// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/terraform-provider-azurestack/internal/locks"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type networkInterfaceIPConfigurationLockingDetails struct {
	subnetNamesToLock         []string
	virtualNetworkNamesToLock []string
}

func (details networkInterfaceIPConfigurationLockingDetails) lock() {
	locks.MultipleByName(&details.subnetNamesToLock, SubnetResourceName)
	locks.MultipleByName(&details.virtualNetworkNamesToLock, VirtualNetworkResourceName)
}

func (details networkInterfaceIPConfigurationLockingDetails) unlock() {
	locks.UnlockMultipleByName(&details.subnetNamesToLock, SubnetResourceName)
	locks.UnlockMultipleByName(&details.virtualNetworkNamesToLock, VirtualNetworkResourceName)
}

func determineResourcesToLockFromIPConfiguration(input *[]network.InterfaceIPConfiguration) (*networkInterfaceIPConfigurationLockingDetails, error) {
	if input == nil {
		return &networkInterfaceIPConfigurationLockingDetails{
			subnetNamesToLock:         []string{},
			virtualNetworkNamesToLock: []string{},
		}, nil
	}

	subnetNamesToLock := make([]string, 0)
	virtualNetworkNamesToLock := make([]string, 0)

	for _, config := range *input {
		if config.Subnet == nil || config.Subnet.ID == nil {
			continue
		}

		id, err := parse.SubnetID(*config.Subnet.ID)
		if err != nil {
			return nil, err
		}

		virtualNetworkName := id.VirtualNetworkName
		subnetName := id.Name

		if !utils.SliceContainsValue(virtualNetworkNamesToLock, virtualNetworkName) {
			virtualNetworkNamesToLock = append(virtualNetworkNamesToLock, virtualNetworkName)
		}

		if !utils.SliceContainsValue(subnetNamesToLock, subnetName) {
			subnetNamesToLock = append(subnetNamesToLock, subnetName)
		}
	}

	return &networkInterfaceIPConfigurationLockingDetails{
		subnetNamesToLock:         subnetNamesToLock,
		virtualNetworkNamesToLock: virtualNetworkNamesToLock,
	}, nil
}
