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

package metrics

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
)

type Collector struct {
	l    logs.Logger
	db   *gorm.DB
	desc *prometheus.Desc
}

func SingletonCollector(db *gorm.DB, l logs.Logger) prometheus.Collector {
	return &Collector{
		l:    l,
		db:   db,
		desc: prometheus.NewDesc("historical_requests", "Total number of ai-proxy requested", new(LabelValues).Labels(), nil),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var audits []models.AIProxyFilterAudit
	if err := c.db.Find(&audits).Error; err != nil {
		c.l.Errorf("failed to db.Find(%T), err: %v", audits, err)
		audits = append(audits, mockedAudit, mockedAudit)
	}

	var m = make(map[string]*lvsDistincter)
	for _, item := range audits {
		lv := LabelValues{
			ChatType:    item.ChatType,
			ChatTitle:   item.ChatTitle,
			Source:      item.Source,
			UserId:      item.JobNumber,
			UserName:    item.Username,
			Provider:    item.Provider,
			Model:       item.Model,
			OperationId: item.OperationId,
			Status:      item.Status,
			StatusCode:  strconv.FormatInt(int64(item.StatusCode), 10),
		}
		key := strings.Join(lv.Values(), "ʕ◔ϖ◔ʔ")
		if value, ok := m[key]; ok {
			value.Count++
		} else {
			m[key] = &lvsDistincter{LVs: lv, Count: 1}
		}
	}
	for _, value := range m {
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, value.Count, value.LVs.Values()...)
	}
}

type lvsDistincter struct {
	LVs   LabelValues
	Count float64
}

var mockedAudit = models.AIProxyFilterAudit{
	ChatType:    "mocked",
	ChatTitle:   "mocked",
	Source:      "mocked",
	JobNumber:   "mocked",
	Username:    "mocked",
	Provider:    "mocked",
	Model:       "mocked",
	OperationId: "mocked",
	Status:      "mocked",
	StatusCode:  200,
}
