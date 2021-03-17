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
