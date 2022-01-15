package features

type UserFeatures struct {
	ResourceGroup ResourceGroupFeatures
}

type ResourceGroupFeatures struct {
	PreventDeletionIfContainsResources bool
}
