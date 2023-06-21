/*
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import jetbrains.buildServer.configs.kotlin.v2019_2.ParametrizedWithType

class ClientConfiguration(var clientId: String,
                          var clientSecret: String,
                          val subscriptionId : String,
                          val tenantId : String,
                          val endpoint : String) {
}

class LocationConfiguration(var primary : String, var secondary : String, var ternary : String, var rotate : Boolean) {
}

fun ParametrizedWithType.ConfigureAzureSpecificTestParameters(config: ClientConfiguration, locationsForEnv: LocationConfiguration) {
    hiddenPasswordVariable("env.ARM_CLIENT_ID", config.clientId, "The ID of the Service Principal used for Testing")
    hiddenPasswordVariable("env.ARM_CLIENT_SECRET", config.clientSecret, "The Client Secret of the Service Principal used for Testing")
    hiddenVariable("env.ARM_METADATA_HOST", config.endpoint, "The Azure Stack Endpoint to use")
    hiddenPasswordVariable("env.ARM_SUBSCRIPTION_ID", config.subscriptionId, "The ID of the Azure Subscription used for Testing")
    hiddenPasswordVariable("env.ARM_TENANT_ID", config.tenantId, "The ID of the Azure Tenant used for Testing")
    hiddenVariable("env.ARM_TEST_LOCATION", locationsForEnv.primary, "The Primary region which should be used for testing")
    hiddenVariable("env.ARM_TEST_LOCATION_ALT", locationsForEnv.secondary, "The Primary region which should be used for testing")
    hiddenVariable("env.ARM_TEST_LOCATION_ALT2", locationsForEnv.ternary, "The Primary region which should be used for testing")
    hiddenVariable("env.ARM_THREEPOINTZERO_BETA_RESOURCES", "true", "Opt into the use of 3.0 beta resources")
}