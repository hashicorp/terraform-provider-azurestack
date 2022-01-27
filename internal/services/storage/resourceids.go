package storage

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=StorageAccount -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Storage/storageAccounts/storageAccount1 -rewrite=true
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=StorageContainerResourceManager -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Storage/storageAccounts/storageAccount1/blobServices/default/containers/container1
