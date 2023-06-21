// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
)

// NormalizeSSHKey attempts to remove invalid formatting and line breaks that can be present in some cases
// when querying the Azure APIs
func NormalizeSSHKey(input string) (*string, error) {
	if input == "" {
		return nil, fmt.Errorf("empty string supplied")
	}

	output := input
	output = strings.ReplaceAll(output, "<<~EOT", "")
	output = strings.ReplaceAll(output, "EOT", "")
	output = strings.ReplaceAll(output, "\r", "")

	lines := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		lines = append(lines, strings.TrimSpace(line))
	}

	normalised := strings.Join(lines, "")

	return pointer.FromString(normalised), nil
}
