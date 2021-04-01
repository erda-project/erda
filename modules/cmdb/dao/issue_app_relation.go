package dao

// IssueAppRelation 事件应用关联
type IssueAppRelation struct {
	BaseModel
	IssueID   int64
	CommentID int64
	AppID     int64
	MRID      int64
}

// TableName 表名
func (IssueAppRelation) TableName() string {
	return "dice_issue_app_relations"
}

// CreateIssueAppRelation 创建事件应用关联关系
func (client *DBClient) CreateIssueAppRelation(issueAppRel *IssueAppRelation) error {
	return client.Create(issueAppRel).Error
}

// DeleteIssueAppRelationsByComment 根据 commentID 删除关联关系
func (client *DBClient) DeleteIssueAppRelationsByComment(commentID int64) error {
	return client.Where("comment_id = ?", commentID).Delete(&IssueAppRelation{}).Error
}

// DeleteIssueAppRelationsByApp 根据 appID 删除关联关系
func (client *DBClient) DeleteIssueAppRelationsByApp(appID int64) error {
	return client.Where("app_id = ?", appID).Delete(&IssueAppRelation{}).Error
}
