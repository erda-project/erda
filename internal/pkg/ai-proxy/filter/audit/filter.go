// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package audit

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Name = "audit"
)

var (
	_ filter.RequestGetterFilter  = (*Audit)(nil)
	_ filter.ResponseGetterFilter = (*Audit)(nil)
)

func init() {
	filter.Register(Name, New)
}

type Audit struct {
	audit *AiAudit
}

func New(_ json.RawMessage) (filter.Filter, error) {
	return &Audit{audit: new(AiAudit)}, nil
}

func (f *Audit) OnHttpRequestGetter(ctx context.Context, g filter.HttpInfor) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("Audit").Sub("OnHttpRequestGetter")
	for _, set := range []func(context.Context, http.Header, *bytes.Buffer) error{
		f.audit.SetSessionId,
		f.audit.SetChats,
		f.audit.SetRequestAt,
		f.audit.SetSource,
		f.audit.SetUserInfo,
		f.audit.SetOperationId,
		f.audit.SetRequestContentType,
		f.audit.SetUserAgent,
		f.audit.SetRequestBody,
		f.audit.SetProvider,
		f.audit.SetModel,
		f.audit.SetPrompt,
	} {
		buf, err := g.Body()
		if err != nil {
			l.Errorf("failed to filter.HttpInfor.Body, err: %v", err)
			continue
		}
		// todo: r.Clone every time is less efficient
		if err := set(ctx, g.Header(), buf); err != nil {
			l.Errorf("failed to %v, err: %v", reflect.TypeOf(set), err)
			continue
		}
	}
	return filter.Continue, nil
}

func (f *Audit) OnHttpResponseGetter(ctx context.Context, getter filter.HttpInfor) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("Audit").Sub("OnHttpResponse")
	for _, set := range []func(context.Context, http.Header, *bytes.Buffer) error{
		f.audit.SetResponseAt,
		f.audit.SetCompletion,
		f.audit.SetResponseContentType,
		f.audit.SetResponseBody,
		f.audit.SetServer,
	} {
		buf, err := getter.Body()
		if err != nil {
			return filter.Intercept, errors.Wrap(err, "failed to filter.HeadBodyGetter.Body()")
		}
		if err := set(ctx, getter.Header(), buf); err != nil {
			l.Errorf("failed to do %T, err: %v", set, err)
		}
	}
	for _, set := range []func(context.Context, filter.HttpInfor) error{
		f.audit.SetStatus,
	} {
		if err := set(ctx, getter); err != nil {
			l.Errorf("failed to do %T, err: %v", set, err)
		}
	}
	return filter.Continue, f.create(ctx)
}

func (f *Audit) create(ctx context.Context) error {
	db, ok := ctx.Value(filter.DBCtxKey{}).(*gorm.DB)
	if !ok {
		panic("no *gorm.DB set")
	}
	return db.Create(f.audit).Error
}

// AiAudit is the table ai_audit
type AiAudit struct {
	Id        fields.UUID      `json:"id" yaml:"id" gorm:"id"`
	CreatedAt time.Time        `json:"createdAt" yaml:"createdAt" gorm:"created_at"`
	UpdatedAt time.Time        `json:"updatedAt" yaml:"updatedAt" gorm:"updated_at"`
	DeletedAt fields.DeletedAt `json:"deletedAt" yaml:"deletedAt" gorm:"deleted_at"`

	// SessionId records the uniqueness of the conversation
	SessionId string `json:"sessionId" yaml:"sessionId" gorm:"session_id"`
	ChatType  string `json:"chatType" yaml:"chatType" gorm:"chat_type"`
	ChatTitle string `json:"chatTitle" yaml:"chatTitle" gorm:"chat_title"`
	ChatId    string `json:"chatId" yaml:"chatId" gorm:"chat_id"`
	// Source is the application source, like dingtalk, webui, vscode-plugin, jetbrains-plugin
	Source string `json:"source" yaml:"source" gorm:"source"`
	// UserInfo is a unique user identifier
	UserInfo string `json:"UserInfo" yaml:"UserInfo" gorm:"user_info"`
	// Provider is an AI capability provider, like openai:chatgpt/v1, baidu:wenxin, alibaba:tongyi
	Provider string `json:"provider" yaml:"provider" gorm:"provider"`
	// Model used for this request, e.g. gpt-3.5-turbo, gpt-4-8k
	Model string `json:"model" yaml:"model" gorm:"model"`
	// OperationId is the unique identifier of the API
	OperationId string `json:"operationId" yaml:"operationId" gorm:"operation_id"`
	// Prompt The prompt(s) to generate completions for, encoded as a string, array of strings, array of tokens, or array of token arrays.
	//
	// Note that <|endoftext|> is the document separator that the model sees during training,
	// so if a prompt is not specified the model will generate as if from the beginning of a new document.
	Prompt string `json:"prompt" yaml:"prompt" gorm:"prompt"`
	// Completion returns the response to the client
	Completion string `json:"completion" yaml:"completion" gorm:"completion"`

	// RequestAt is the request arrival time
	RequestAt time.Time `json:"requestAt" yaml:"requestAt" gorm:"request_at"`
	// ResponseAt is the response arrival time
	ResponseAt time.Time `json:"responseAt" yaml:"responseAt" gorm:"response_at"`
	// UserAgent http client's User-Agent
	UserAgent           string `json:"userAgent" yaml:"userAgent" gorm:"user_agent"`
	RequestContentType  string `json:"requestContentType" yaml:"requestContentType" gorm:"request_content_type"`
	RequestBody         string `json:"requestBody" yaml:"requestBody" gorm:"request_body"`
	ResponseContentType string `json:"responseContentType" yaml:"responseContentType" gorm:"response_content_type"`
	ResponseBody        string `json:"responseBody" yaml:"responseBody" gorm:"response_body"`
	Server              string `json:"server" yaml:"server" gorm:"server"`
	Status              string `json:"status" yaml:"status" gorm:"status"`
	StatusCode          int    `json:"statusCode" yaml:"statusCode" gorm:"status_code"`
}

