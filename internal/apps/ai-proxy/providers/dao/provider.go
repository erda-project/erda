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
	"database/sql"
	"time"

	"github.com/pkg/errors"
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
	Model(value any) *gorm.DB
	Create(value any) *gorm.DB
	Update(column string, value any) *gorm.DB
	Updates(values any) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB

	PagingChatLogs(sessionId string, pageNum, pageSize int) (int64, []*pb.ChatLog, error)
	CreateSession(userId, name, topic string, contextLength uint32, source, model string, temperature float64) (id string, err error)
	UpdateSession(id string, updates map[string]any) error
	DeleteSession(id string) error
	ListSessions(where map[string]any, pageNum, pageSise int) (int64, []*pb.Session, error)
	GetSession(id string) (*pb.Session, error)
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

func (p *provider) Model(value any) *gorm.DB {
	return p.DB.Model(value)
}

func (p *provider) Create(value any) *gorm.DB {
	return p.DB.Create(value)
}

func (p *provider) Update(column string, value any) *gorm.DB {
	return p.DB.Update(column, value)
}

func (p *provider) Updates(values any) *gorm.DB {
	return p.DB.Updates(values)
}

func (p *provider) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return p.DB.Find(dest, conds)
}

func (p *provider) PagingChatLogs(sessionId string, pageNum, pageSize int) (int64, []*pb.ChatLog, error) {
	var (
		audits []*models.AIProxyFilterAudit
		count  int64
	)
	if err := p.DB.Model(new(models.AIProxyFilterAudit)).
		Where(map[string]any{"session_id": sessionId}).
		Count(&count).
		Limit(pageSize).Offset((pageNum - 1) * pageSize).Order("created_at DESC").Find(&audits).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, nil
		}
		return 0, nil, err
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
		ContextLength: contextLength,
		Source:        source,
		IsArchived:    false,
		ResetAt:       sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Model:         model,
		Temperature:   temperature,
	}
	if err := p.DB.Create(&session).Error; err != nil {
		return "", err
	}
	return session.Id.String, nil
}

func (p *provider) UpdateSession(id string, updates map[string]any) error {
	var where = map[string]any{"id": id}
	if err := p.DB.Where(where).First(new(models.AIProxySessions)).Error; err != nil {
		return err
	}
	return p.DB.Model(new(models.AIProxySessions)).Where(where).Updates(updates).Error
}

func (p *provider) DeleteSession(id string) error {
	return p.DB.Delete(new(models.AIProxySessions), map[string]any{"id": id}).Error
}

func (p *provider) ListSessions(where map[string]any, pageNum, pageSize int) (int64, []*pb.Session, error) {
	var (
		items []models.AIProxySessions
		count int64
	)
	if err := p.DB.Model(new(models.AIProxySessions)).Where(where).Count(&count).
		Limit(pageSize).Offset((pageNum - 1) * pageSize).
		Find(&items).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, nil
		}
		return 0, nil, err
	}
	var sessions []*pb.Session
	for _, item := range items {
		sessions = append(sessions, item.ToProtobuf())
	}
	return count, sessions, nil
}

func (p *provider) GetSession(id string) (*pb.Session, error) {
	var session models.AIProxySessions
	if err := p.DB.First(&session, map[string]any{"id": id}).Error; err != nil {
		return nil, err
	}
	return session.ToProtobuf(), nil
}
