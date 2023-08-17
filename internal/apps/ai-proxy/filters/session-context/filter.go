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
	"time"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "session-context"
)

var (
	_ reverseproxy.RequestFilter = (*SessionContext)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type SessionContext struct {
	Config *Config
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &SessionContext{Config: &cfg}, nil
}

func (c *SessionContext) Enable(_ context.Context, req *http.Request) bool {
	for _, item := range c.Config.On {
		if item == nil {
			continue
		}
		if ok, _ := item.On(req.Header); ok {
			return true
		}
	}
	return false
}

func (c *SessionContext) OnRequest(ctx context.Context, _ http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var (
		l  = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
		db = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
	)

	sessionId := infor.Header().Get(vars.XAIProxySessionId)
	if sessionId == "" {
		l.Debugf("sessionId is not specified, continue")
		return reverseproxy.Continue, nil
	}

	session, ok, err := db.GetSession(sessionId)
	if err != nil {
		l.Errorf("failed to db.GetSession(%s), err: %v", sessionId, err)
		return reverseproxy.Continue, nil
	}
	if !ok {
		l.Errorf("session not found, sessionId: %s", sessionId)
		return reverseproxy.Continue, nil
	}
	if session.IsArchived {
		l.Debugf("session(id=%s) is archived, continue", sessionId)
		return reverseproxy.Continue, nil
	}

	// make the session is the latest updated
	defer func() { go c.updateSession(ctx, sessionId) }()

	var m = make(map[string]json.RawMessage)
	if err = json.NewDecoder(infor.BodyBuffer()).Decode(&m); err != nil {
		l.Errorf("failed to json decode request body to %T, err: %v", m, err)
		return reverseproxy.Continue, nil
	}
	data, ok := m["messages"]
	if !ok {
		l.Debugf("not found messages in the request body, session id %s", sessionId)
		return reverseproxy.Continue, nil
	}
	var messages []Message
	if err = json.Unmarshal(data, &messages); err != nil {
		l.Errorf("failed to json.Unmarshal, data: %s, struct: %T, err: %v", string(data), messages, err)
		return reverseproxy.Continue, nil
	}
	if len(messages) == 0 {
		l.Warnf("0 message found in the request")
		return reverseproxy.Continue, nil
	}
	messages = []Message{messages[len(messages)-1]}
	if session.GetContextLength() > 1 {
		var audits models.AIProxyFilterAuditList
		total, err := (&audits).Pager(db.Q().Debug()).
			Where(
				audits.FieldSessionID().Equal(sessionId),
				audits.FieldUpdatedAt().MoreThan(session.ResetAt.AsTime()),
			).
			Paging(int(session.GetContextLength()-1), 1, audits.FieldCreatedAt().DESC())
		if err != nil {
			l.Errorf("failed to Find audits in the session %s, err: %v", sessionId, err)
		}
		if total == 0 {
			l.Debugf("no context in the session %s", sessionId)
		}
		for _, item := range audits {
			messages = append(messages,
				Message{
					Role:    "assistant",
					Content: item.Completion,
					Name:    "CodeAI", // todo: hard code here
				},
				Message{
					Role:    "user",
					Content: item.Prompt,
					Name:    "erda",
				})
		}
	} else {
		l.Debugf("session context length is less then 1, no context appended")
	}
	messages = append(messages,
		Message{
			Role:    "system",
			Content: "topic: " + session.GetTopic(),
			Name:    "system",
		},
		Message{
			Role:    "system",
			Content: c.Config.SysMsg,
			Name:    "system",
		})
	strutil.ReverseSlice(messages)
	m["messages"], err = json.Marshal(messages)
	if err != nil {
		l.Errorf("failed to json.Marshal(messages), err: %v", err)
		return reverseproxy.Intercept, nil
	}
	data, err = json.Marshal(m)
	if err != nil {
		l.Errorf("failed to json.Marshal(m), err: %v", err)
		return reverseproxy.Intercept, nil
	}
	infor.SetBody(io.NopCloser(bytes.NewBuffer(data)), int64(len(data)))
	l.Debugf("infor new body buffer: %s", infor.BodyBuffer().String())
	return reverseproxy.Continue, nil
}

func (c *SessionContext) updateSession(ctx context.Context, id string) {
	var (
		l       = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
		db      = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
		session models.AIProxySessions
	)
	if _, err := (&session).Updater(db.Q().Debug()).
		Where(session.FieldID().Equal(id)).
		Set(session.FieldUpdatedAt().Set(time.Now())).
		Updates(); err != nil {
		l.Errorf("failed to update session, err: %v", err)
	}
}

type Config struct {
	SysMsg string       `json:"sysMsg" yaml:"sysMsg"`
	On     []*common.On `json:"on" yaml:"on"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}
