// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package metricmeta

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/jinzhu/gorm"
)

type MetaSource string

// MetaSource values
const (
	MappingsMetaSource MetaSource = "mappings"
	IndexMetaSource    MetaSource = "index"
	FileMetaSource     MetaSource = "file"
	DatabaseMetaSource MetaSource = "database"
)

type Manager struct {
	sources         map[MetaSource]bool
	groupProviders  []func() GroupProvider
	metricProviders []func() MetricMetaProvider

	db         *gorm.DB
	index      indexmanager.Index
	metaPath   string
	groupFiles []string

	i18n i18n.I18n
	log  logs.Logger
}

func NewManager(ms []string, db *gorm.DB, index indexmanager.Index, metaPath string, groupFiles []string, i18n i18n.I18n, log logs.Logger) *Manager {
	sources := make(map[MetaSource]bool)
	for _, item := range ms {
		sources[MetaSource(item)] = true
	}
	return &Manager{
		sources:    sources,
		db:         db,
		index:      index,
		metaPath:   metaPath,
		groupFiles: groupFiles,
		i18n:       i18n,
		log:        log,
	}
}

func (m *Manager) Start() error {
	if m.sources[MappingsMetaSource] {
		gp, err := NewIndexMappingsGroupProvider(m.index)
		if err != nil {
			return err
		}
		m.groupProviders = append(m.groupProviders, func() GroupProvider { return gp })
		mp, err := NewIndexMappingsMetricMetaProvider(m.index)
		if err != nil {
			return err
		}
		m.metricProviders = append(m.metricProviders, func() MetricMetaProvider { return mp })
	}
	if m.sources[IndexMetaSource] {
		gp, err := NewMetaIndexGroupProvider(m.index)
		if err != nil {
			return err
		}
		m.groupProviders = append(m.groupProviders, func() GroupProvider { return gp })
		mp, err := NewMetaIndexMetricMetaProvider(m.index, m.log)
		if err != nil {
			return err
		}
		m.metricProviders = append(m.metricProviders, func() MetricMetaProvider { return mp })
	}
	if m.sources[FileMetaSource] {
		gp, err := NewFileGroupProvider(m.groupFiles, m.log)
		if err != nil {
			return err
		}
		m.groupProviders = append(m.groupProviders, func() GroupProvider { return gp })
		mp, err := NewFileMetricMetaProvider(m.metaPath, m.log)
		if err != nil {
			return err
		}
		m.metricProviders = append(m.metricProviders, func() MetricMetaProvider { return mp })
	}
	if m.sources[DatabaseMetaSource] {
		gp, err := NewDatabaseGroupProvider(m.db, m.log)
		if err != nil {
			return err
		}
		m.groupProviders = append(m.groupProviders, func() GroupProvider { return gp })
		mp, err := NewDatabaseMetaProvider(m.db, m.log)
		if err != nil {
			return err
		}
		m.metricProviders = append(m.metricProviders, func() MetricMetaProvider { return mp })
	}
	return nil
}

func (m *Manager) Close() error { return nil }
