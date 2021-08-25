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

package make_metric_meta_files

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	metrics "github.com/erda-project/erda/modules/core/monitor/metric"
)

type define struct{}

func (d *define) Service() []string      { return []string{"make-metric-meta-files"} }
func (d *define) Dependencies() []string { return []string{"elasticsearch"} }
func (d *define) Summary() string        { return "make metric meta from _metric_meta" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{} {
	return &config{
		MetaPath: "conf/monitor/monitor/metricmeta/metrics",
		I18nPath: "conf/monitor/monitor/metricmeta/i18n",
	}
}
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	MetaPath string `file:"meta_path"`
	I18nPath string `file:"i18n_path"`
}

type provider struct {
	C   *config
	L   logs.Logger
	es  *elastic.Client
	dic map[string]string
}

func (p *provider) Init(ctx servicehub.Context) error {
	es := ctx.Service("elasticsearch").(elasticsearch.Interface)
	p.es = es.Client()
	p.dic = make(map[string]string)
	return nil
}

// Start .
func (p *provider) Start() error {
	boolQuery := elastic.NewBoolQuery()
	// boolQuery = boolQuery.Filter(elastic.NewTermQuery("tags.metric_name", "status_page"))
	searchSource := elastic.NewSearchSource().Query(boolQuery)
	terms := elastic.NewTermsAggregation().
		Size(5000).
		Field("tags.metric_name").
		SubAggregation("top", elastic.NewTopHitsAggregation().Size(5).Sort("timestamp", false))
	searchSource.Aggregation("metrics", terms)
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := p.es.Search("spot-_metric_meta-*").
		IgnoreUnavailable(true).AllowNoIndices(true).
		SearchSource(searchSource).Do(context)
	if err != nil || (resp != nil && resp.Error != nil) {
		if resp != nil && resp.Error != nil {
			return fmt.Errorf("fail to request elasticsearch: %s", jsonx.MarshalAndIndent(resp.Error))
		}
		return fmt.Errorf("fail to request elasticsearch: %s", err)
	}
	if resp == nil || resp.Aggregations == nil {
		p.L.Warnf("elasticsearch response is empty")
		return nil
	}
	metricTerms, ok := resp.Aggregations.Terms("metrics")
	if !ok || metricTerms == nil {
		p.L.Warnf("elasticsearch Aggregations '%s' is Terms", "metrics")
		return nil
	}
	for _, b := range metricTerms.Buckets {
		if b.Aggregations == nil {
			continue
		}
		tophits, ok := b.Aggregations.TopHits("top")
		if !ok || tophits == nil || tophits.Hits == nil {
			continue
		}
		metric := metrics.Metric{
			Tags:   make(map[string]string),
			Fields: make(map[string]interface{}),
		}
		for _, h := range tophits.Hits.Hits {
			if h.Source == nil {
				continue
			}
			var m metrics.Metric
			err := json.Unmarshal(*h.Source, &m)
			if err != nil {
				p.L.Warnf("fail to Unmarshal source: %s", err)
				continue
			}
			metric.Name = m.Name
			metric.Timestamp = m.Timestamp
			for t, v := range m.Tags {
				_, ok := metric.Tags[t]
				if ok {
					continue
				}
				metric.Tags[t] = v
			}
			for f, v := range m.Fields {
				fv, ok := metric.Fields[f]
				if ok {
					fv, ok = mergeSlice(fv, v)
					if ok {
						metric.Fields[f] = fv
					} else {
						p.L.Warnf("fail to mergeSlice %s=%v", f, fv)
					}
					continue
				}
				metric.Fields[f] = v
			}
		}
		p.createMetricMetaFile(&metric)
	}
	// p.createMetricMetaFile(commonMetric)
	p.makeI18nFile("_common", zhTrans)
	return nil
}

func mergeSlice(a, b interface{}) (interface{}, bool) {
	as, ok := a.([]interface{})
	if !ok {
		return a, false
	}
	bs, ok := b.([]interface{})
	if !ok {
		return b, false
	}
	m := make(map[string]bool)
	for _, item := range as {
		key := fmt.Sprint(item)
		m[key] = true
	}
	for _, item := range bs {
		key := fmt.Sprint(item)
		m[key] = true
	}
	var list []interface{}
	for k := range m {
		list = append(list, k)
	}
	return list, true
}

