// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acceptance

// Regions is a list of Azure Regions which can be used for test purposes
type Regions struct {
	// Primary is the Primary/Default Azure Region which should be used for testing
	Primary string

	// Secondary is the Secondary Azure Region which should be used for testing
	Secondary string

	// Ternary is the Ternary Azure Region which should be used for testing
	Ternary string
}
