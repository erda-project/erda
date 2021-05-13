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

package analysis

import (
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/recallsong/go-utils/errorx"

	"github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors"
	"github.com/erda-project/erda/modules/monitor/core/logs"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (p *provider) invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	pv := p.processors.Load()
	if pv == nil {
		// processors not ready, so return
		return nil
	}
	log := &logs.Log{}
	if err := json.Unmarshal(value, log); err != nil {
		return err
	}

	// do filter
	if log.Tags == nil || len(p.C.Filters) <= 0 {
		return nil
	}
	for k, v := range p.C.Filters {
		val, ok := log.Tags[k]
		if !ok {
			return nil
		}
		if len(v) > 0 && v != val {
			return nil
		}
	}

	level, ok := log.Tags["level"]
	if !ok {
		level = "INFO" // default log level
	} else {
		level = strings.ToUpper(level)
	}
	log.Tags["level"] = level
	log.Tags["org_name"] = log.Tags["dice_org_name"]
	log.Tags["cluster_name"] = log.Tags["dice_cluster_name"]
	log.Tags["_meta"] = "true"
	log.Tags["_metric_scope"] = p.C.Processors.Scope
	scopeID := p.C.Processors.ScopeID
	if len(scopeID) <= 0 {
		if len(p.C.Processors.ScopeIDKey) > 0 {
			scopeID = log.Tags[p.C.Processors.ScopeIDKey]
			log.Tags["_metric_scope_id"] = scopeID
		} else if p.C.Processors.Scope == "org" {
			scopeID = log.Tags["dice_org_name"]
			log.Tags["_metric_scope_id"] = scopeID
		}
	} else {
		log.Tags["_metric_scope_id"] = scopeID
	}

	// fmt.Println(jsonx.MarshalAndIndent(log.Tags))

	ps := (pv.(*processors.Processors)).Find("", scopeID, log.Tags)
	var errs errorx.Errors
	for _, processor := range ps {
		name, fields, err := processor.Process(log.Content)
		if err != nil {
			// invalid processor or not match content
			continue
		}
		for k, v := range fields {
			if s, ok := v.(string); ok {
				if _, ok := log.Tags[k]; !ok {
					// 直接在 Tags 上修改，因为这里 len(ps) == 1，不会混淆
					// 后面大盘支持 field 过滤了，再调整
					log.Tags[k] = s
				}
			}
		}
		metric := &metrics.Metric{
			Name:      name,
			Timestamp: log.Timestamp,
			Tags:      log.Tags,
			Fields:    fields,
		}
		err = p.output.Write(metric)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs.MaybeUnwrap()
}
