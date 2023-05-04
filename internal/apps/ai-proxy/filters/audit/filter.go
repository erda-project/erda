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
	"regexp"
	"strconv"
	"strings"
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
	NoPromptByDefault NoPromptReason = iota
	NoPromptByHttpMethod
	NoPromptByContentType
	NoPromptByNilBody
	NoPromptByNotParsed
	NoPromptByMissingField
	NoPromptByNoItem
	NoPromptByNoPrompt
	NoPromptByDeprecated
	NoPromptByNoSuchRoute
)

const (
	Name = "audit"
)

var (
	_ filter.RequestInforFilter  = (*Audit)(nil)
	_ filter.ResponseInforFilter = (*Audit)(nil)
)

func init() {
	filter.Register(Name, New)
}

type Audit struct {
	Audit *models.AIProxyFilterAudit
}

func New(_ json.RawMessage) (filter.Filter, error) {
	return &Audit{Audit: new(models.AIProxyFilterAudit)}, nil
}

func (f *Audit) OnHttpRequestInfor(ctx context.Context, g filter.HttpInfor) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("Audit").Sub("OnHttpRequestInfor")
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
		f.SetPrompt,
	} {
		if err := set(ctx, g); err != nil {
			l.Errorf("failed to %v, err: %v", reflect.TypeOf(set), err)
			continue
		}
	}
	return filter.Continue, nil
}

func (f *Audit) OnHttpResponseInfor(ctx context.Context, getter filter.HttpInfor) (filter.Signal, error) {
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
	return db.Create(f.Audit).Error
}

func (f *Audit) SetSessionId(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.Audit.SessionId = header.Get("X-Erda-AI-Proxy-SessionId") // todo: Temporary
	return nil
}

func (f *Audit) SetChats(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.Audit.ChatType = header.Get("X-Erda-AI-Proxy-ChatType")
	f.Audit.ChatTitle = header.Get("X-Erda-AI-Proxy-ChatTitle")
	f.Audit.ChatId = header.Get("X-Erda-AI-Proxy-ChatId")
	for _, v := range []*string{
		&f.Audit.ChatType,
		&f.Audit.ChatTitle,
		&f.Audit.ChatId,
	} {
		if decoded, err := base64.StdEncoding.DecodeString(*v); err == nil {
			*v = string(decoded)
		}
	}
	return nil
}

func (f *Audit) SetRequestAt(_ context.Context, _ http.Header, _ *bytes.Buffer) error {
	f.Audit.RequestAt = time.Now()
	return nil
}

func (f *Audit) SetResponseAt(_ context.Context, _ http.Header, _ *bytes.Buffer) error {
	f.Audit.ResponseAt = time.Now()
	return nil
}

func (f *Audit) SetSource(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.Audit.Source = header.Get("X-Erda-AI-Proxy-Source")
	return nil
}

func (f *Audit) SetUserInfo(ctx context.Context, header http.Header, _ *bytes.Buffer) error {
	f.Audit.Username = header.Get("X-Erda-AI-Proxy-Name")
	f.Audit.PhoneNumber = header.Get("X-Erda-AI-Proxy-Phone")
	f.Audit.JobNumber = header.Get("X-Erda-AI-Proxy-JobNumber")
	f.Audit.Email = header.Get("X-Erda-AI-Proxy-Email")
	f.Audit.DingtalkStaffId = header.Get("X-Erda-AI-Proxy-DingTalkStaffID")
	for _, v := range []*string{
		&f.Audit.Username,
		&f.Audit.PhoneNumber,
		&f.Audit.JobNumber,
		&f.Audit.Email,
		&f.Audit.DingtalkStaffId,
	} {
		if decoded, err := base64.StdEncoding.DecodeString(*v); err == nil {
			*v = string(decoded)
		}
	}
	return nil
}

func (f *Audit) SetProvider(ctx context.Context, _ http.Header, _ *bytes.Buffer) error {
	// a.Provider is passed in by filter reverse-proxy
	prov, ok := ctx.Value(filter.ProviderCtxKey{}).(*provider.Provider)
	if !ok || prov == nil {
		panic(`provider was not set into the context`)
	}
	f.Audit.Provider = prov.Name
	return nil
}

func (f *Audit) SetModel(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	return f.setFieldFromRequestBody(ctx, header, buf, "model", &f.Audit.Model)
}

func (f *Audit) SetOperationId(ctx context.Context, infor filter.HttpInfor) error {
	f.Audit.OperationId = infor.Method()
	if infor.URL() != nil {
		f.Audit.OperationId += " " + infor.URL().Path
	}
	return nil
}

