package features

type UserFeatures struct {
	ResourceGroup          ResourceGroupFeatures
	VirtualMachine         VirtualMachineFeatures
	VirtualMachineScaleSet VirtualMachineScaleSetFeatures
}

type ResourceGroupFeatures struct {
	PreventDeletionIfContainsResources bool
}

type VirtualMachineFeatures struct {
	DeleteOSDiskOnDeletion     bool
	GracefulShutdown           bool
	SkipShutdownAndForceDelete bool
}

type VirtualMachineScaleSetFeatures struct {
	ForceDelete               bool
	RollInstancesWhenRequired bool
	ScaleToZeroOnDelete       bool
}
