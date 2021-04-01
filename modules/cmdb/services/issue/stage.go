package issue

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
)

func (svc *Issue) CreateIssueStage(req *apistructs.IssueStageRequest) error {
	err := svc.db.DeleteIssuesStage(req.OrgID, req.IssueType)
	if err != nil {
		return err
	}
	var stages []dao.IssueStage
	for _, v := range req.List {
		stages = append(stages, dao.IssueStage{
			OrgID:     req.OrgID,
			IssueType: req.IssueType,
			Name:      v.Name,
			Value:     v.Value,
		})
	}
	return svc.db.CreateIssueStage(stages)
}

func (svc *Issue) GetIssueStage(req *apistructs.IssueStageRequest) ([]apistructs.IssueStage, error) {
	stages, err := svc.db.GetIssuesStage(req.OrgID, req.IssueType)
	if err != nil {
		return nil, err
	}
	var res []apistructs.IssueStage
	for _, v := range stages {
		res = append(res, apistructs.IssueStage{
			ID:    v.ID,
			Name:  v.Name,
			Value: v.Value,
		})
	}
	return res, nil
}
