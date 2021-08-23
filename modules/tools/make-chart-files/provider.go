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

package make_chart_files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysql"
)

type define struct{}

func (d *define) Service() []string      { return []string{"make-chart-files"} }
func (d *define) Dependencies() []string { return []string{"mysql"} }
func (d *define) Summary() string        { return "make chart files from database" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{} {
	return &config{
		Path: "conf/monitor/monitor/charts2",
	}
}
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Path string `file:"path"`
}

type provider struct {
	C  *config
	L  logs.Logger
	db *gorm.DB
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.db = ctx.Service("mysql").(mysql.Interface).DB()
	return nil
}

// ChartMeta .
type ChartMeta struct {
	Name        string `gorm:"column:name"`
	Title       string `gorm:"column:title"`
	MetricsName string `gorm:"column:metricsName"`
	Fields      string `gorm:"column:fields"`
	Parameters  string `gorm:"column:parameters"`
	Type        string `gorm:"column:type"`
	Order       int    `gorm:"column:order"`
	Unit        string `gorm:"column:unit"`
}

// FieldMeta .
type FieldMeta struct {
	Label        string `yaml:"label" json:"label"`
	Unit         string `yaml:"unit" json:"unit"`
	UnitType     string `yaml:"unit_type" json:"unit_type"`
	OriginalUnit string `yaml:"original_unit" json:"original_unit"`
	AxisIndex    int    `yaml:"axis_index" json:"axis_index"`
	ChartType    string `yaml:"chart_type" json:"chart_type"`
	Column       *int   `yaml:"column" json:"column"`
}

// Start .
func (p *provider) Start() error {
	charts := make(map[string]*ChartMeta)
	var list []*ChartMeta
	err := p.db.Table("chart_meta").Find(&list).Error
	if err != nil {
		return err
	}
	toListMap(list, charts)
	list = nil
	err = p.db.Table("sp_chart_meta").Find(&list).Error
	if err != nil {
		return err
	}
	toListMap(list, charts)
	// fmt.Println(jsonx.MarshalAndIndent(charts))

	for name, c := range charts {
		c.Type = strings.TrimSpace(c.Type)
		if c.Type == "0" {
			c.Type = ""
		}
		if strings.HasPrefix(c.Name, "ta_m_") {
			c.Type = "ta_m"
		} else if strings.HasPrefix(c.Name, "ta_") {
			c.Type = "ta"
		} else if strings.HasPrefix(c.Name, "ai_") {
			c.Type = "ai"
		} else if strings.HasPrefix(c.Name, "kong_") {
			c.Type = "kong"
		} else if strings.HasPrefix(c.Name, "gateway_") {
			c.Type = "gateway"
		} else if strings.HasPrefix(c.Name, "alert_") {
			c.Type = "alert"
		}
		file := name + ".yml"
		if len(c.Type) > 0 {
			dir := filepath.Join(p.C.Path, c.Type)
			os.MkdirAll(dir, os.ModePerm)
			file = filepath.Join(dir, file)
		} else {
			file = filepath.Join(p.C.Path, file)
		}
		err := p.writeToFile(file, c)
		if err != nil {
			p.L.Error(err)
		}
	}
	return nil
}

func toListMap(list []*ChartMeta, m map[string]*ChartMeta) {
	for _, item := range list {
		m[item.Name] = item
	}
}

func (p *provider) writeToFile(filename string, c *ChartMeta) error {
	var params map[string]interface{}
	c.Parameters = strings.TrimSpace(c.Parameters)
	if len(c.Parameters) > 0 {
		err := json.Unmarshal([]byte(c.Parameters), &params)
		if err != nil {
			return fmt.Errorf("invalid parameters in %s: %s", c.Name, err)
		}
	}
	var fields map[string]*FieldMeta
	c.Fields = strings.TrimSpace(c.Fields)
	if len(c.Fields) > 0 {
		err := json.Unmarshal([]byte(c.Fields), &fields)
		if err != nil {
			return fmt.Errorf("invalid fields in %s: %s", c.Name, err)
		}
	}
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("fail to create file: %s", err)
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("- %s:\n", strings.TrimSpace(c.Name)))
	file.WriteString(fmt.Sprintf("    title: \"%s\"\n", normalizeString(c.Title)))
	file.WriteString(fmt.Sprintf("    metric_names: \"%s\"\n", normalizeString(c.MetricsName)))
	file.WriteString(fmt.Sprintf("    type: \"%s\"\n", normalizeString(c.Type)))
	file.WriteString(fmt.Sprintf("    order: %d\n", c.Order))
	file.WriteString(fmt.Sprintf("    defines:\n"))
	if fields != nil {
		for key, field := range fields {
			file.WriteString(fmt.Sprintf("      %s:\n", strings.TrimSpace(key)))
			file.WriteString(fmt.Sprintf("        label: \"%s\"\n", normalizeString(field.Label)))
			file.WriteString(fmt.Sprintf("        unit: \"%s\"\n", normalizeString(field.Unit)))
			file.WriteString(fmt.Sprintf("        unit_type: \"%s\"\n", normalizeString(field.UnitType)))
			file.WriteString(fmt.Sprintf("        axis_index: %d\n", field.AxisIndex))
			file.WriteString(fmt.Sprintf("        chart_type: \"%s\"\n", normalizeString(field.ChartType)))
			if len(c.Unit) > 0 && c.Unit != "%" {
				file.WriteString(fmt.Sprintf("        original_unit: \"%s\"\n", normalizeString(c.Unit)))
			}
			if field.Column != nil {
				file.WriteString(fmt.Sprintf("        column: \"%s\"\n", normalizeString(field.ChartType)))
			}
		}
	}
	file.WriteString(fmt.Sprintf("    parameters:\n"))
	if params != nil {
		for key, value := range params {
			switch val := value.(type) {
			case string:
				file.WriteString(fmt.Sprintf("      %s:\n", strings.TrimSpace(key)))
				file.WriteString(fmt.Sprintf("        - \"%s\"\n", normalizeString(val)))
			case []string:
				file.WriteString(fmt.Sprintf("      %s:\n", strings.TrimSpace(key)))
				for _, v := range val {
					file.WriteString(fmt.Sprintf("        - \"%s\"\n", normalizeString(v)))
				}
			default:
				return fmt.Errorf("invalid parameters value type : %v", val)
			}
		}
	}
	file.WriteString("\n")
	return nil
}

func normalizeString(text string) string {
	return strings.Replace(strings.TrimSpace(text), `"`, `\"`, -1)
}

func (p *provider) Close() error {
	return nil
}

func init() {
	servicehub.RegisterProvider("make-chart-files", &define{})
}
