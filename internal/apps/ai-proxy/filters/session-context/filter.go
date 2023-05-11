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

package session_context

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/dspo/roundtrip"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "session-context"
)

var (
	_ roundtrip.RequestFilter = (*Context)(nil)
)

func init() {
	filters.RegisterFilterCreator(Name, New)
}

type Context struct {
	sources map[string]struct{}
}

func New(config json.RawMessage) (roundtrip.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	var sources = make(map[string]struct{})
	for _, source := range cfg.Sources {
		sources[source] = struct{}{}
	}
	return &Context{sources: sources}, nil
}

func (c *Context) OnRequest(ctx context.Context, w http.ResponseWriter, infor roundtrip.HttpInfor) (signal roundtrip.Signal, err error) {
	var (
		l  = ctx.Value(roundtrip.CtxKeyLogger{}).(logs.Logger)
		db = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
	)

	if source := infor.Header().Get(vars.XErdaAIProxySource); source != "" {
		if _, ok := c.sources[source]; !ok {
			l.Debugf("source %s is not in config, continue", source)
			return roundtrip.Continue, nil
		}
	}
	sessionId := infor.Header().Get(vars.XErdaAIProxySessionId)
	if sessionId == "" {
		l.Debugf("sessionId is not specified, continue")
		return roundtrip.Continue, nil
	}

	session, err := db.GetSession(sessionId)
	if err != nil {
		l.Errorf("failed to db.GetSession(%s), err: %v", sessionId, err)
		return roundtrip.Continue, nil
	}
	if session.IsArchived {
		l.Debugf("session(id=%s) is archived, continue", sessionId)
		return roundtrip.Continue, nil
	}
	if session.GetContextLength() <= 1 {
		l.Debugf("session(id=%s)'s context length is less then 1, continue", sessionId)
		return roundtrip.Continue, nil
	}

	var m = make(map[string]json.RawMessage)
	if err = json.NewDecoder(infor.BodyBuffer()).Decode(&m); err != nil {
		l.Errorf("failed to json decode request body to %T, err: %v", m, err)
		return roundtrip.Continue, nil
	}
	data, ok := m["messages"]
	if !ok {
		l.Debugf("not found messages in the request body, session id %s", sessionId)
		return roundtrip.Continue, nil
	}
	var messages []Message
	if err = json.Unmarshal(data, &messages); err != nil {
		l.Errorf("failed to json.Unmarshal, data: %s, struct: %T, err: %v", string(data), messages, err)
		return roundtrip.Continue, nil
	}
	if len(messages) == 0 {
		l.Warnf("0 message found in the request")
		return roundtrip.Continue, nil
	}
	messages = []Message{messages[len(messages)-1]}
	_, chatLogs, err := db.PagingChatLogs(sessionId, int(session.GetContextLength()), 1)
	if err != nil {
		l.Debugf("failed to db.PagingChatLogs(id=%s), err: %v", sessionId)
		return roundtrip.Continue, nil
	}
	for _, chatLog := range chatLogs {
		messages = append(messages,
			Message{
				Role:    "assistant",
				Content: chatLog.GetCompletion(),
				Name:    "CodeAI", // todo: hard code here
			},
			Message{
				Role:    "user",
				Content: chatLog.GetPrompt(),
				Name:    "",
			})
	}
	messages = append(messages, Message{
		Role:    "system",
		Content: session.GetTopic(),
		Name:    "",
	})
	strutil.ReverseSlice(messages)
	m["messages"], err = json.Marshal(messages)
	if err != nil {
		l.Errorf("failed to json.Marshal(messages), err: %v", err)
		return roundtrip.Intercept, nil
	}
	data, err = json.Marshal(m)
	if err != nil {
		l.Errorf("failed to json.Marshal(m), err: %v", err)
		return roundtrip.Intercept, nil
	}
	infor.SetBody(io.NopCloser(bytes.NewBuffer(data)))
	l.Debugf("infor new body buffer: %s", infor.BodyBuffer().String())
	return roundtrip.Continue, nil
}

type Config struct {
	Sources []string `json:"sources" yaml:"sources"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}
