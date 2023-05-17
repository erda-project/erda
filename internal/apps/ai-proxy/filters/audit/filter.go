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
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
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
	_ reverseproxy.RequestFilter  = (*Audit)(nil)
	_ reverseproxy.ResponseFilter = (*Audit)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Audit struct {
	*reverseproxy.DefaultResponseFilter

	Audit *models.AIProxyFilterAudit
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &Audit{Audit: new(models.AIProxyFilterAudit), DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}, nil
}

func (f *Audit) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
	for _, set := range []any{
		f.SetSessionId,
		f.SetChats,
		f.SetRequestAt,
		f.SetSource,
		f.SetUserInfo,
		f.SetProvider,
		f.SetModel,
		f.SetOperationId,
		f.SetPrompt,
		f.SetRequestContentType,
		f.SetRequestBody,
		f.SetUserAgent,
	} {
		switch f := set.(type) {
		case func(context.Context) error:
			if err := f(ctx); err != nil {
				l.Errorf("failed to %v, err: %v", reflect.TypeOf(set), err)
			}
		case func(context.Context, http.Header) error:
			if err := f(ctx, infor.Header()); err != nil {
				l.Errorf("failed to %v, err: %v", reflect.TypeOf(set), err)
			}
		case func(context.Context, *bytes.Buffer) error:
			if err := f(ctx, infor.BodyBuffer()); err != nil {
				l.Errorf("failed to %v, err: %v", reflect.TypeOf(set), err)
			}
		case func(context.Context, reverseproxy.HttpInfor) error:
			if err := f(ctx, infor); err != nil {
				l.Errorf("failed to %v, err: %v", reflect.TypeOf(set), err)
			}
		case func(context.Context, http.Header, *bytes.Buffer) error:
			if err := f(ctx, infor.Header(), infor.BodyBuffer()); err != nil {
				l.Errorf("failed to %v, err: %v", reflect.TypeOf(set), err)
			}
		default:
			l.Fatalf("%T not in cases", set)
		}
	}
	if err := f.create(ctx); err != nil {
		l.Errorf("failed to create audit row, err: %v", err)
	}

	return reverseproxy.Continue, nil
}

func (f *Audit) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) error {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
	if err := f.DefaultResponseFilter.OnResponseEOF(ctx, infor, w, chunk); err != nil {
		l.Errorf("failed to f.DefaultResponseFilter.OnResponseEOF, err: %v", err)
		return err
	}

	for _, set := range []any{
		f.SetResponseAt,
		f.SetCompletion,
		f.SetResponseContentType,
		f.SetResponseBody,
		f.SetServer,
		f.SetStatus,
	} {
		switch fn := set.(type) {
		case func(context.Context) error:
			if err := fn(ctx); err != nil {
				l.Errorf("failed to do %T, err: %v", set, err)
			}
		case func(context.Context, http.Header) error:
			if err := fn(ctx, infor.Header()); err != nil {
				l.Errorf("failed to do %T, err: %v", set, err)
			}
		case func(context.Context, *bytes.Buffer) error:
			if err := fn(ctx, f.Buffer); err != nil {
				l.Errorf("failed to do %T, err: %v", set, err)
			}
		case func(context.Context, http.Header, *bytes.Buffer) error:
			if err := fn(ctx, infor.Header(), f.Buffer); err != nil {
				l.Errorf("failed to do %T, err: %v", set, err)
			}
		case func(context.Context, reverseproxy.HttpInfor) error:
			if err := fn(ctx, infor); err != nil {
				l.Errorf("failed to do %T, err: %v", set, err)
			}
		default:
			l.Fatalf("%T not in cases", set)
		}
	}
	if err := f.update(ctx); err != nil {
		l.Errorf("failed to create audit row, err: %v", err)
	}
	return nil
}

func (f *Audit) SetSessionId(_ context.Context, header http.Header) error {
	f.Audit.SessionId = header.Get("X-Erda-AI-Proxy-SessionId")
	return nil
}

