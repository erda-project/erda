// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package collector

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/kafka"
)

var (
	topics = map[string]string{
		"metrics":       topicPrefix + "metrics",
		"trace":         topicPrefix + "trace",
		"container_log": topicPrefix + "container-log",
		"job_log":       topicPrefix + "job-log",
		"analytics":     topicPrefix + "analytics",
		// 白名单过滤
		"alert":                topicPrefix + "alert",
		"alert-event":          topicPrefix + "alert-event",
		"error":                topicPrefix + "error",
		"metaserver_container": topicPrefix + "metaserver_container",
		"metaserver_host":      topicPrefix + "metaserver_host",
		"metrics-temp":         topicPrefix + "metrics-temp",
		"monitor-log":          topicPrefix + "monitor-log",
	}
)

const (
	topicPrefix = "spot-"
)

func (c *collector) send(name string, data []byte) error {
	topic, err := c.getTopic(name)
	if err != nil {
		return err
	}
	return c.writer.Write(&kafka.Message{
		Topic: &topic,
		Data:  data,
	})
}

func (c *collector) getTopic(typ string) (string, error) {
	if topic, ok := topics[typ]; ok {
		return topic, nil
	}
	return "", errors.Errorf("not support type")
}