func (f *Audit) SetPrompt(ctx context.Context, infor filter.HttpInfor) error {
	f.Audit.Prompt = "-"
	if infor.Method() == http.MethodGet || infor.Method() == http.MethodDelete {
		f.Audit.Prompt = NoPromptByHttpMethod.String()
		return nil
	}
	if !httputil.HeaderContains(infor.Header(), httputil.ApplicationJson) {
		f.Audit.Prompt = NoPromptByContentType.String()
		return nil
	}
	body, err := infor.Body()
	if err != nil {
		f.Audit.Prompt = NoPromptByNilBody.String()
		return err
	}
	if body == nil {
		f.Audit.Prompt = NoPromptByNilBody.String()
		return nil
	}
	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(body).Decode(&m); err != nil {
		f.Audit.Prompt = NoPromptByNotParsed.String()
		return err
	}
	switch method, path, operation := infor.Method(), infor.URL().Path, infor.Method()+" "+infor.URL().Path; {
	case method == http.MethodPost && path == "/v1/completions":
		message, ok := m["prompt"]
		if !ok {
			f.Audit.Prompt = NoPromptByMissingField.String()
			return errors.Errorf(`no field "prompt" in the request body, operation: %s`, operation)
		}
		{
			var prompt []int
			if err := json.Unmarshal(message, &prompt); err == nil {
				f.Audit.Prompt = string(message)
				return nil
			}
		}
		{
			var prompt [][]int
			if err := json.Unmarshal(message, &prompt); err == nil {
				f.Audit.Prompt = string(message)
				return nil
			}
		}
		{
			var prompt []string
			if err := json.Unmarshal(message, &prompt); err != nil {
				f.Audit.Prompt = strings.Join(prompt, "\n")
				return nil
			}
		}
		{
			var prompt string
			if err := json.Unmarshal(message, &prompt); err != nil {
				f.Audit.Prompt = string(message)
				return nil
			}
		}
		f.Audit.Prompt = NoPromptByNotParsed.String()
	case method == http.MethodPost && path == "/v1/chat/completions":
		message, ok := m["messages"]
		if !ok {
			f.Audit.Prompt = NoPromptByMissingField.String()
			return errors.Errorf(`no field "messages" in the request body, operation: %s`, operation)
		}
		var messages []struct {
			Content string `json:"content" yaml:"content"`
		}
		if err := json.Unmarshal(message, &messages); err != nil {
			f.Audit.Prompt = NoPromptByNotParsed.String()
			return err
		}
		if len(messages) == 0 {
			f.Audit.Prompt = NoPromptByNoItem.String()
			return errors.Errorf(`no itmes in the request body messages`)
		}
		f.Audit.Prompt = messages[len(messages)-1].Content
	case method == http.MethodPost && path == "/v1/edits":
		message, ok := m["edit"]
		if !ok {
			f.Audit.Prompt = NoPromptByMissingField.String()
			return errors.Errorf(`no field "edit" in the request body, operation: %s`, operation)
		}
		f.Audit.Prompt = string(message)
	case method == http.MethodPost && path == "/v1/images/generations":
		message, ok := m["prompt"]
		if !ok {
			f.Audit.Prompt = NoPromptByMissingField.String()
			return errors.Errorf(`no field "prompt" in the request body, operation: %s`, operation)
		}
		f.Audit.Prompt = string(message)
	case method == http.MethodPost && path == "/v1/images/edits": // multipart/form-data
		f.Audit.Prompt = NoPromptByContentType.String()
	case method == http.MethodPost && path == "/v1/images/variations": // multipart/form-data
		f.Audit.Prompt = NoPromptByContentType.String()
	case method == http.MethodPost && path == "/v1/embeddings":
		message, ok := m["input"]
		if !ok {
			f.Audit.Prompt = NoPromptByMissingField.String()
			return errors.Errorf(`no field "input" in the request body, operation: %s`, operation)
		}
		{
			var prompt []int
			if err := json.Unmarshal(message, &prompt); err == nil {
				f.Audit.Prompt = string(message)
				return nil
			}
		}
		{
			var prompt [][]int
			if err := json.Unmarshal(message, &prompt); err == nil {
				f.Audit.Prompt = string(message)
				return nil
			}
		}
		{
			var prompt []string
			if err := json.Unmarshal(message, &prompt); err != nil {
				f.Audit.Prompt = strings.Join(prompt, "\n")
				return nil
			}
		}
		{
			var prompt string
			if err := json.Unmarshal(message, &prompt); err != nil {
				f.Audit.Prompt = string(message)
				return nil
			}
		}
		f.Audit.Prompt = NoPromptByNotParsed.String()
	case method == http.MethodPost && path == "/v1/audio/transcriptions": // multipart/form-data
		f.Audit.Prompt = NoPromptByContentType.String()
	case method == http.MethodPost && path == "/v1/audio/translations": // multipart/form-data
		f.Audit.Prompt = NoPromptByContentType.String()
	case method == http.MethodGet && path == "/v1/files": // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodPost && path == "/v1/files": // multipart/form-data
		f.Audit.Prompt = NoPromptByContentType.String()
	case method == http.MethodPost && path == "/v1/fine-tunes": // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodGet && path == "/v1/fine-tunes": // no prompt
		f.Audit.Prompt = NoPromptByNoPrompt.String()
	case method == http.MethodGet && path == "/v1/models": // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodPost && path == "/v1/moderations":
		message, ok := m["input"]
		if !ok {
			f.Audit.Prompt = NoPromptByMissingField.String()
			return errors.Errorf(`no field "input" in the request body, operation: %s`, operation)
		}
		{
			var prompt []string
			if err := json.Unmarshal(message, &prompt); err != nil {
				f.Audit.Prompt = strings.Join(prompt, "\n")
				return nil
			}
		}
		{
			var prompt string
			if err := json.Unmarshal(message, &prompt); err != nil {
				f.Audit.Prompt = string(message)
				return nil
			}
		}
		f.Audit.Prompt = NoPromptByNotParsed.String()
	case method == http.MethodGet && path == "/v1/engine": // deprecated:
		f.Audit.Prompt = NoPromptByDeprecated.String()
	case method == http.MethodGet && regexp.MustCompile(`/v1/models/([^/.]+)$`).MatchString(path): // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodDelete && regexp.MustCompile(`/v1/models/([^/.]+)$`).MatchString(path): // http.MethodDelete
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodDelete && regexp.MustCompile(`^/v1/files/([^/.]+)$`).MatchString(path): // http.MethodDelete
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodGet && regexp.MustCompile(`^/v1/files/([^/.]+)$`).MatchString(path): // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodGet && regexp.MustCompile(`^/v1/files/([^/.]+)/content$`).MatchString(path): // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodGet && regexp.MustCompile("/v1/fine-tunes/([^/.]+)$").MatchString(path): // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodPost && regexp.MustCompile(`/v1/fine-tunes/([^/.]+)/cancel$`).MatchString(path): // no request body
		f.Audit.Prompt = NoPromptByNilBody.String()
	case method == http.MethodGet && regexp.MustCompile(`/v1/fine-tunes/([^/.]+)/events$`).MatchString(path): // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	case method == http.MethodGet && regexp.MustCompile(`/v1/engines/([^/.]+)$`).MatchString(path): // http.MethodGet
		f.Audit.Prompt = NoPromptByHttpMethod.String()
	default:
		f.Audit.Prompt = NoPromptByNoSuchRoute.String()
	}
	return nil
}

