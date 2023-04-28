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
	"net/http"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
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
	audit *models.AIProxyFilterAudit
}

func New(_ json.RawMessage) (filter.Filter, error) {
	return &Audit{audit: new(models.AIProxyFilterAudit)}, nil
}

func (f *Audit) OnHttpRequestGetter(ctx context.Context, g filter.HttpInfor) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("Audit").Sub("OnHttpRequestGetter")
	for _, set := range []func(context.Context, http.Header, *bytes.Buffer) error{
		f.SetSessionId,
		f.SetChats,
		f.SetRequestAt,
		f.SetSource,
		f.SetUserInfo,
		f.SetRequestContentType,
		f.SetUserAgent,
		f.SetRequestBody,
		f.SetProvider,
		f.SetModel,
		f.SetPrompt,
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
	for _, set := range []func(ctx2 context.Context, infor filter.HttpInfor) error{
		f.SetOperationId,
	} {
		if err := set(ctx, g); err != nil {
			l.Errorf("failed to %v, err: %v", reflect.TypeOf(set), err)
			continue
		}
	}
	return filter.Continue, nil
}

func (f *Audit) OnHttpResponseGetter(ctx context.Context, getter filter.HttpInfor) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("Audit").Sub("OnHttpResponse")
	for _, set := range []func(context.Context, http.Header, *bytes.Buffer) error{
		f.SetResponseAt,
		f.SetCompletion,
		f.SetResponseContentType,
		f.SetResponseBody,
		f.SetServer,
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
		f.SetStatus,
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

func (f *Audit) SetSessionId(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.audit.SessionId = header.Get("X-Erda-AI-Proxy-SessionId") // todo: Temporary
	return nil
}

func (f *Audit) SetChats(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.audit.ChatType = header.Get("X-Erda-AI-Proxy-ChatType")
	f.audit.ChatTitle = header.Get("X-Erda-AI-Proxy-ChatTitle")
	f.audit.ChatId = header.Get("X-Erda-AI-Proxy-ChatId")
	for _, v := range []*string{
		&f.audit.ChatType,
		&f.audit.ChatTitle,
		&f.audit.ChatId,
	} {
		if decoded, err := base64.StdEncoding.DecodeString(*v); err == nil {
			*v = string(decoded)
		}
	}
	return nil
}

func (f *Audit) SetRequestAt(_ context.Context, _ http.Header, _ *bytes.Buffer) error {
	f.audit.RequestAt = time.Now()
	return nil
}

func (f *Audit) SetResponseAt(_ context.Context, _ http.Header, _ *bytes.Buffer) error {
	f.audit.ResponseAt = time.Now()
	return nil
}

func (f *Audit) SetSource(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.audit.Source = header.Get("X-Erda-AI-Proxy-Source")
	return nil
}

func (f *Audit) SetUserInfo(ctx context.Context, header http.Header, _ *bytes.Buffer) error {
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
	f.audit.UserInfo = string(data)
	return nil
}

func (f *Audit) SetProvider(ctx context.Context, _ http.Header, _ *bytes.Buffer) error {
	// a.Provider is passed in by filter reverse-proxy
	prov, ok := ctx.Value(filter.ProviderCtxKey{}).(*provider.Provider)
	if !ok || prov == nil {
		panic(`provider was not set into the context`)
	}
	f.audit.Provider = prov.Name
	return nil
}

func (f *Audit) SetModel(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	return f.setFieldFromRequestBody(ctx, header, buf, "model", &f.audit.Model)
}

func (f *Audit) SetOperationId(ctx context.Context, infor filter.HttpInfor) error {
	f.audit.OperationId = infor.Method()
	if infor.URL() != nil {
		f.audit.OperationId += " " + infor.URL().Path
	}
	return nil
}

func (f *Audit) SetPrompt(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	return f.setFieldFromRequestBody(ctx, header, buf, "prompt", &f.audit.Prompt) // todo: does not take into consideration the case where prompt is an array
}

func (f *Audit) SetCompletion(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit").Sub(f.audit.OperationId)
	if !httputil.HeaderContains(header, httputil.ApplicationJson) {
		return nil
	}
	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		return errors.Wrapf(err, "failed to json.NewDecoder(%T).Decode(&%T)", buf, m)
	}
	switch f.audit.OperationId {
	case "POST /v1/completions", "POST /v1/edits":
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
		f.audit.Completion = choices[0].Text
		return nil
	case "POST /v1/chat/completions":
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
		f.audit.Completion = message.Content
		return nil
	default:
		return nil
	}
}

func (f *Audit) SetRequestContentType(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.audit.RequestContentType = header.Get(httputil.ContentTypeKey)
	return nil
}

func (f *Audit) SetResponseContentType(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.audit.ResponseContentType = header.Get(httputil.ContentTypeKey)
	return nil
}

func (f *Audit) SetRequestBody(_ context.Context, _ http.Header, buf *bytes.Buffer) error {
	f.audit.RequestBody = buf.String()
	return nil
}

func (f *Audit) SetResponseBody(_ context.Context, _ http.Header, buf *bytes.Buffer) error {
	f.audit.ResponseBody = buf.String()
	return nil
}

func (f *Audit) SetServer(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.audit.Server = header.Get("Server")
	return nil
}

func (f *Audit) SetStatus(_ context.Context, getter filter.HttpInfor) error {
	f.audit.Status = getter.Status()
	f.audit.StatusCode = getter.StatusCode()
	return nil
}

func (f *Audit) SetUserAgent(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.audit.UserAgent = header.Get("User-Agent")
	if f.audit.UserAgent == "" {
		f.audit.UserAgent = header.Get("X-User-Agent")
	}
	if f.audit.UserAgent == "" {
		f.audit.UserAgent = header.Get("X-Device-User-Agent")
	}
	return nil
}

func (f *Audit) setFieldFromRequestBody(ctx context.Context, header http.Header, buf *bytes.Buffer, key string, value any) error {
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
