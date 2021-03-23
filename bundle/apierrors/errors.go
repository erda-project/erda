package apierrors

import "github.com/erda-project/erda/pkg/httpserver/errorresp"

var (
	ErrInvoke            = err("ErrInvoke", "调用失败")
	ErrUnavailableClient = err("ErrUnavailableClient", "客户端不可用")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
