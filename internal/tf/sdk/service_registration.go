// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdk

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

// TypedServiceRegistration is a Service Registration using Types
// meaning that we can abstract on top of the Plugin SDK and use
// Native Types where possible
type TypedServiceRegistration interface {
	// Name is the name of this Service
	Name() string

	// DataSources returns a list of Data Sources supported by this Service
	DataSources() []DataSource

	// Resources returns a list of Resources supported by this Service
	Resources() []Resource

	// WebsiteCategories returns a list of categories which can be used for the sidebar
	WebsiteCategories() []string
}

// UntypedServiceRegistration is the interface used for untyped/raw Plugin SDK resources
// in the future this'll be superseded by the TypedServiceRegistration which allows for
// stronger Typed resources to be used.
type UntypedServiceRegistration interface {
	// Name is the name of this Service
	Name() string

	// WebsiteCategories returns a list of categories which can be used for the sidebar
	WebsiteCategories() []string

	// SupportedDataSources returns the supported Data Sources supported by this Service
	SupportedDataSources() map[string]*pluginsdk.Resource

	// SupportedResources returns the supported Resources supported by this Service
	SupportedResources() map[string]*pluginsdk.Resource
}
