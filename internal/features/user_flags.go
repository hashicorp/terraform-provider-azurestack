package features

type UserFeatures struct {
	ResourceGroup  ResourceGroupFeatures
	VirtualMachine VirtualMachineFeatures
}

type ResourceGroupFeatures struct {
	PreventDeletionIfContainsResources bool
}

type VirtualMachineFeatures struct {
	DeleteOSDiskOnDeletion     bool
	GracefulShutdown           bool
	SkipShutdownAndForceDelete bool
}