func (f *Audit) SetCompletion(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("AiAudit").Sub(f.Audit.OperationId)
	if !httputil.HeaderContains(header, httputil.ApplicationJson) {
		return nil
	}
	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		return errors.Wrapf(err, "failed to json.NewDecoder(%T).Decode(&%T)", buf, m)
	}
	switch f.Audit.OperationId {
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
		f.Audit.Completion = choices[0].Text
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
		f.Audit.Completion = message.Content
		return nil
	default:
		return nil
	}
}

func (f *Audit) SetRequestContentType(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.Audit.RequestContentType = header.Get(httputil.ContentTypeKey)
	return nil
}

func (f *Audit) SetResponseContentType(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.Audit.ResponseContentType = header.Get(httputil.ContentTypeKey)
	return nil
}

func (f *Audit) SetRequestBody(_ context.Context, _ http.Header, buf *bytes.Buffer) error {
	f.Audit.RequestBody = buf.String()
	return nil
}

func (f *Audit) SetResponseBody(_ context.Context, _ http.Header, buf *bytes.Buffer) error {
	f.Audit.ResponseBody = buf.String()
	return nil
}

func (f *Audit) SetServer(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.Audit.Server = header.Get("Server")
	return nil
}

func (f *Audit) SetStatus(_ context.Context, getter filter.HttpInfor) error {
	f.Audit.Status = getter.Status()
	f.Audit.StatusCode = getter.StatusCode()
	return nil
}

func (f *Audit) SetUserAgent(_ context.Context, header http.Header, _ *bytes.Buffer) error {
	f.Audit.UserAgent = header.Get("User-Agent")
	if f.Audit.UserAgent == "" {
		f.Audit.UserAgent = header.Get("X-User-Agent")
	}
	if f.Audit.UserAgent == "" {
		f.Audit.UserAgent = header.Get("X-Device-User-Agent")
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

type NoPromptReason int

func (n NoPromptReason) String() string {
	return strconv.FormatInt(int64(n), 10)
}
