/*
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

// specifies the default hour (UTC) at which tests should be triggered, if enabled
var defaultStartHour = 0

// specifies the default level of parallelism per-service-package
var defaultParallelism = 20

// specifies the default version of Terraform Core which should be used for testing
var defaultTerraformCoreVersion = "1.0.3"

// This represents a cron view of days of the week, Monday - Friday.
const val defaultDaysOfWeek = "2,3,4,5,6"

// Cron value for any day of month
const val defaultDaysOfMonth = "*"

var locations = mapOf(
        "stack" to LocationConfiguration("ppe5", "ppe5", "ppe5", false)
)

// specifies the list of Azure Environments where tests should be run nightly
var runNightly = mapOf(
        "stack" to true
)

// specifies a list of services which should be run with a custom test configuration
var serviceTestConfigurationOverrides = mapOf(
        // these tests all conflict with one another
        "authorization" to testConfiguration(parallelism = 1),
)
