// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apistructs

const (
	LabelOrgID   = "orgID"
	LabelOrgName = "orgName"

	LabelProjectID   = "projectID"
	LabelProjectName = "projectName"

	LabelTestPlanID = "testPlanID"

	LabelAppID   = "appID"
	LabelAppName = "appName"

	LabelDiceWorkspace = "diceWorkspace"

	LabelBranch       = "branch"
	LabelCommit       = "commit"
	LabelCommitDetail = "commitDetail"

	LabelPipelineYmlSource       = "pipelineYmlSource"
	LabelPipelineType            = "pipelineType"
	LabelPipelineTriggerMode     = "pipelineTriggerMode"
	LabelPipelineCronTriggerTime = "pipelineCronTriggerTime"
	LabelPipelineCronID          = "pipelineCronID"
	LabelPipelineCronCompensated = "cronCompensated"

	LabelBindPipelineQueueID             = "__bind_queue_id"
	LabelBindPipelineQueueInsidePriority = "__bind_queue_inside_priority"

	LabelUserID = "userID"

	// ---------------------- snippet some global labels
	// action
	LabelActionVersion = "actionVersion"
	LabelActionJson    = "actionJson"

	// dice
	LabelDiceSnippetScopeID   = "scopeID"
	LabelChooseSnippetVersion = "chooseVersion"
	// snippet
	LabelSnippetScope     = "snippet_scope"
	LabelActionEnv        = "action_env"
	DiceApplicationId     = "DICE_APPLICATION_ID"
	DiceApplicationName   = "DICE_APPLICATION_NAME"
	DiceWorkspaceEnv      = "DICE_WORKSPACE"
	GittarBranchEnv       = "GITTAR_BRANCH"
	LabelGittarYmlPath    = "gittarYmlPath"    // app snippetConfig label in order to specify the address of yml to address
	LabelAutotestExecType = "autotestExecType" // 新版自动化测试的snippet的执行类型
	LabelSceneSetID       = "sceneSetID"       // 新版自动化测试的场景集的 id
	LabelSceneID          = "sceneID"          // 新版自动化测试的场景的 id
	LabelSpaceID          = "spaceID"          // 空间 id
	// FDP
	LabelFdpWorkflowID          = "CDP_WF_ID"
	LabelFdpWorkflowName        = "CDP_WF_NAME"
	LabelFdpWorkflowProcessType = "CDP_WF_PROCESS_TYPE"
	LabelFdpWorkflowRuntype     = "CDP_WF_RUNTYPE"
)
