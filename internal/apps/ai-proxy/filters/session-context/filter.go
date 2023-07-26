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

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

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

func (c *SessionContext) OnRequest(ctx context.Context, _ http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var (
		l  = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)
		db = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
	)

	// check if this filter is enabled on this request
	ok, err := c.checkIfIsEnabledOnTheRequest(infor)
	if err != nil {
		return reverseproxy.Intercept, err
	}
	if !ok {
		return reverseproxy.Continue, nil
	}

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
		var audits []*models.AIProxyFilterAudit
		if err := db.Q().
			Where(map[string]any{"session_id": sessionId}).
			Where("updated_at > ?", session.ResetAt.AsTime()).
			Order("created_at DESC").
			Limit(int(session.GetContextLength()) - 1).
			Find(&audits).
			Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorf("failed to Find audits for session %s", sessionId)
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
					Name:    "",
				})
		}
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
		return reverseproxy.Intercept, nil
	}
	data, err = json.Marshal(m)
	if err != nil {
		l.Errorf("failed to json.Marshal(m), err: %v", err)
		return reverseproxy.Intercept, nil
	}
	infor.SetBody(io.NopCloser(bytes.NewBuffer(data)))
	l.Debugf("infor new body buffer: %s", infor.BodyBuffer().String())
	return reverseproxy.Continue, nil
}

func (c *SessionContext) checkIfIsEnabledOnTheRequest(infor reverseproxy.HttpInfor) (bool, error) {
	for i, item := range c.Config.On {
		if item == nil {
			continue
		}
		ok, err := item.On(infor.Header())
		if err != nil {
			return false, errors.Wrapf(err, "invalid config: config.on[%d]", i)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

type Config struct {
	On []*common.On
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}
