/*
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import jetbrains.buildServer.configs.kotlin.v2019_2.vcs.GitVcsRoot

object providerRepository : GitVcsRoot({
    name = "terraform-provider-azurestack"
    url = "https://github.com/hashicorp/terraform-provider-azurestack.git"
    agentCleanPolicy = AgentCleanPolicy.ALWAYS
    agentCleanFilesPolicy = AgentCleanFilesPolicy.ALL_UNTRACKED
    branchSpec = "+:*"
    branch = "refs/heads/main"
    authMethod = anonymous()
})
