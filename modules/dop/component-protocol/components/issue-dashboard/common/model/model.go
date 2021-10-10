package model

import "github.com/erda-project/erda/modules/dop/dao"

type LabelIssueItem struct {
	LabelRel *dao.IssueLabel
	Bug      *dao.IssueItem
}
