/*
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import AzureStack
import ClientConfiguration
import jetbrains.buildServer.configs.kotlin.*

version = "2023.05"

var clientId = DslContext.getParameter("clientId", "")
var clientSecret = DslContext.getParameter("clientSecret", "")
var subscriptionId = DslContext.getParameter("subscriptionId", "")
var tenantId = DslContext.getParameter("tenantId", "")
var endpoint = DslContext.getParameter("endpoint", "")

var clientConfig = ClientConfiguration(clientId, clientSecret, subscriptionId, tenantId, endpoint)

project(AzureStack("stack", clientConfig))
