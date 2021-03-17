// Package errorresp 封装了错误定义
package errorresp

import (
	"github.com/erda-project/erda/pkg/i18n"
	"github.com/erda-project/erda/pkg/strutil"
)

type MetaMessage struct {
	Key     string
	Args    []interface{}
	Default string
}

// APIError 接口错误
type APIError struct {
	httpCode           int
	code               string
	msg                string
	localeMetaMessages []MetaMessage
}

// Error 错误信息
func (e *APIError) Error() string {
	if e.msg == "" {
		e.Render(&i18n.LocaleResource{})
	}
	return e.msg
}

// Code 错误码
func (e *APIError) Code() string {
	return e.code
}

// HttpCode HTTP错误码
func (e *APIError) HttpCode() int {
	return e.httpCode
}

// Option 可选参数
type Option func(*APIError)

// New 新建接口错误
func New(options ...Option) *APIError {
	e := &APIError{}
	for _, op := range options {
		op(e)
	}
	return e
}

// WithMessage 初始化 msg
func WithMessage(msg string) Option {
	return func(a *APIError) {
		a.msg = msg
	}
}

// WithTemplateMessage 初始化 msg
func WithTemplateMessage(template, defaultValue string, args ...interface{}) Option {
	return func(a *APIError) {
		_ = a.appendMeta(template, defaultValue, args...)
	}
}

func WithCode(httpCode int, code string) Option {
	return func(a *APIError) {
		_ = a.appendCode(httpCode, code)
	}
}

func (e *APIError) appendCode(httpCode int, code string) *APIError {
	e.httpCode = httpCode
	e.code = code
	return e
}

func (e *APIError) appendMsg(template *i18n.Template, args ...interface{}) *APIError {
	msg := template.Render(args...)
	if e.msg == "" {
		e.msg = msg
		return e
	}
	e.msg = strutil.Concat(e.msg, ": ", msg)
	return e
}

func (e *APIError) appendMeta(key string, defaultContent string, args ...interface{}) *APIError {
	e.localeMetaMessages = append(e.localeMetaMessages, MetaMessage{
		Key:     key,
		Args:    args,
		Default: defaultContent,
	})
	return e
}

func (e *APIError) appendLocaleTemplate(template *i18n.Template, args ...interface{}) *APIError {
	e.localeMetaMessages = append(e.localeMetaMessages, MetaMessage{
		Key:     template.Key(),
		Args:    args,
		Default: template.Content(),
	})
	return e
}

func (e *APIError) Render(localeResource *i18n.LocaleResource) string {
	for _, metaMessage := range e.localeMetaMessages {
		var template *i18n.Template
		// 不存在key使用默认值
		if !localeResource.ExistKey(metaMessage.Key) && metaMessage.Default != "" {
			template = i18n.NewTemplate(metaMessage.Key, metaMessage.Default)
		} else {
			template = localeResource.GetTemplate(metaMessage.Key)
		}
		msg := template.Render(metaMessage.Args...)
		if e.msg == "" {
			e.msg = msg
		} else {
			e.msg = strutil.Concat(e.msg, ": ", msg)
		}
	}
	return e.msg
}

func (e *APIError) dup() *APIError {
	return &APIError{
		httpCode:           e.httpCode,
		code:               e.code,
		msg:                e.msg,
		localeMetaMessages: e.localeMetaMessages,
	}
}
