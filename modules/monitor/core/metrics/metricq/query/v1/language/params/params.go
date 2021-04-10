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

package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/chartmeta"
	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
	query "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query"
	queryv1 "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1"
	"github.com/recallsong/go-utils/reflectx"
)

// Parser .
type Parser struct{}

// Parse .
func (p *Parser) Parse(statement string) (*queryv1.Request, error) {
	statement = strings.Replace(statement, "\n", "", -1)
	statement = strings.Replace(statement, " ", "", -1)
	statement = strings.Replace(statement, "\t", "", -1)
	parts := strings.SplitN(statement, "?", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid query statement: %s", statement)
	}
	params, err := url.ParseQuery(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid query statement: %s", statement)
	}
	mparts := strings.SplitN(parts[0], "/", 2)
	req := &queryv1.Request{
		ExistKeys:        make(map[string]struct{}),
		LegendMap:        make(map[string]*chartmeta.DataMeta),
		ChartType:        params.Get("chartType"),
		TimeKey:          tsql.TimestampKey,
		OriginalTimeUnit: tsql.Nanosecond,
	}
	err = req.InitTimestamp(params.Get("start"), params.Get("end"), params.Get("timestamp"), params.Get("latest"))
	if err != nil {
		return nil, err
	}

	tf := params.Get("time_field")
	if len(tf) > 0 {
		req.TimeKey = tf
		req.OriginalTimeUnit = tsql.Millisecond
		tu := params.Get("time_unit")
		if len(tu) > 0 {
			unit, err := tsql.ParseTimeUnit(tu)
			if err != nil {
				return nil, err
			}
			req.OriginalTimeUnit = unit
		}
	}

	req.Name = mparts[0]
	req.Metrics = strings.Split(mparts[0], ",")
	req.Where, _ = query.ParseFilters(params)
	for key, vals := range params {
		if key == "group" {
			for _, val := range vals {
				var script string
				if val, ok := getScript(val); ok {
					script, _, err = parseScript(val, query.TagKey)
					if err != nil {
						return nil, fmt.Errorf("invalid script %s", val)
					}
				}
				req.GroupBy = append(req.GroupBy, &queryv1.Group{
					Property: queryv1.Property{
						Name:   val,
						Script: script,
					},
					Limit: 20,
				})
			}
		} else if key == "limit" {
			for _, val := range vals {
				limit, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("invalid limit %s", val)
				}
				req.Limit = append(req.Limit, limit)
			}
		} else if key == "sort" {
			for _, val := range vals {
				sort := "DESC"
				if strings.HasPrefix(val, "histogram_") {
					val = val[len("histogram_"):]
				}
				idx := strings.LastIndex(val, ":")
				if idx >= 0 {
					sort = val[idx+1:]
					val = val[:idx]
				}
				if val == "count" {
					req.OrderBy = append(req.OrderBy, &queryv1.Order{
						FuncName: val,
						Sort:     sort,
					})
					continue
				}
				preIdx := strings.Index(val, ".")
				idx = strings.Index(val, "_")
				if idx < 0 || (preIdx >= 0 && idx > preIdx) {
					if val == "timestamp" {
						val = "." + val
					}
					req.OrderBy = append(req.OrderBy, &queryv1.Order{
						Property: queryv1.Property{
							Name: val,
						},
						FuncName: "",
						Sort:     sort,
					})
				} else {
					if val[idx+1:] == "" || val[:idx] == "" {
						return nil, fmt.Errorf("invalid sort %s", val)
					}
					req.OrderBy = append(req.OrderBy, &queryv1.Order{
						Property: queryv1.Property{
							Name: val[idx+1:],
						},
						FuncName: val[:idx],
						Sort:     sort,
					})
				}
			}
		} else if key == "range" {
			for _, val := range vals {
				var script string
				if s, ok := getScript(val); ok {
					var keys map[string]struct{}
					script, keys, err = parseScript(s, query.FieldKey)
					if err != nil {
						return nil, fmt.Errorf("invalid script %s", val)
					}
					if keys != nil {
						for k := range keys {
							req.ExistKeys[k] = struct{}{}
						}
					}
				}
				values, err := getRangeParams(params)
				if err != nil {
					return nil, err
				}

				req.Select = append(req.Select, &queryv1.Column{
					Property: queryv1.Property{
						Name:   val,
						Script: script,
					},
					FuncName: key,
					Params:   values,
				})
			}
		} else if strings.HasPrefix(key, "alias_") { // 数据别名解析。例子，alias_last.tags.host_ip=主机IP
			var legend string
			if idx := strings.Index(key, "_"); idx == -1 {
				continue
			} else {
				legend = key[idx+1:]
			}
			if len(vals) < 1 {
				continue
			}

			if v, ok := req.LegendMap[legend]; ok {
				v.Label = &vals[len(vals)-1]
				req.LegendMap[legend] = v
			} else {
				req.LegendMap[legend] = &chartmeta.DataMeta{Label: &vals[len(vals)-1]}
			}
		} else if key == "columns" { // 数据别名排序。例子，columns=last.tags.host_ip,last.tags.host_ip,last.tags.host_ip
			if len(vals) < 1 {
				continue
			}
			ss := strings.Split(vals[len(vals)-1], ",")
			for i := 0; i < len(ss); i++ {
				legend, col := ss[i], i
				if v, ok := req.LegendMap[legend]; ok {
					v.Column = &col
					req.LegendMap[legend] = v
				} else {
					req.LegendMap[legend] = &chartmeta.DataMeta{Column: &col}
				}
			}
		} else {
			if _, ok := queryv1.Functions[key]; ok {
				for _, val := range vals {
					var script string
					if s, ok := getScript(val); ok {
						var keys map[string]struct{}
						script, keys, err = parseScript(s, query.FieldKey)
						if err != nil {
							return nil, fmt.Errorf("invalid script %s", val)
						}
						if keys != nil {
							for k := range keys {
								req.ExistKeys[k] = struct{}{}
							}
						}
					}
					req.Select = append(req.Select, &queryv1.Column{
						Property: queryv1.Property{
							Name:   val,
							Script: script,
						},
						FuncName: key,
					})
				}
			}
		}
	}

	if len(req.GroupBy) > 0 {
		for key, vals := range params {
			if strings.HasPrefix(key, "gfilter_") {
				key = key[len("gfilter_"):]
				idx := strings.Index(key, "_")
				if idx > 0 {
					gf := &queryv1.GroupFilter{
						Operator: key[:idx],
					}
					key = key[idx+1:]
					idx = strings.Index(key, "_")
					if idx > 0 {
						gf.FuncName = key[:idx]
						gf.Property.Name = key[idx+1:]
					}
					if len(gf.Operator) == 0 || len(gf.Property.Name) == 0 || len(gf.FuncName) == 0 {
						continue
					}
					for _, val := range vals {
						var value interface{}
						err := json.Unmarshal([]byte(val), &value)
						if err != nil {
							return nil, fmt.Errorf("invalid group filter value %s", val)
						}
						gf.Value = value
						req.GroupBy[len(req.GroupBy)-1].Filters = append(req.GroupBy[len(req.GroupBy)-1].Filters, gf)
						break
					}
				}
			}
		}
	}

	align := strings.ToLower(params.Get("align"))
	if align == "true" {
		req.TimeAlign = queryv1.TimeAlignAuto
	} else if align == "false" {
		req.TimeAlign = queryv1.TimeAlignNone
	} else {
		req.TimeAlign = queryv1.TimeAlignUnset
	}
	req.Trans = len(params.Get("trans")) > 0
	req.TransGroup = len(params.Get("trans_group")) > 0

	if params.Get("defaultNullValue") != "" {
		v, err := strconv.ParseInt(params.Get("defaultNullValue"), 10, 64)
		if err == nil {
			req.DefaultNullValue = v
		}
	}

	if len(mparts) > 1 && req.ChartType != "chart:bar" {
		switch mparts[1] {
		case "histogram":
			req.Aggregate = &queryv1.Column{
				FuncName: mparts[1],
			}
			if params.Get("points") == "" {
				req.Points = -1
				if params.Get("interval") != "" {
					d, err := time.ParseDuration(params.Get("interval"))
					if err != nil {
						return nil, fmt.Errorf("invalid interval %s", params.Get("interval"))
					}
					req.Interval = float64(d)
				}
			} else {
				points, err := strconv.ParseFloat(params.Get("points"), 64)
				if err != nil {
					return nil, fmt.Errorf("invalid points %s", params.Get("points"))
				}
				req.Points = points
			}
			req.AlignEnd = req.TimeAlign != queryv1.TimeAlignNone
		case "range":
		case "apdex":
		}
	}
	if val, ok := params["debug"]; ok {
		if len(val) > 0 && val[0] != "" {
			req.Debug, err = strconv.ParseBool(val[0])
			if err != nil {
				return nil, fmt.Errorf("invalid debug value: %s", val[0])
			}
		} else {
			req.Debug = true
		}
	}
	return req, nil
}

