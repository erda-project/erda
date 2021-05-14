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

package chartmeta

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
)

type DataMeta struct {
	Label        *string `yaml:"label" json:"label"`
	Unit         *string `yaml:"unit" json:"unit"`
	UnitType     *string `yaml:"unit_type" json:"unit_type"`
	OriginalUnit *string `yaml:"original_unit" json:"original_unit"`
	AxisIndex    *int    `yaml:"axis_index" json:"axis_index"`
	ChartType    *string `yaml:"chart_type" json:"chart_type"`
	Column       *int    `yaml:"column" json:"column"`
}

type ChartMeta struct {
	Name        string               `yaml:"name" json:"name"`
	Title       string               `yaml:"title" json:"title"`
	MetricNames string               `yaml:"metric_names" json:"metric_names"`
	Defines     map[string]*DataMeta `yaml:"defines" json:"defines"`
	Parameters  map[string][]string  `yaml:"parameters" json:"parameters"`
	Type        string               `yaml:"type" json:"type"`
	// ChartType   string               `yaml:"chart_type" json:"chart_type"`
	Order int `yaml:"order" json:"order"`
}

type Manager struct {
	fixedNameMap map[string]*ChartMeta
	nameMap      map[string]*ChartMeta
	typeMap      map[string][]*ChartMeta
	lock         sync.RWMutex
	reload       time.Duration
	db           *gorm.DB
	files        string
	log          logs.Logger
	t            i18n.Translator
}

func NewManager(db *gorm.DB, reload time.Duration, files string, t i18n.Translator, log logs.Logger) *Manager {
	return &Manager{
		nameMap: make(map[string]*ChartMeta),
		typeMap: make(map[string][]*ChartMeta),
		reload:  reload,
		files:   files,
		db:      db,
		log:     log,
		t:       t,
	}
}

func (m *Manager) Start() error {
	err := m.LoadFiles()
	if err != nil {
		return fmt.Errorf("fail to load chart from files %s", err)
	}
	// err = m.LoadDatabase()
	// if err != nil {
	// 	return fmt.Errorf("fail to load chart from db %s", err)
	// }
	// go func() {
	// 	tick := time.Tick(m.reload)
	// 	for range tick {
	// 		err = m.LoadDatabase()
	// 		if err != nil {
	// 			m.log.Errorf("fail to load chart from db %s", err)
	// 		}
	// 	}
	// }()
	return nil
}

func (m *Manager) Close() error { return nil }

type ChartMetas []*ChartMeta

func (s ChartMetas) Len() int {
	return len(s)
}

func (s ChartMetas) Less(i, j int) bool {
	return s[i].Order < s[j].Order
}

func (s ChartMetas) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (cm *ChartMeta) mergeParams() {
	if cm.Defines == nil {
		return
	}
	if cm.Parameters == nil {
		cm.Parameters = make(map[string][]string)
	}
	for key := range cm.Defines {
		idx := strings.Index(key, ".")
		if idx < 0 {
			continue
		}
		name := key[:idx]
		val := key[idx+1:]
		cm.Parameters[name] = append(cm.Parameters[name], val)
	}
}

func (m *Manager) reloadCharts(nameMap map[string]*ChartMeta, typeMap map[string][]*ChartMeta) {
	if m.fixedNameMap != nil {
		for key, val := range m.fixedNameMap {
			_, ok := nameMap[key]
			if ok {
				continue
			}
			nameMap[key] = val
			typeMap[val.Type] = append(typeMap[val.Type], val)
		}
	}
	for _, list := range typeMap {
		sort.Sort(ChartMetas(list))
	}
	m.lock.Lock()
	m.nameMap = nameMap
	m.typeMap = typeMap
	m.lock.Unlock()
}

func (m *Manager) translator(langCode i18n.LanguageCodes, c *ChartMeta) *ChartMeta {
	cm := *c
	cm.Title = m.t.Text(langCode, cm.Title)
	if c.Defines != nil {
		defines := make(map[string]*DataMeta)
		for key, d := range c.Defines {
			data := *d
			if data.Label != nil {
				label := m.t.Text(langCode, *data.Label)
				data.Label = &label
			}
			if data.Unit != nil {
				unit := m.t.Text(langCode, *data.Unit)
				data.Unit = &unit
			}
			defines[key] = &data
		}
		cm.Defines = defines
	}
	return &cm
}

// ChartMeta .
func (m *Manager) ChartMeta(langCodes i18n.LanguageCodes, name string) *ChartMeta {
	m.lock.RLock()
	cm := m.nameMap[name]
	m.lock.RUnlock()
	if cm == nil {
		return &ChartMeta{
			Name:        name,
			MetricNames: name,
		}
	}
	return m.translator(langCodes, cm)
}

func (m *Manager) ChartMetaList(langCode i18n.LanguageCodes, typ string) []*ChartMeta {
	m.lock.RLock()
	typs := m.typeMap[typ]
	m.lock.RUnlock()
	var list []*ChartMeta
	for _, item := range typs {
		list = append(list, m.translator(langCode, item))
	}
	return list
}
