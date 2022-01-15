package azurestack

// handle the case of using the same name for different kinds of resources
func azureStackLockByName(name string, resourceType string) {
	updatedName := resourceType + "." + name
	locks.ByID(updatedName)
}

func azureStackLockMultipleByName(names *[]string, resourceType string) {
	for _, name := range *names {
		azureStackLockByName(name, resourceType)
	}
}

func azureStackUnlockByName(name string, resourceType string) {
	updatedName := resourceType + "." + name
	locks.UnlockByID(updatedName)
}

func azureStackUnlockMultipleByName(names *[]string, resourceType string) {
	for _, name := range *names {
		azureStackUnlockByName(name, resourceType)
	}
}