func (a *AiAudit) SetSessionId(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	a.SessionId = header.Get("X-Erda-AI-Proxy-SessionId") // todo: Temporary
	return nil
}

func (a *AiAudit) SetChats(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	a.ChatType = header.Get("X-Erda-AI-Proxy-ChatType")
	a.ChatTitle = header.Get("X-Erda-AI-Proxy-ChatTitle")
	a.ChatId = header.Get("X-Erda-AI-Proxy-ChatId")
	for _, v := range []*string{
		&a.ChatType,
		&a.ChatTitle,
		&a.ChatId,
	} {
		if decoded, err := base64.StdEncoding.DecodeString(*v); err == nil {
			*v = string(decoded)
		}
	}
	return nil
}

func (a *AiAudit) SetRequestAt(_ context.Context, _ http.Header, _ *bytes.Buffer) error {
	a.RequestAt = time.Now()
	return nil
}

func (a *AiAudit) SetResponseAt(_ context.Context, _ http.Header, _ *bytes.Buffer) error {
	a.ResponseAt = time.Now()
	return nil
}

func (a *AiAudit) SetSource(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	a.Source = header.Get("X-Erda-AI-Proxy-Source")
	return nil
}

func (a *AiAudit) SetUserInfo(ctx context.Context, header http.Header, _ *bytes.Buffer) error {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit")
	var m = map[string]string{
		"dingTalkStaffId": header.Get("X-Erda-AI-Proxy-DingTalkStaffID"),
		"jobNumber":       header.Get("X-Erda-AI-Proxy-JobNumber"),
		"phone":           header.Get("X-Erda-AI-Proxy-Phone"),
		"email":           header.Get("X-Erda-AI-Proxy-Email"),
		"name":            header.Get("X-Erda-AI-Proxy-Name"),
	}
	for k, v := range m {
		if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
			m[k] = string(decoded)
		}
	}
	data, err := json.Marshal(m)
	if err != nil {
		l.Errorf("failed to json.Marshal(%+v), err: %v", m, err)
		return err
	}
	a.UserInfo = string(data)
	return nil
}

func (a *AiAudit) SetProvider(ctx context.Context, _ http.Header, _ *bytes.Buffer) error {
	// a.Provider is passed in by filter reverse-proxy
	prov, ok := ctx.Value(filter.ProviderCtxKey{}).(*provider.Provider)
	if !ok || prov == nil {
		panic(`provider was not set into the context`)
	}
	a.Provider = prov.Name
	return nil
}

func (a *AiAudit) SetModel(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	return a.setFieldFromRequestBody(ctx, header, buf, "model", &a.Model)
}

func (a *AiAudit) SetOperationId(ctx context.Context, _ http.Header, _ *bytes.Buffer) error {
	// a.OperationId is passed in by filter reverse-proxy
	operation, ok := ctx.Value(filter.OperationCtxKey{}).(*openapi3.Operation)
	if !ok {
		panic(fmt.Sprintf(`operation was not set into the context, ctx.Value(filter.OperationCtxKey{}) got %T`, ctx.Value(filter.OperationCtxKey{})))
	}
	if operation == nil {
		return errors.New("operation not found")
	}
	a.OperationId = operation.OperationID
	return nil
}

