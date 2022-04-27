package features

func Default() UserFeatures {
	return UserFeatures{
		// NOTE: ensure all nested objects are fully populated
		ResourceGroup: ResourceGroupFeatures{
			PreventDeletionIfContainsResources: false,
		},
		VirtualMachine: VirtualMachineFeatures{
			DeleteOSDiskOnDeletion:     true,
			GracefulShutdown:           false,
			SkipShutdownAndForceDelete: false,
		},
		VirtualMachineScaleSet: VirtualMachineScaleSetFeatures{
			ForceDelete:               false,
			RollInstancesWhenRequired: true,
			ScaleToZeroOnDelete:       true,
		},
	}
}
