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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "audit"
)

var (
	_ filter.Filter = (*Audit)(nil)
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

func (f *Audit) OnHttpRequestHeader(ctx context.Context, header http.Header) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("Audit")
	for name, set := range map[string]func(context.Context, http.Header) error{
		"SetSessionId":          f.audit.SetSessionId,
		"SetChat":               f.audit.SetChats,
		"SetRequestAt":          f.audit.SetRequestAt,
		"SetSource":             f.audit.SetSource,
		"SetUserInfo":           f.audit.SetUserInfo,
		"SetOperationId":        f.audit.SetOperationId,
		"SetRequestContentType": f.audit.SetRequestContentType,
		"SetUserAgent":          f.audit.SetUserAgent,
	} {
		if err := set(ctx, header); err != nil {
			l.Errorf(`failed to %s, err: %v`, name)
			continue
		}
	}
	return filter.Continue, nil
}

func (f *Audit) OnHttpRequestBodyCopy(ctx context.Context, reader io.Reader) (filter.Signal, error) {
	for _, set := range []func(context.Context, io.Reader) error{
		f.audit.SetRequestBody,
	} {
		if err := set(ctx, reader); err != nil {
			return filter.Intercept, err
		}
	}
	return filter.Continue, nil
}

func (f *Audit) OnHttpRequest(ctx context.Context, _ http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	for _, set := range []func(ctx2 context.Context, r *http.Request) error{
		f.audit.SetModel,
		f.audit.SetPrompt,
	} {
		if err := set(ctx, r); err != nil {
			return filter.Intercept, nil
		}
	}

	return filter.Continue, nil
}

