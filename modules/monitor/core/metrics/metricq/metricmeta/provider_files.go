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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
	"gopkg.in/yaml.v2"
)

// scopeDefine
type scopeGroupDefine struct {
	Groups       []*Group
	GroupsMap    map[string]*Group
	GroupMetrics map[string][]*GroupMetricMap
}

// FileGroupProvider .
type FileGroupProvider struct {
	scopes map[string]*scopeGroupDefine
	log    logs.Logger
}

// NewFileGroupProvider .
func NewFileGroupProvider(files []string, log logs.Logger) (*FileGroupProvider, error) {
	p := &FileGroupProvider{
		scopes: make(map[string]*scopeGroupDefine),
		log:    log,
	}
	err := p.loadMetricGroupFromFile(files)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *FileGroupProvider) loadMetricGroupFromFile(files []string) error {
	type groupsMappingDefine struct {
		Groups  []*Group                     `json:"groups" yaml:"groups"`
		Mapping map[string][]*GroupMetricMap `json:"mapping" yaml:"mapping"`
	}
	for _, f := range files {
		base := filepath.Base(f)
		name := base[0 : len(base)-len(filepath.Ext(base))]
		var define groupsMappingDefine
		err := readFile(f, &define)
		if err != nil {
			return err
		}
		sd := p.scopes[name]
		if sd == nil {
			sd = &scopeGroupDefine{
				GroupsMap:    make(map[string]*Group),
				GroupMetrics: make(map[string][]*GroupMetricMap),
			}
			p.scopes[name] = sd
		}
		sd.Groups = define.Groups
		if define.Mapping != nil {
			define.Mapping = initGroupMapping(define.Mapping)
			sd.GroupMetrics = define.Mapping
		}
		p.loadGroup(sd, "", define.Groups, define.Mapping)
	}
	return nil
}

func initGroupMapping(mapping map[string][]*GroupMetricMap) map[string][]*GroupMetricMap {
	for _, list := range mapping {
		for _, item := range list {
			for _, f := range item.Filters {
				if len(f.Op) <= 0 {
					f.Op = "eq"
				}
			}
		}
	}
	return mapping
}

func (p *FileGroupProvider) loadGroup(sd *scopeGroupDefine, prefix string, groups []*Group, mapping map[string][]*GroupMetricMap) {
	for _, g := range groups {
		if !strings.HasPrefix(g.ID, prefix) {
			p.log.Warnf("group %s is not starts with %s", g.ID, prefix)
		}
		if len(g.ID) <= 0 {
			p.log.Warnf("group id is empty with prefix %s", prefix)
			continue
		}
		sd.GroupsMap[g.ID] = g
		p.loadGroup(sd, g.ID, g.Children, mapping)
	}
}

// MappingsByID .
func (p *FileGroupProvider) MappingsByID(id, scope, scopeID string, metric []string, ms map[string]*metrics.MetricMeta) (gmm []*GroupMetricMap, err error) {
	sd := p.scopes[scope]
	if sd == nil {
		return gmm, nil
	}
	return sd.GroupMetrics[id], nil
}

// Groups .
func (p *FileGroupProvider) Groups(langCodes i18n.LanguageCodes, t i18n.Translator, scope, scopeID string, ms map[string]*metrics.MetricMeta) (groups []*Group, err error) {
	sd := p.scopes[scope]
	if sd == nil {
		return nil, nil
	}
	groups = copyMetricGroupsWithLang(langCodes, t, sd.Groups)
	groups = appendMetricToGroup(groups, "@", ms, sd.GroupMetrics, false)
	return groups, nil
}

func copyMetricGroupsWithLang(langCodes i18n.LanguageCodes, t i18n.Translator, groups []*Group) []*Group {
	var list []*Group
	for _, g := range groups {
		ng := &Group{
			ID:   g.ID,
			Name: t.Text(langCodes, g.Name),
		}
		ng.Children = copyMetricGroupsWithLang(langCodes, t, g.Children)
		list = append(list, ng)
	}
	return list
}

// FileMetricMetaProvider .
type FileMetricMetaProvider struct {
	scopes  map[string]map[string]*metrics.MetricMeta
	metrics map[string]*metrics.MetricMeta
	log     logs.Logger
}

