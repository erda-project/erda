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

package dao

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
)

var (
	_ DAO = (*provider)(nil)
)

var (
	name = "erda.apps.ai-proxy.dao"
	spec = servicehub.Spec{
		Services:    []string{"erda.apps.ai-proxy.dao"},
		Summary:     "erda.apps.ai-proxy.dao",
		Description: "erda.apps.ai-proxy.dao",
		ConfigFunc: func() any {
			return new(struct{})
		},
		Types: pb.Types(),
		Creator: func() servicehub.Provider {
			return new(provider)
		},
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type DAO interface {
	Q() *gorm.DB
	Tx() *gorm.DB

	PagingChatLogs(sessionId string, pageNum, pageSize int) (int64, []*pb.ChatLog, error)
	CreateSession(userId, name, topic string, contextLength uint32, source, model string, temperature float64) (id string, err error)
	UpdateSession(id string, setters ...models.Setter) error
	DeleteSession(id string) error
	GetSession(id string) (*pb.Session, bool, error)
}

type provider struct {
	DB *gorm.DB `autowired:"mysql-gorm.v2-client"`
}

func (p *provider) Provide(ctx servicehub.DependencyContext, options ...any) any {
	return p
}

func (p *provider) Q() *gorm.DB {
	return p.DB
}

func (p *provider) Tx() *gorm.DB {
	return p.DB.Session(&gorm.Session{})
}

func (p *provider) PagingChatLogs(sessionId string, pageNum, pageSize int) (int64, []*pb.ChatLog, error) {
	var audits models.AIProxyFilterAuditList
	total, err := (&audits).Pager(p.DB).
		Where(audits.FieldSessionID().Equal(sessionId)).
		Paging(pageSize, pageNum, audits.FieldCreatedAt().DESC())
	if err != nil {
		return 0, nil, err
	}
	if total == 0 {
		return 0, nil, nil
	}
	var chatLogs []*pb.ChatLog
	for _, item := range audits {
		chatLogs = append(chatLogs, item.ToProtobufChatLog())
	}
	return 0, chatLogs, nil
}

func (p *provider) CreateSession(userId, name, topic string, contextLength uint32, source, model string, temperature float64) (id string, err error) {
	var session = models.AIProxySessions{
		UserID:        userId,
		Name:          name,
		Topic:         topic,
		ContextLength: int64(contextLength),
		Source:        source,
		IsArchived:    false,
		ResetAt:       time.Unix(0, 0),
		Model:         model,
		Temperature:   strconv.FormatFloat(temperature, 'f', 1, 64),
	}
	if err := (&session).Creator(p.DB).Create(); err != nil {
		return "", err
	}
	return session.ID.String, nil
}

func (p *provider) UpdateSession(id string, setters ...models.Setter) error {
	var session models.AIProxySessions
	var where = session.FieldID().Equal(id)
	ok, err := (&session).Getter(p.DB).Where(where).Get()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	_, err = (&session).Updater(p.DB).Where(where).Updates(setters...)
	return err
}

func (p *provider) DeleteSession(id string) error {
	var session = new(models.AIProxySessions)
	_, err := session.Deleter(p.DB).Where(session.FieldID().Equal(id)).Delete()
	return err
}

func (p *provider) GetSession(id string) (*pb.Session, bool, error) {
	var session models.AIProxySessions
	ok, err := (&session).Getter(p.DB).Where(session.FieldID().Equal(id)).Get()
	return session.ToProtobuf(), ok, err
}
