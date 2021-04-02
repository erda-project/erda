package manual_review

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
)

// Member 成员操作封装
type ManualReview struct {
	db *dao.DBClient
}

// Option 定义 ManualReview 对象配置选项
type Option func(*ManualReview)

// New 新建 ManualReview 实例
func New(options ...Option) *ManualReview {
	mem := &ManualReview{}
	for _, op := range options {
		op(mem)
	}
	return mem
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(m *ManualReview) {
		m.db = db
	}
}

func (m *ManualReview) CreateReviewUser(param *apistructs.CreateReviewUser) error {
	return m.db.CreateReviewUser(param)
}

func (m *ManualReview) GetReviewByTaskId(param *apistructs.GetReviewByTaskIdIdRequest) (apistructs.GetReviewByTaskIdIdResponse, error) {
	return m.db.GetReviewByTaskId(param)
}
func (m *ManualReview) CreateOrUpdate(param *apistructs.CreateReviewRequest) error {
	return m.db.CreateReview(param)
}

func (m *ManualReview) GetReviewsByUserId(param *apistructs.GetReviewsByUserIdRequest) (int, []apistructs.GetReviewsByUserIdResponse, error) {
	tasks, err := m.db.GetTaskIDByOperator(param)
	if err != nil {
		return 0, nil, err
	}
	total, reviews, err := m.db.GetReviewsByUserId(param, tasks)
	var list []apistructs.GetReviewsByUserIdResponse

	for _, v := range reviews {
		list = append(list, apistructs.GetReviewsByUserIdResponse{
			Id:              v.ID,
			BuildId:         v.BuildId,
			BranchName:      v.BranchName,
			ProjectName:     v.ProjectName,
			ProjectId:       v.ProjectId,
			ApplicationId:   v.ApplicationId,
			ApplicationName: v.ApplicationName,
			Operator:        v.SponsorId,
			CommitMessage:   v.CommitMessage,
			CommitId:        v.CommitID,
			ApprovalStatus:  v.ApprovalStatus,
			ApprovalReason:  v.ApprovalReason,
			ApprovalContent: "pipeline",
		})
	}

	return total, list, nil
}
func (m *ManualReview) GetAuthorityByUserId(param *apistructs.GetAuthorityByUserIdRequest) (apistructs.GetAuthorityByUserIdResponse, error) {
	return m.db.GetAuthorityByUserId(param)
}

func (m *ManualReview) GetReviewsBySponsorId(param *apistructs.GetReviewsBySponsorIdRequest) (int, []apistructs.GetReviewsBySponsorIdResponse, error) {
	total, manualReviews, err := m.db.GetReviewsBySponsorId(param)
	if err != nil {
		return 0, nil, err
	}
	var reviews []int
	for _, v := range manualReviews {
		reviews = append(reviews, v.TaskId)
	}
	reviewusers, _ := m.db.GetOperatorByTaskID(reviews)

	operatorMap := make(map[int][]string)

	for _, v := range reviewusers {
		operatorMap[int(v.TaskId)] = append(operatorMap[int(v.TaskId)], v.Operator)
	}
	var list []apistructs.GetReviewsBySponsorIdResponse
	for _, v := range manualReviews {

		list = append(list, apistructs.GetReviewsBySponsorIdResponse{
			Id:              v.ID,
			BuildId:         v.BuildId,
			BranchName:      v.BranchName,
			ProjectName:     v.ProjectName,
			ProjectId:       v.ProjectId,
			ApplicationName: v.ApplicationName,
			ApplicationId:   v.ApplicationId,
			CommitMessage:   v.CommitMessage,
			CommitId:        v.CommitID,
			Approver:        operatorMap[v.TaskId],
			ApprovalReason:  v.ApprovalReason,
			ApprovalContent: "pipeline",
		})
	}
	return total, list, nil
}

func (m *ManualReview) UpdateApproval(param *apistructs.UpdateApproval) error {
	return m.db.UpdateApproval(param)
}