func (p *provider) createMetricMetaFile(m *metrics.Metric) {
	name, ok := m.Tags["metric_name"]
	if !ok {
		p.L.Warnf("fail to read metric_name: %v", m)
		return
	}
	scope, ok := m.Tags["metric_scope"]
	if !ok {
		p.L.Warnf("fail to read metric_scope: %v", m)
	}
	if strings.HasPrefix(name, "docker_container_") && name != "docker_container_summary" {
		return
	}
	if name == "mem" || name == "cpu" || name == "system" || name == "swap" {
		return
	}
	fields, ok := m.Fields["fields"].([]interface{})
	if !ok {
		p.L.Warnf("fail to read fields: %v", m)
		return
	}
	sort.Slice(fields, func(i, j int) bool {
		f1, f2 := fields[i].(string), fields[j].(string)
		return f1 < f2
	})
	tags, ok := m.Fields["tags"].([]interface{})
	if !ok {
		p.L.Warnf("fail to read tags: %v", m)
		return
	}
	sort.Slice(tags, func(i, j int) bool {
		t1, t2 := tags[i].(string), tags[j].(string)
		return t1 < t2
	})
	var outfile string
	os.MkdirAll(p.C.MetaPath, os.ModePerm)
	if len(scope) > 0 {
		os.MkdirAll(path.Join(p.C.MetaPath, scope), os.ModePerm)
		outfile = path.Join(p.C.MetaPath, scope, name+".yml")
	} else {
		outfile = path.Join(p.C.MetaPath, name+".yml")
	}
	file, err := os.Create(outfile)
	if err != nil {
		p.L.Warnf("fail to create file: %s", err)
		return
	}
	defer file.Close()
	dic := make(map[string]string)
	file.WriteString(fmt.Sprintf("name: \"%s\"\n", p.normalizeName(dic, name)))
	file.WriteString("tags:\n")
	for _, t := range tags {
		k := fmt.Sprint(t)
		if p.skipName(name, k) {
			continue
		}
		// if isCommonTags(k) {
		// 	continue
		// }
		file.WriteString(fmt.Sprintf("    %s:\n", t))
		file.WriteString(fmt.Sprintf("        name: \"%s\"\n", p.normalizeName(dic, k)))
	}
	file.WriteString("fields:\n")
	for _, f := range fields {
		k := fmt.Sprint(f)
		idx := strings.LastIndex(k, ":")
		if idx <= 0 {
			continue
		}
		typ := k[idx+1:]
		if typ != "number" && typ != "string" && typ != "bool" {
			p.L.Warnf("invalid field type '%s'", typ)
		}
		k = k[0:idx]
		if p.skipName(name, k) {
			continue
		}
		file.WriteString(fmt.Sprintf("    %s:\n", k))
		file.WriteString(fmt.Sprintf("        type: \"%s\"\n", typ))
		file.WriteString(fmt.Sprintf("        name: \"%s\"\n", p.normalizeName(dic, k)))
		file.WriteString(fmt.Sprintf("        uint: \n"))
	}
	file.WriteString("\n")
	p.makeI18nFile(name, dic)
	// fmt.Println(jsonx.MarshalAndIndent(m))
}

func (p *provider) skipName(metric, name string) bool {
	if len(name) <= 0 {
		return true
	}
	if strings.HasPrefix(name, "_") {
		return true
	}
	if unicode.IsDigit((([]rune)(name))[0]) {
		p.L.Warnf("invalid field name '%s' in metric '%s'", name, metric)
		return true
	}
	return false
}

func (p *provider) normalizeName(dic map[string]string, text string) string {
	text = normalizeName(text)
	dic[text] = text
	return text
}

func (p *provider) makeI18nFile(name string, dic map[string]string) {
	var keys []string
	for k := range dic {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	os.MkdirAll(p.C.I18nPath, os.ModePerm)
	outfile := path.Join(p.C.I18nPath, name+".yml")
	file, err := os.Create(outfile)
	if err != nil {
		p.L.Warnf("fail to create file: %s", err)
		return
	}
	defer file.Close()
	file.WriteString("zh:\n")
	for _, k := range keys {
		if w, ok := zhTrans[k]; ok {
			file.WriteString(fmt.Sprintf("    \"%s\": \"%s\"\n", k, w))
			continue
		}
		file.WriteString(fmt.Sprintf("    \"%s\": \"%s\"\n", k, k))
	}
}

func normalizeName(text string) string {
	var parts []string
	for _, name := range strings.Split(text, "_") {
		names := []rune(name)
		j, i, flag, l := 0, 0, 0, len(names)
		for ; i < l; i++ {
			c := names[i]
			if unicode.IsUpper(c) {
				if flag == 0 {
					flag = 2
					continue
				}
				if flag >= 2 {
					flag++
					continue
				}
				parts = append(parts, toTitle(string(names[j:i])))
				j = i
				flag = 2
			} else {
				if flag > 2 {
					parts = append(parts, toTitle(string(names[j:i-1])))
					j = i - 1
				}
				flag = 1
			}
		}
		if j < l {
			parts = append(parts, toTitle(string(names[j:])))
		}
	}
	return strings.Join(parts, " ")
}

func toTitle(name string) string {
	switch strings.ToLower(name) {
	case "api", "id", "ip", "uid", "http", "sls", "waf", "ua", "os":
		return strings.ToUpper(name)
	case "max":
		return "Maximum"
	case "min":
		return "Minimum"
	}
	return strings.Title(name)
}

var zhTrans = map[string]string{
	"Cluster Name":   "集群名",
	"Host IP":        "机器IP",
	"Org Name":       "企业名",
	"Org ID":         "企业ID",
	"Addon ID":       "中间件ID",
	"Project ID":     "项目ID",
	"Application ID": "应用ID",
	"Service Name":   "服务名",
	"Host":           "主机",
	"Container ID":   "容器ID",
	"Pod ID":         "Pod ID",
	"Pod Name":       "Pod Name",
}

var commonTags = []interface{}{
	"cluster_name",
	"host_ip",
	"org_name",
	"org_id",
}

func isCommonTags(tag string) bool {
	for _, t := range commonTags {
		if tag == t {
			return true
		}
	}
	return false
}

var commonMetric = &metrics.Metric{
	Fields: map[string]interface{}{
		"tags":   commonTags,
		"fields": []interface{}{},
	},
	Tags: map[string]string{
		"metric_name": "_common",
	},
}

func (p *provider) Close() error {
	return nil
}

func init() {
	servicehub.RegisterProvider("make-metric-meta-files", &define{})
}
