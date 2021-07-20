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

package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda/modules/monitor/core/logs"
	"github.com/erda-project/erda/modules/monitor/core/logs/pb"
)

func (p *provider) invokeV2(key []byte, value []byte, topic *string, timestamp time.Time) error {
	lb := &pb.LogBatch{}
	if err := lb.Unmarshal(value); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	for _, log := range lb.Logs {
		p.processLogV2(log)

		cacheKey := log.Source + "_" + log.Id
		if !p.cache.Has(cacheKey) {
			// store meta
			meta := &logs.LogMeta{
				ID:     log.Id,
				Source: log.Source,
				Tags:   log.Tags,
			}
			p.output.Write(meta)
			p.cache.SetWithExpire(cacheKey, meta, time.Hour)
		}

		if err := p.output.Write(log); err != nil {
			return err
		}
	}

	return nil
}

func (p *provider) processLogV2(log *pb.Log) {
	if log.Tags == nil {
		log.Tags = make(map[string]string)
	}

	level, ok := log.Tags["level"]
	if !ok {
		level = "INFO" // default log level
	} else {
		level = strings.ToUpper(level)
	}
	log.Tags["level"] = level

	for _, key := range p.Cfg.Output.IDKeys {
		if val, ok := log.Tags[key]; ok {
			log.Id = val
			break
		}
	}

	if log.Stream == "" {
		log.Stream = "stdout" // default log stream
	}
}
