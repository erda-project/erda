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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"
	yaml "gopkg.in/yaml.v2"
)

func (m *Manager) LoadFiles() error {
	if m.files == "" {
		return nil
	}
	type ChartInfo struct {
		Title       string               `yaml:"title"`
		MetricNames string               `yaml:"metric_names"`
		Defines     map[string]*DataMeta `yaml:"defines"`
		Default     *DataMeta            `yaml:"default"`
		Parameters  map[string][]string  `yaml:"parameters"`
		Type        string               `yaml:"type"`
		// ChartType   string               `yaml:"chart_type"`
		Order *int `yaml:"order"`
	}
	nameMap := make(map[string]*ChartMeta)
	typeMap := make(map[string][]*ChartMeta)
	if m.files == "/" {
		return fmt.Errorf("invalid charts path")
	}
	err := filepath.Walk(m.files, func(p string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() || err != nil {
			return nil
		}
		if strings.Contains(p, "vendor") || strings.Contains(p, ".git") {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".yml") && !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}
		_, file := filepath.Split(p)
		typ := file[:strings.LastIndex(file, ".")]
		var list []map[string]*ChartInfo
		err = readFile(p, &list)
		if err != nil {
			return err
		}
		for _, infos := range list {
			var order int
			for name, info := range infos {
				if _, ok := nameMap[name]; ok {
					return fmt.Errorf("name %s conflicts", name)
				}
				if info.Order != nil {
					order = *info.Order
				}
				// 兼容处理
				// for _, item := range info.Defines {
				// 	item.ChartType = info.ChartType
				// }
				cm := &ChartMeta{
					Name:        name,
					Title:       info.Title,
					MetricNames: info.MetricNames,
					Defines:     info.Defines,
					Parameters:  info.Parameters,
					// ChartType:   info.ChartType,
					Type:  info.Type,
					Order: order,
				}
				if len(info.Type) <= 0 {
					cm.Type = typ
				}
				order++
				if cm.Defines != nil && info.Default != nil {
					for _, d := range cm.Defines {
						if d.OriginalUnit == nil {
							d.OriginalUnit = info.Default.OriginalUnit
						}
						if d.Unit == nil {
							d.Unit = info.Default.Unit
						}
						if d.UnitType == nil {
							d.UnitType = info.Default.UnitType
						}
						if d.Label == nil {
							d.Label = info.Default.Label
						}
						if d.AxisIndex == nil {
							d.AxisIndex = info.Default.AxisIndex
						}
						if d.ChartType == nil {
							d.ChartType = info.Default.ChartType
						}
					}
				}
				cm.mergeParams()
				nameMap[name] = cm
				typeMap[cm.Type] = append(typeMap[cm.Type], cm)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	m.fixedNameMap = nameMap
	m.reloadCharts(nameMap, typeMap)
	return nil
}

func readFile(file string, out interface{}) error {
	exts := []string{"json", "yaml", "yml", "toml"}
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
	"toml": toml.Unmarshal,
}