func getRangeParams(params url.Values) (values []interface{}, err error) {
	val := params.Get("ranges")
	if len(val) > 0 {
		ranges := strings.Split(val, ",")
		for _, item := range ranges {
			var from, to interface{}
			vals := strings.SplitN(item, ":", 2)
			if len(vals[0]) > 0 {
				from, err = strconv.ParseFloat(vals[0], 64)
				if err != nil {
					return nil, fmt.Errorf("invalid ranges %s", val)
				}
			}
			if len(vals) > 1 && len(vals[1]) > 0 {
				to, err = strconv.ParseFloat(vals[1], 64)
				if err != nil {
					return nil, fmt.Errorf("invalid ranges %s", val)
				}
			}
			values = append(values, &queryv1.ValueRange{
				From: from,
				To:   to,
			})
		}
	} else {
		val = params.Get("rangeSize")
		rangeSize, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid rangeSize %s", val)
		}
		val = params.Get("split")
		split, err := strconv.Atoi(val)
		if err != nil || split > 1000 {
			return nil, fmt.Errorf("invalid split %s", val)
		}
		for i, last := 0, split-1; i < split; i++ {
			if i >= last {
				values = append(values, &queryv1.ValueRange{
					From: float64(i) * rangeSize,
					To:   nil,
				})
			} else {
				values = append(values, &queryv1.ValueRange{
					From: float64(i) * rangeSize,
					To:   float64(i+1) * rangeSize,
				})
			}
		}
	}
	return values, nil
}

