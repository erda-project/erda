package model

import "github.com/erda-project/erda/apistructs"

type BranchRule struct {
	BaseModel
	ScopeType         apistructs.ScopeType
	ScopeID           int64
	Rule              string
	IsProtect         bool
	IsTriggerPipeline bool
	NeedApproval      bool
	Desc              string //规则说明
	Workspace         string `json:"workspace"`
	ArtifactWorkspace string `json:"artifactWorkspace"`
}

// TableName 设置模型对应数据库表名称
func (BranchRule) TableName() string {
	return "dice_branch_rules"
}

func (rule *BranchRule) ToApiData() *apistructs.BranchRule {
	return &apistructs.BranchRule{
		ID:                rule.ID,
		Rule:              rule.Rule,
		ScopeID:           rule.ScopeID,
		ScopeType:         rule.ScopeType,
		IsProtect:         rule.IsProtect,
		NeedApproval:      rule.NeedApproval,
		IsTriggerPipeline: rule.IsTriggerPipeline,
		Desc:              rule.Desc,
		Workspace:         rule.Workspace,
		ArtifactWorkspace: rule.ArtifactWorkspace,
	}
}