func (f *Audit) SetChats(_ context.Context, header http.Header) error {
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

func (f *Audit) SetRequestAt(_ context.Context) error {
	f.Audit.RequestAt = time.Now()
	f.Audit.ResponseAt = time.Now()
	return nil
}

func (f *Audit) SetResponseAt(_ context.Context) error {
	f.Audit.ResponseAt = time.Now()
	return nil
}

func (f *Audit) SetSource(_ context.Context, header http.Header) error {
	f.Audit.Source = header.Get("X-Erda-AI-Proxy-Source")
	return nil
}

func (f *Audit) SetUserInfo(ctx context.Context, header http.Header) error {
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

func (f *Audit) SetProvider(ctx context.Context) error {
	// a.Provider is passed in by filter reverse-proxy
	prov, ok := ctx.Value(reverseproxy.ProviderCtxKey{}).(*provider.Provider)
	if !ok || prov == nil {
		panic(`provider was not set into the context`)
	}
	f.Audit.Provider = prov.Name
	return nil
}

func (f *Audit) SetModel(ctx context.Context, header http.Header, buf *bytes.Buffer) error {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
	if !httputil.HeaderContains(header[httputil.ContentTypeKey], httputil.ApplicationJson) {
		return nil // todo: Only Content-Type: application/json auditing is supported for now.
	}
	if buf == nil {
		return nil
	}

	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		l.Errorf("failed to Decode r.Body to m (%T), err: %v", m, err)
		return err
	}
	var key = "model"
	data, ok := m[key]
	if !ok {
		l.Debugf(`no field "%s" in r.Body`, key)
		return nil
	}
	if err := json.Unmarshal(data, &f.Audit.Model); err != nil {
		l.Errorf("failed to json.Unmarshal %s to string, err: %v", string(data), err)
		return err
	}
	return nil
}

func (f *Audit) SetOperationId(ctx context.Context, infor reverseproxy.HttpInfor) error {
	f.Audit.OperationId = infor.Method()
	if infor.URL() != nil {
		f.Audit.OperationId += " " + infor.URL().Path
	}
	return nil
}

func (f *Audit) SetPrompt(ctx context.Context, infor reverseproxy.HttpInfor) error {
	f.Audit.Prompt = "-"
	if value := infor.Header().Get("X-Erda-AI-Proxy-Prompt"); value != "" {
		prompt, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return err
		}
		f.Audit.Prompt = string(prompt)
		return nil
	}

	if infor.Method() == http.MethodGet || infor.Method() == http.MethodDelete {
		f.Audit.Prompt = NoPromptByHttpMethod.String()
		return nil
	}
	if !httputil.HeaderContains(infor.Header(), httputil.ApplicationJson) {
		f.Audit.Prompt = NoPromptByContentType.String()
		return nil
	}
	body := infor.BodyBuffer()
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
	if buf == nil {
		return nil
	}
	if httputil.HeaderContains(header, httputil.ApplicationJson) {
		return f.setCompletionForApplicationJson(ctx, header, bytes.NewBuffer(buf.Bytes()))
	}
	if httputil.HeaderContains(header, httputil.TextEventStream) {
		return f.setCompletionForEventStream(ctx, header, bytes.NewBuffer(buf.Bytes()))
	}
	f.Audit.Completion = "unexpected response Content-Type: " + header.Get("Content-Type")
	return nil
}

func (f *Audit) SetRequestContentType(_ context.Context, header http.Header) error {
	f.Audit.RequestContentType = header.Get(httputil.ContentTypeKey)
	return nil
}

func (f *Audit) SetResponseContentType(_ context.Context, header http.Header) error {
	f.Audit.ResponseContentType = header.Get(httputil.ContentTypeKey)
	return nil
}

func (f *Audit) SetRequestBody(_ context.Context, buf *bytes.Buffer) error {
	if buf == nil {
		return nil
	}
	f.Audit.RequestBody = buf.String()
	return nil
}