func (f *Audit) OnHttpResponse(ctx context.Context, response *http.Response) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("Audit").Sub("OnHttpResponse")
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return filter.Intercept, err
	}
	defer func() {
		_ = response.Body.Close()
		response.Body = io.NopCloser(bytes.NewReader(data))
	}()
	for _, set := range []func(context.Context, http.Header, io.Reader) error{
		f.audit.SetResponseAt,
		f.audit.SetCompletion,
		f.audit.SetResponseContentType,
		f.audit.SetResponseBody,
	} {
		if err := set(ctx, response.Header, bytes.NewReader(data)); err != nil {
			l.Errorf("failed to do %T, err: %v", set, err)
			continue
		}
	}
	for _, set := range []func(ctx2 context.Context, response2 *http.Response) error{
		f.audit.SetServer,
		f.audit.SetStatus,
	} {
		if err := set(ctx, response); err != nil {
			l.Errorf("failed to do %T, err: %v", set, err)
			continue
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

func (a *AiAudit) SetSessionId(_ context.Context, header http.Header) error {
	a.SessionId = header.Get("X-Erda-AI-Proxy-SessionId") // todo: Temporary
	return nil
}

func (a *AiAudit) SetChats(ctx context.Context, header http.Header) error {
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

func (a *AiAudit) SetRequestAt(_ context.Context, _ http.Header) error {
	a.RequestAt = time.Now()
	return nil
}

func (a *AiAudit) SetResponseAt(_ context.Context, _ http.Header, _ io.Reader) error {
	a.ResponseAt = time.Now()
	return nil
}

func (a *AiAudit) SetSource(_ context.Context, header http.Header) error {
	a.Source = header.Get("X-Erda-AI-Proxy-Source")
	return nil
}

func (a *AiAudit) SetUserInfo(ctx context.Context, header http.Header) error {
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

func (a *AiAudit) SetProvider(ctx context.Context, _ http.Header) error {
	// a.Provider is passed in by filter reverse-proxy
	prov, ok := ctx.Value(filter.ProviderCtxKey{}).(*provider.Provider)
	if !ok || prov == nil {
		panic(`provider was not set into the context`)
	}
	a.Provider = prov.Name
	return nil
}

func (a *AiAudit) SetModel(ctx context.Context, r *http.Request) error {
	return a.setFieldFromRequestBody(ctx, r, "model", &a.Model)
}

func (a *AiAudit) SetOperationId(ctx context.Context, _ http.Header) error {
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

func (a *AiAudit) SetPrompt(ctx context.Context, r *http.Request) error {
	return a.setFieldFromRequestBody(ctx, r, "prompt", &a.Prompt) // todo: does not take into consideration the case where prompt is an array
}

func (a *AiAudit) SetCompletion(ctx context.Context, _ http.Header, body io.Reader) error {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit").Sub(a.OperationId)
	switch a.OperationId {
	case "CreateCompletion", "CreateEdit":
		var m = make(map[string]json.RawMessage)
		if err := json.NewDecoder(body).Decode(&m); err != nil {
			l.Errorf("failed to Decode p.Body to m (%T), err: %v", m, err)
			return err
		}
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
		var m = make(map[string]json.RawMessage)
		if err := json.NewDecoder(body).Decode(&m); err != nil {
			l.Errorf("failed to Decode p.Body to m (%T), err: %v", m, err)
			return err
		}
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
		a.Prompt = message.Content
		return nil
	default:
		return nil
	}
}

func (a *AiAudit) SetRequestContentType(_ context.Context, header http.Header) error {
	a.RequestContentType = header.Get("Content-Type")
	return nil
}

func (a *AiAudit) SetResponseContentType(_ context.Context, header http.Header, _ io.Reader) error {
	a.ResponseContentType = header.Get("Content-Type")
	return nil
}

func (a *AiAudit) SetRequestBody(ctx context.Context, reader io.Reader) error {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit")
	data, err := io.ReadAll(reader)
	if err != nil {
		l.Errorf(`failed to ReadAll(r.Body), err: %v`, err)
		return err
	}
	a.RequestBody = string(data)
	return nil
}

func (a *AiAudit) SetResponseBody(ctx context.Context, _ http.Header, body io.Reader) error {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit")
	data, err := io.ReadAll(body)
	if err != nil {
		l.Errorf(`failed to ReadAll(response.Body), err: %v`, err)
		return err
	}
	if len(data) == 0 {
		l.Warnf(`[SetResponseBody] no data in response body`)
		return nil
	}
	a.ResponseBody = string(data)
	return nil
}

func (a *AiAudit) SetServer(ctx context.Context, response *http.Response) error {
	a.Server = response.Header.Get("Server")
	return nil
}

func (a *AiAudit) SetStatus(ctx context.Context, response *http.Response) error {
	a.Status = response.Status
	a.StatusCode = response.StatusCode
	return nil
}

func (a *AiAudit) SetUserAgent(_ context.Context, header http.Header) error {
	a.UserAgent = header.Get("User-Agent")
	if a.UserAgent == "" {
		a.UserAgent = header.Get("X-User-Agent")
	}
	if a.UserAgent == "" {
		a.UserAgent = header.Get("X-Device-User-Agent")
	}
	return nil
}

func (a *AiAudit) setFieldFromRequestBody(ctx context.Context, r *http.Request, key string, value any) error {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit")
	if !strutil.Equal(r.Method, http.MethodPost) {
		return nil
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return nil // todo: Only Content-Type: application/json auditing is supported for now.
	}
	if r.Body == nil {
		return nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Errorf(`failed to io.ReadAll(r.Body): %v`, err)
		return err
	}
	defer func() {
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}()
	var m = make(map[string]json.RawMessage)
	if err := json.Unmarshal(body, &m); err != nil {
		l.Errorf("failed to Decode r.Body to m (%T), err: %v", m, err)
		return err
	}
	data, ok := m[key]
	if !ok {
		l.Debug(`no field "model" in r.Body`)
		return nil
	}
	if err := json.Unmarshal(data, &value); err != nil {
		l.Errorf("failed to json.Unmarshal %s to string, err: %v", string(data), err)
		return err
	}
	return nil
}

func (*AiAudit) TableName() string {
	return "filter_audit"
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
