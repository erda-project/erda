package model

import "github.com/erda-project/erda-infra/providers/legacy/httpendpoints/errorresp"

var (
	ErrCreateNotify = err("ErrCreateNotify", "创建通知失败")
	ErrGetNotify    = err("ErrGetNotify", "获取通知失败")
	ErrDeleteNotify = err("ErrDeleteNotify", "删除通知失败")
	ErrUpdateNotify = err("ErrUpdateNotify", "更新通知失败")
	ErrNotifyEnable = err("ErrNotifyEnable", "启用通知失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
