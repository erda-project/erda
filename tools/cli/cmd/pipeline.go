// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/utils"
)

var PIPELINE = command.Command{
	Name:      "pipeline",
	ShortHelp: "Pipeline operations",
	Example:   `erda-cli pipeline logs -i <pipelineID>`,
}

var (
	getWorkspaceBranch          = utils.GetWorkspaceBranch
	getWorkspaceInfo            = utils.GetWorkspaceInfo
	getOrgDetail                = common.GetOrgDetail
	resolveWorkspaceApplication = common.ResolveWorkspaceApplication
	getLatestPipelineID         = common.GetLatestPipelineID
	listPipelinesCICD           = common.ListPipelinesCICD
)