func getScript(script string) (string, bool) {
	if strings.HasPrefix(script, "(") && strings.HasSuffix(script, ")") {
		return script, true
	} else if strings.ContainsAny(script, "+-*/") {
		return script, true
	}
	return "", false
}

func parseScript(script, keyType string) (string, map[string]struct{}, error) {
	if strings.HasPrefix(script, "(") && strings.HasSuffix(script, ")") {
		if keyType == query.TagKey {
			if match, _ := regexp.Match("doc\\[\\'[a-zA-Z0-9_.]+\\'\\]", reflectx.StringToBytes(script)); match {
				// 作为原始的es脚本
				return script, nil, nil
			}
		} else {
			if match, _ := regexp.Match("return .*", reflectx.StringToBytes(script)); match {
				// 作为原始的es脚本
				return script, nil, nil
			}
		}
	}
	result := &bytes.Buffer{}
	fields := make(map[string]struct{})
	var field []rune
	for i, c := range script {
		if unicode.IsSpace(c) {
			continue
		}
		if unicode.IsLetter(c) || c == '_' || c == '.' {
			field = append(field, c)
		} else {
			if len(field) > 0 {
				fieldName := string(field)
				if !strings.HasPrefix(fieldName, "tags.") && !strings.HasPrefix(fieldName, "fields.") &&
					fieldName != "@timestamp" && fieldName != "timestamp" && fieldName != "name" {
					result.WriteString("doc['" + keyType + "." + fieldName + "'].value")
					fields[keyType+"."+fieldName] = struct{}{}
				} else {
					result.WriteString("doc['" + fieldName + "'].value")
					fields[fieldName] = struct{}{}
				}
				field = nil
			}
			if c == ',' {
				if i == len(script)-2 { // (xxx,)
					result.WriteString("+','")
				} else {
					result.WriteString("+','+")
				}
			} else {
				result.WriteRune(c)
			}
		}
	}
	if len(field) > 0 {
		fieldName := string(field)
		if !strings.HasPrefix(fieldName, "tags.") && !strings.HasPrefix(fieldName, "fields.") &&
			fieldName != "@timestamp" && fieldName != "timestamp" && fieldName != "name" {
			result.WriteString("doc['" + keyType + "." + fieldName + "'].value")
			fields[keyType+"."+fieldName] = struct{}{}
		} else {
			result.WriteString("doc['" + fieldName + "'].value")
			fields[fieldName] = struct{}{}
		}
	}
	return result.String(), fields, nil
}

func init() {
	queryv1.RegisterQueryParser("params", &Parser{})
}
