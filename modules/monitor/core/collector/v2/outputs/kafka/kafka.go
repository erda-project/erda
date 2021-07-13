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

package kafka

import (
	"context"
	"errors"

	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/kafka"
)

const (
	topicPrefix = "spot-v2-"

	Selector = "kafka"
)

var (
	ErrTopicMustSpecify = errors.New("topic must specify")
	ErrTopicInvalid     = errors.New("invalid topic")

	topics = map[string]string{
		"metrics":       topicPrefix + "metrics",
		"trace":         topicPrefix + "trace",
		"container_log": topicPrefix + "container-log",
		"job_log":       topicPrefix + "job-log",
		"analytics":     topicPrefix + "analytics",
		// white list
		"alert":                topicPrefix + "alert",
		"alert-event":          topicPrefix + "alert-event",
		"error":                topicPrefix + "error",
		"metaserver_container": topicPrefix + "metaserver_container",
		"metaserver_host":      topicPrefix + "metaserver_host",
		"metrics-temp":         topicPrefix + "metrics-temp",
		"monitor-log":          topicPrefix + "monitor-log",
	}
)

func New(w writer.Writer) *Output {
	return &Output{
		w: w,
	}
}

type Output struct {
	w writer.Writer
}

func (o *Output) Send(ctx context.Context, data []byte) error {
	topic, ok := ctx.Value("topic").(string)
	if !ok {
		return ErrTopicMustSpecify
	}

	topic, err := getTopic(topic)
	if err != nil {
		return err
	}

	return o.w.Write(&kafka.Message{
		Topic: &topic,
		Data:  data,
	})
}

func getTopic(typ string) (string, error) {
	if topic, ok := topics[typ]; ok {
		return topic, nil
	}
	return "", ErrTopicInvalid
}
