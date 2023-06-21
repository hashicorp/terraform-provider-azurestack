// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package compute

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=AvailabilitySet -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Compute/availabilitySets/set1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=DataDisk -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Compute/virtualMachines/machine1/dataDisks/disk1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=ManagedDisk -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Compute/disks/disk1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=VirtualMachine -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Compute/virtualMachines/machine1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=VirtualMachineExtension -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Compute/virtualMachines/machine1/extensions/extension1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=VirtualMachineScaleSet -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Compute/virtualMachineScaleSets/scaleSet1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=VirtualMachineScaleSetExtension -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Compute/virtualMachineScaleSets/scaleSet1/extensions/extension1
