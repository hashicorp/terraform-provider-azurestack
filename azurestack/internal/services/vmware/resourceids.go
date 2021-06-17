package vmware

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=PrivateCloud -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.AVS/privateClouds/privateCloud1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Cluster -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.AVS/privateClouds/privateCloud1/clusters/cluster1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=ExpressRouteAuthorization -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.AVS/privateClouds/privateCloud1/authorizations/authorization1