func (f *Audit) SetResponseBody(_ context.Context, buf *bytes.Buffer) error {
	if buf == nil {
		return nil
	}
	f.Audit.ResponseBody = buf.String()
	return nil
}

func (f *Audit) SetServer(ctx context.Context, header http.Header) error {
	f.Audit.Server = header.Get("Server")
	if f.Audit.Server == "" {
		f.Audit.Server = ctx.Value(reverseproxy.ProviderCtxKey{}).(*provider.Provider).Name
	}
	return nil
}

func (f *Audit) SetStatus(_ context.Context, infor reverseproxy.HttpInfor) error {
	f.Audit.Status = infor.Status()
	f.Audit.StatusCode = infor.StatusCode()
	return nil
}

func (f *Audit) SetUserAgent(_ context.Context, header http.Header) error {
	f.Audit.UserAgent = header.Get("User-Agent")
	if f.Audit.UserAgent == "" {
		f.Audit.UserAgent = header.Get("X-User-Agent")
	}
	if f.Audit.UserAgent == "" {
		f.Audit.UserAgent = header.Get("X-Device-User-Agent")
	}
	return nil
}

func (f *Audit) create(ctx context.Context) error {
	return ctx.Value(reverseproxy.DBCtxKey{}).(*gorm.DB).Create(f.Audit).Error
}

func (f *Audit) update(ctx context.Context) error {
	return ctx.Value(reverseproxy.DBCtxKey{}).(*gorm.DB).Updates(f.Audit).Error
}

func (f *Audit) setCompletionForApplicationJson(ctx context.Context, header http.Header, reader io.Reader) error {
	if reader == nil {
		f.Audit.Completion = "response body nil"
		return nil
	}

	l := ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("setCompletionForApplicationJson")
	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(reader).Decode(&m); err != nil {
		return errors.Wrapf(err, "failed to json.NewDecoder(%T).Decode(&%T)", reader, m)
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

func (f *Audit) setCompletionForEventStream(ctx context.Context, header http.Header, reader io.Reader) error {
	if reader == nil {
		f.Audit.Completion = "response body nil"
		return nil
	}
	f.Audit.Completion = ExtractEventStreamCompletion(reader)
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

type EventStreamChunk struct {
	Id      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created json.Number              `json:"created"`
	Model   string                   `json:"model"`
	Choices []EventStreamChunkChoice `json:"choices"`
	Usage   interface{}              `json:"usage"`
}

type EventStreamChunkChoice struct {
	Index        int         `json:"index"`
	FinishReason interface{} `json:"finish_reason"`
	Delta        struct {
		Content string `json:"content"`
	} `json:"delta"`
}

type NoPromptReason int

func (n NoPromptReason) String() string {
	return map[NoPromptReason]string{
		NoPromptByDefault:      "NoPromptByDefault",
		NoPromptByHttpMethod:   "NoPromptByHttpMethod",
		NoPromptByContentType:  "NoPromptByContentType",
		NoPromptByNilBody:      "NoPromptByNilBody",
		NoPromptByNotParsed:    "NoPromptByNotParsed",
		NoPromptByMissingField: "NoPromptByMissingField",
		NoPromptByNoItem:       "NoPromptByNoItem",
		NoPromptByNoPrompt:     "NoPromptByNoPrompt",
		NoPromptByDeprecated:   "NoPromptByDeprecated",
		NoPromptByNoSuchRoute:  "NoPromptByNoSuchRoute",
	}[n]
}

func ExtractEventStreamCompletion(reader io.Reader) string {
	var completion string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var line = scanner.Text()
		left := strings.Index(line, "{")
		right := strings.LastIndex(line, "}")
		if left < 0 || right < 1 {
			continue
		}
		line = line[left : right+1]

		var m = make(map[string]json.RawMessage)
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		choices, ok := m["choices"]
		if !ok {
			continue
		}
		var items []EventStreamChunkChoice
		if err := json.Unmarshal(choices, &items); err != nil {
			continue
		}
		if len(items) == 0 {
			continue
		}
		completion += items[len(items)-1].Delta.Content
	}

	return completion
}
