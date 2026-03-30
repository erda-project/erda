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

package archive

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda-infra/base/logs"
	auditmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	eventmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/event"
)

type ListRequest struct {
	PageNum  uint64
	PageSize uint64
	Day      string
}

type Service struct {
	Config      Config
	EventClient *eventmodel.DBClient
	AuditClient *auditmodel.DBClient
	Logger      logs.Logger
}

func (s *Service) SetStart(ctx context.Context, value bool) error {
	if !s.Config.Enable {
		return fmt.Errorf("audit archive is disabled")
	}
	_, err := s.EventClient.Create(ctx, EventArchiveStart, strconv.FormatBool(value))
	return err
}

func (s *Service) GetStatus(ctx context.Context) (Status, error) {
	if !s.Config.Enable {
		return buildStatus(s.Config, nil), nil
	}
	latestStart, err := s.EventClient.LatestByEvent(ctx, EventArchiveStart)
	if err != nil {
		return Status{}, err
	}
	latestDayStart, err := s.EventClient.LatestByEvent(ctx, EventArchiveDayStart)
	if err != nil {
		return Status{}, err
	}
	latestDayEnd, err := s.EventClient.LatestByEvent(ctx, EventArchiveDayEnd)
	if err != nil {
		return Status{}, err
	}
	latestResult, err := s.EventClient.LatestByEvents(ctx, EventArchiveDaySuccess, EventArchiveDayFailed, EventArchiveDayInterrupted)
	if err != nil {
		return Status{}, err
	}
	return buildStatus(s.Config, latestStart, latestDayStart, latestDayEnd, latestResult), nil
}

func (s *Service) ListEvents(ctx context.Context, req ListRequest) (int64, eventmodel.Events, error) {
	if !s.Config.Enable {
		return 0, nil, fmt.Errorf("audit archive is disabled")
	}
	return s.EventClient.ListDayEvents(ctx, eventmodel.ListOptions{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Day:      req.Day,
	})
}

func NewService(cfg Config, eventClient *eventmodel.DBClient, auditClient *auditmodel.DBClient, logger logs.Logger) *Service {
	return &Service{
		Config:      cfg,
		EventClient: eventClient,
		AuditClient: auditClient,
		Logger:      logger,
	}
}
