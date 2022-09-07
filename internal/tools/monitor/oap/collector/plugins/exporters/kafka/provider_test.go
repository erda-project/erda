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

package kafka

import (
	"testing"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
)

func Test_provider_ExportMetric(t *testing.T) {
	type args struct {
		items []*metric.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{items: []*metric.Metric{
				&metric.Metric{
					Name:      "cpu",
					Timestamp: 1,
					Tags:      map[string]string{"org_name": "erda"},
					Fields: map[string]interface{}{
						"usege_system": 11,
					},
					OrgName: "erda",
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				writer: &mockWriter{},
				Cfg:    &config{Topic: "test"},
			}
			if err := p.ExportMetric(tt.args.items...); (err != nil) != tt.wantErr {
				t.Errorf("ExportMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockWriter struct {
}

func (m *mockWriter) Write(data interface{}) error {
	return nil
}

func (m *mockWriter) WriteN(data ...interface{}) (int, error) {
	return 0, nil
}

func (m *mockWriter) Close() error {
	return nil
}