func (a *AiAudit) SetPrompt(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	return a.setFieldFromRequestBody(ctx, header, buf, "prompt", &a.Prompt) // todo: does not take into consideration the case where prompt is an array
}

func (a *AiAudit) SetCompletion(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit").Sub(a.OperationId)
	if !httputil.HeaderContains(header, httputil.ApplicationJson) {
		return nil
	}
	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		return errors.Wrapf(err, "failed to json.NewDecoder(%T).Decode(&%T)", buf, m)
	}
	switch a.OperationId {
	case "CreateCompletion", "CreateEdit":
		data, ok := m["choices"]
		if !ok {
			l.Debug(`no field "choices" in the response body`)
			return nil
		}
		var choices []*CreateCompletionChoice
		if err := json.Unmarshal(data, &choices); err != nil {
			l.Errorf("failed to json.Unmarshal %s to choices (%T), err: %v", string(data), choices, err)
			return err
		}
		if len(choices) == 0 {
			l.Debug(`no choice item in p.Body["choices"]`)
			return nil
		}
		a.Completion = choices[0].Text
		return nil
	case "CreateChatCompletion":
		data, ok := m["choices"]
		if !ok {
			l.Debug(`no field "choices" in the response body`)
			return nil
		}
		var choices []*CreateChatCompletionChoice
		if err := json.Unmarshal(data, &choices); err != nil {
			l.Errorf("failed to json.Unmarshal %s to choices (%T), err: %v", string(data), choices, err)
			return err
		}
		if len(choices) == 0 {
			l.Debug(`no choice item in p.Body["choices"]`)
			return nil
		}
		message := choices[0].Message
		if message == nil {
			l.Debug(`message not found in p.Body["choices"][0]`)
			return nil
		}
		a.Completion = message.Content
		return nil
	default:
		return nil
	}
}

func (a *AiAudit) SetRequestContentType(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	a.RequestContentType = header.Get(httputil.ContentTypeKey)
	return nil
}

func (a *AiAudit) SetResponseContentType(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	a.ResponseContentType = header.Get(httputil.ContentTypeKey)
	return nil
}

func (a *AiAudit) SetRequestBody(_ context.Context, _ http.Header, buf *bytes.Buffer) error {
	a.RequestBody = buf.String()
	return nil
}

func (a *AiAudit) SetResponseBody(_ context.Context, _ http.Header, buf *bytes.Buffer) error {
	a.ResponseBody = buf.String()
	return nil
}

func (a *AiAudit) SetServer(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	a.Server = header.Get("Server")
	return nil
}

func (a *AiAudit) SetStatus(_ context.Context, getter filter.HttpInfor) error {
	a.Status = getter.Status()
	a.StatusCode = getter.StatusCode()
	return nil
}

func (a *AiAudit) SetUserAgent(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	a.UserAgent = header.Get("User-Agent")
	if a.UserAgent == "" {
		a.UserAgent = header.Get("X-User-Agent")
	}
	if a.UserAgent == "" {
		a.UserAgent = header.Get("X-Device-User-Agent")
	}
	return nil
}

func (a *AiAudit) setFieldFromRequestBody(ctx context.Context, header http.Header, buf *bytes.Buffer, key string, value any) error {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit").Sub("setFieldFromRequestBody")
	if !httputil.HeaderContains(header[httputil.ContentTypeKey], httputil.ApplicationJson) {
		return nil // todo: Only Content-Type: application/json auditing is supported for now.
	}

	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		l.Errorf("failed to Decode r.Body to m (%T), err: %v", m, err)
		return err
	}
	data, ok := m[key]
	if !ok {
		l.Debugf(`no field "%s" in r.Body`, key)
		return nil
	}
	if err := json.Unmarshal(data, &value); err != nil {
		l.Errorf("failed to json.Unmarshal %s to string, err: %v", string(data), err)
		return err
	}
	return nil
}

func (*AiAudit) TableName() string {
	return "ai_proxy_filter_audit"
}

type CreateCompletionChoice struct {
	Text         string      `json:"text" yaml:"text"`
	Index        json.Number `json:"index" yaml:"index"`
	Logprobs     interface{} `json:"logprobs,omitempty" yaml:"logprobs,omitempty"`
	FinishReason string      `json:"finish_reason,omitempty" yaml:"finish_reason,omitempty"`
}

type CreateChatCompletionChoice struct {
	Index        json.Number                        `json:"index" yaml:"index"`
	Message      *CreateChatCompletionChoiceMessage `json:"message" yaml:"message"`
	FinishReason string                             `json:"finish_reason" yaml:"finish_reason"`
}

type CreateChatCompletionChoiceMessage struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
}
