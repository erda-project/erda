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

package metrics

import (
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
)

var counter *prometheus.CounterVec

func init() {
	_ = CounterVec()
}

func CounterVec() *prometheus.CounterVec {
	if counter == nil {
		counter = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "erda",
			Subsystem:   "ai_proxy",
			Name:        "request_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: nil,
		}, new(LabelValues).Labels())
		prometheus.MustRegister(counter)
	}
	return counter
}

type LabelValues struct {
	ChatType    string `json:"chat_type"`
	ChatTitle   string `json:"chat_title"`
	Source      string `json:"source"`
	UserId      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	OperationId string `json:"operation_id"`
	Status      string `json:"status"`
	StatusCode  string `json:"status_code"`
}

func (l LabelValues) Values() []string {
	t := reflect.TypeOf(l)
	v := reflect.ValueOf(l)
	var result = make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		result[i] = v.Field(i).String()
	}
	return result
}

func (l LabelValues) Labels() []string {
	t := reflect.TypeOf(l)
	var keys = make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		keys[i] = t.Field(i).Tag.Get("json")
	}
	return keys
}

func (l LabelValues) LabelValues() []string {
	labels := l.Labels()
	values := l.Values()
	result := make([]string, len(labels))
	for i := 0; i < len(labels); i++ {
		result[i] = labels[i] + "=" + values[i]
	}
	return result
}
