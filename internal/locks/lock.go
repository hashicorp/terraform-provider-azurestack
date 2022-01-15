package locks

// armMutexKV is the instance of MutexKV for ARM resources
var armMutexKV = NewMutexKV()

func ByID(id string) {
	locks.ByID(id)
}

// handle the case of using the same name for different kinds of resources
func ByName(name string, resourceType string) {
	updatedName := resourceType + "." + name
	locks.ByID(updatedName)
}

func MultipleByName(names *[]string, resourceType string) {
	newSlice := removeDuplicatesFromStringArray(*names)

	for _, name := range newSlice {
		ByName(name, resourceType)
	}
}

func UnlockByID(id string) {
	locks.UnlockByID(id)
}

func UnlockByName(name string, resourceType string) {
	updatedName := resourceType + "." + name
	locks.UnlockByID(updatedName)
}

func UnlockMultipleByName(names *[]string, resourceType string) {
	newSlice := removeDuplicatesFromStringArray(*names)

	for _, name := range newSlice {
		UnlockByName(name, resourceType)
	}
}
