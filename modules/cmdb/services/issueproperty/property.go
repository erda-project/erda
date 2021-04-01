package issueproperty

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
)

// IssueProperty issue property 对象
type IssueProperty struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 IssueProperty 对象配置选项
type Option func(*IssueProperty)

// New 新建 issue stream 对象
func New(options ...Option) *IssueProperty {
	is := &IssueProperty{}
	for _, op := range options {
		op(is)
	}
	return is
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(is *IssueProperty) {
		is.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(is *IssueProperty) {
		is.bdl = bdl
	}
}