// NewFileMetricMetaProvider .
func NewFileMetricMetaProvider(path string, log logs.Logger) (*FileMetricMetaProvider, error) {
	p := &FileMetricMetaProvider{
		scopes:  make(map[string]map[string]*metrics.MetricMeta),
		metrics: make(map[string]*metrics.MetricMeta),
		log:     log,
	}
	err := p.loadMetricMetaFromFile(path)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type metricMetaDefine struct {
	Name   string                          `json:"name" yaml:"name"`
	Tags   map[string]*metrics.TagDefine   `json:"tags" yaml:"tags"`
	Fields map[string]*metrics.FieldDefine `json:"fields" yaml:"fields"`
}

func (p *FileMetricMetaProvider) loadMetricMetaFromFile(path string) error {
	absPath, _ := filepath.Abs(path)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() || err != nil {
			return nil
		}
		if strings.Contains(path, "vendor") || strings.Contains(path, ".git") {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".yml") && !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}
		dir, file := filepath.Split(path)
		name := file[:strings.LastIndex(file, ".")]
		dir = strings.TrimRight(dir, "/")
		_, scope := filepath.Split(dir)
		dirPath, _ := filepath.Abs(dir)
		if dirPath == absPath {
			scope = ""
		}
		var md metricMetaDefine
		err = readFile(path, &md)
		if err != nil {
			return err
		}
		meta := convertMetricMeta(name, &md)
		p.metrics[name] = meta
		if len(scope) > 0 {
			sm := p.scopes[scope]
			if sm == nil {
				sm = make(map[string]*metrics.MetricMeta)
				p.scopes[scope] = sm
			}
			sm[name] = meta
		}
		return nil
	})
	if err != nil {
		return err
	}
	p.log.Infof("load %d metric meta from file", len(p.metrics))
	return nil
}

func convertMetricMeta(name string, md *metricMetaDefine) *metrics.MetricMeta {
	var meta metrics.MetricMeta
	meta.Name.Key = name
	meta.Name.Name = md.Name
	if md.Tags != nil {
		for key, td := range md.Tags {
			if td == nil {
				td = &metrics.TagDefine{}
				md.Tags[key] = td
			}
			td.Key = key
			if len(td.Name) <= 0 {
				td.Name = key
			}
		}
		meta.Tags = md.Tags
	} else {
		meta.Tags = make(map[string]*metrics.TagDefine)
	}
	if md.Fields != nil {
		for key, fd := range md.Fields {
			if fd == nil {
				fd = &metrics.FieldDefine{}
				md.Fields[key] = fd
			}
			fd.Key = key
			if len(fd.Name) <= 0 {
				fd.Name = key
			}
		}
		meta.Fields = md.Fields
	} else {
		meta.Fields = make(map[string]*metrics.FieldDefine)
	}
	return &meta
}

// MetricMeta .
func (p *FileMetricMetaProvider) MetricMeta(langCodes i18n.LanguageCodes, i i18n.I18n, scope, scopeID string, names ...string) (map[string]*metrics.MetricMeta, error) {
	namesMap := make(map[string]bool)
	for _, name := range names {
		namesMap[name] = true
	}
	metas := make(map[string]*metrics.MetricMeta)
	for _, scope := range strings.Split(scope, ",") {
		ms := p.scopes[scope]
		if ms != nil {
			for key, item := range ms {
				if len(namesMap) > 0 && !namesMap[key] {
					continue
				}
				item = copyMetricMeta(item)
				metas[key] = item
			}
		}
	}
	return transMetricMetas(langCodes, i, metas), nil
}

func readFile(file string, out interface{}) error {
	exts := []string{"json", "yaml", "yml"}
	for _, ext := range exts {
		if !strings.HasSuffix(file, ext) {
			continue
		}
		reader, ok := fileReaders[ext]
		if !ok {
			return fmt.Errorf("not exit %s file reader", ext)
		}
		_, err := os.Stat(file)
		if err != nil {
			continue
		}
		byts, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		err = reader(byts, out)
		if err != nil {
			return fmt.Errorf("fail to Unmarshal %s: %s", file, err)
		}
		return nil
	}
	return fmt.Errorf("not exit file reader")
}

var fileReaders = map[string]func([]byte, interface{}) error{
	"json": json.Unmarshal,
	"yaml": yaml.Unmarshal,
	"yml":  yaml.Unmarshal,
}
