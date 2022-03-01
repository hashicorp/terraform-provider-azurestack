package validate

import (
	"fmt"
)

func StorageAccountTags(v interface{}, _ string) (warnings []string, errors []error) {
	tagsMap := v.(map[string]interface{})

	if len(tagsMap) > 50 {
		errors = append(errors, fmt.Errorf("a maximum of 50 tags can be applied to storage account ARM resource"))
	}

	for k, v := range tagsMap {
		if len(k) > 128 {
			errors = append(errors, fmt.Errorf("the maximum length for a tag key is 128 characters: %q is %d characters", k, len(k)))
		}

		value, err := TagValueToString(v)
		if err != nil {
			errors = append(errors, err)
		} else if len(value) > 256 {
			errors = append(errors, fmt.Errorf("the maximum length for a tag value is 256 characters: the value for %q is %d characters", k, len(value)))
		}
	}

	return warnings, errors
}

func TagValueToString(v interface{}) (string, error) {
	switch value := v.(type) {
	case string:
		return value, nil
	case int:
		return fmt.Sprintf("%d", value), nil
	default:
		return "", fmt.Errorf("unknown tag type %T in tag value", value)
	}
}
