package issuestate

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
)

// GetIssueStatesRelations 获取工作流
func (is *IssueState) GetIssueStatesRelations(req apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateRelation, error) {
	issueRelations, err := is.db.GetIssuesStateRelations(req.ProjectID, req.IssueType)
	if err != nil {
		return nil, err
	}
	var response []apistructs.IssueStateRelation
	var index int = -1
	for _, v := range issueRelations {
		// 按照开始状态归类
		if index == -1 || v.ID != response[index].StateID {
			index++
			response = append(response, apistructs.IssueStateRelation{
				IssueStatus: apistructs.IssueStatus{
					ProjectID:   v.ProjectID,
					IssueType:   v.IssueType,
					StateID:     v.ID,
					StateName:   v.Name,
					StateBelong: v.Belong,
					Index:       v.Index,
				},
			})
		}
		if v.EndStateID != 0 {
			response[index].StateRelation = append(response[index].StateRelation, v.EndStateID)
		}
	}
	return response, err
}

// UpdateIssueStates 更新工作流
func (is *IssueState) UpdateIssueStates(updateReq *apistructs.IssueStateUpdateRequest) (apistructs.IssueType, error) {
	for _, v := range updateReq.Data {
		sta := &dao.IssueState{
			BaseModel: dao.BaseModel{
				ID: v.StateID,
			},
			ProjectID: v.ProjectID,
			IssueType: v.IssueType,
			Name:      v.StateName,
			Belong:    v.StateBelong,
			Index:     v.Index,
			Role:      "Ops,Dev,QA,Owner,Lead",
		}
		if err := is.db.UpdateIssueState(sta); err != nil {
			return "", err
		}
	}
	var updateIssueStateRelations []dao.IssueStateRelation
	for _, st := range updateReq.Data {
		for _, endID := range st.StateRelation {
			if endID == st.StateID {
				continue
			}
			updateIssueStateRelations = append(updateIssueStateRelations, dao.IssueStateRelation{
				ProjectID:    updateReq.ProjectID,
				IssueType:    st.IssueType,
				StartStateID: st.StateID,
				EndStateID:   endID,
			})
		}
	}
	if len(updateIssueStateRelations) == 0 {
		return "", apierrors.ErrUpdateIssueState.MissingParameter("工作流不能为空")
	}
	return updateReq.Data[0].IssueType, is.db.UpdateIssueStateRelations(updateReq.ProjectID, updateReq.Data[0].IssueType, updateIssueStateRelations)
}
