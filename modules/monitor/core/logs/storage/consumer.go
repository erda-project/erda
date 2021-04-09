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
	"context"
	"strings"
	"time"

	"github.com/erda-project/erda/modules/monitor/core/logs"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (p *provider) invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	log := &logs.Log{}
	if err := json.Unmarshal(value, log); err != nil {
		return err
	}
	p.processLog(log)

	cacheKey := log.Source + "_" + log.ID
	if !p.cache.Has(cacheKey) {
		// store meta
		meta := &logs.LogMeta{
			ID:     log.ID,
			Source: log.Source,
			Tags:   log.Tags,
		}
		p.output.Write(meta)
		p.cache.SetWithExpire(cacheKey, meta, time.Hour)
	}

	count(log)
	return p.output.Write(log)
}

func (p *provider) processLog(log *logs.Log) {
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
			log.ID = val
			break
		}
	}

	if log.Stream == "" {
		log.Stream = "stdout" // default log stream
	}
}

func (p *provider) storeMetaCache() {
	for _, meta := range p.cache.GetALL(true) {
		if err := p.output.Write(meta); err != nil {
			logrus.Errorf("fail to write log meta: %s", err)
		}
	}
}

func (p *provider) startStoreMetaCache(ctx context.Context) {
	ticker := time.NewTicker(p.Cfg.Output.Cassandra.CacheStoreInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.storeMetaCache()
		case <-ctx.Done():
			return
		}
	}
}

const (
	platformKey        = "platform"
	componentKey       = "component"
	componentNameKey   = "component_name"
	componentTypeKey   = "component_type"
	orgIDKey           = "org_id"
	orgNameKey         = "org_name"
	clusterNameKey     = "cluster_name"
	projectIDKey       = "project_id"
	projectNameKey     = "project_name"
	applicationIDKey   = "application_id"
	applicationNameKey = "application_name"
	workspaceKey       = "workspace"
	levelKey           = "level"

	dicePrefix             = "dice_"
	diceComponentKey       = dicePrefix + componentKey
	diceOrgIDKey           = dicePrefix + orgIDKey
	diceOrgNameKey         = dicePrefix + orgNameKey
	diceClusterNameKey     = dicePrefix + clusterNameKey
	diceProjectIDKey       = dicePrefix + projectIDKey
	diceProjectNameKey     = dicePrefix + projectNameKey
	diceApplicationIDKey   = dicePrefix + applicationIDKey
	diceApplicationNameKey = dicePrefix + applicationNameKey
	diceWorkspaceKey       = dicePrefix + workspaceKey

	srcKey                = "src"
	srcPrefix             = "src_"
	srcComponentNameKey   = srcPrefix + componentNameKey
	srcComponentTypeKey   = srcPrefix + componentTypeKey
	srcOrgNameKey         = srcPrefix + orgNameKey
	srcClusterNameKey     = srcPrefix + clusterNameKey
	srcProjectIDKey       = srcPrefix + projectIDKey
	srcProjectNameKey     = srcPrefix + projectNameKey
	srcApplicationIDKey   = srcPrefix + applicationIDKey
	srcApplicationNameKey = srcPrefix + applicationNameKey
	srcWorkspaceKey       = srcPrefix + workspaceKey
)

func count(log *logs.Log) {
	componentName := log.Tags[diceComponentKey]
	var componentType string
	if componentName != "" {
		componentType = platformKey
	}
	logBytesCounter.WithLabelValues(
		log.Tags[levelKey],
		log.Source,
		componentType,
		componentName,
		log.Tags[diceClusterNameKey],
		log.Tags[diceOrgNameKey],
		log.Tags[diceProjectIDKey],
		log.Tags[diceProjectNameKey],
		log.Tags[diceApplicationIDKey],
		log.Tags[diceApplicationNameKey],
		log.Tags[diceWorkspaceKey],
	).Add(float64(len(log.Content)))
}
