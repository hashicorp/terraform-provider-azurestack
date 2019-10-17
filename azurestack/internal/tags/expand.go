package tags

// TODO: change function signature to match azurerm
// func Expand(tagsMap map[string]interface{}) map[string]*string

func Expand(tagsMap map[string]interface{}) *map[string]*string {
	output := make(map[string]*string, len(tagsMap))

	for i, v := range tagsMap {
		//Validate should have ignored this error already
		value, _ := TagValueToString(v)
		output[i] = &value
	}

	return &output
}
