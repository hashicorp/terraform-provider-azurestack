// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utils

func AllEquals(v ...interface{}) bool {
	if len(v) > 1 {
		a := v[0]
		for _, s := range v {
			if a != s {
				return false
			}
		}
	}
	return true
}
