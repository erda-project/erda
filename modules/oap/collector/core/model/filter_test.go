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

package model

import (
	"testing"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"

	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	"github.com/stretchr/testify/assert"
)

func TestDataFilter_Selected(t *testing.T) {
	type fields struct {
		cfg FilterConfig
	}
	type args struct {
		od odata.ObservableData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			fields: fields{cfg: FilterConfig{
				Namepass: []string{"abc"},
			}},
			args: args{od: odata.NewMetric(&mpb.Metric{
				Name:         "abcd",
				TimeUnixNano: 0,
			})},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			df, err := NewDataFilter(tt.fields.cfg)
			assert.Nil(t, err)
			if got := df.Selected(tt.args.od); got != tt.want {
				t.Errorf("Selected() = %v, want %v", got, tt.want)
			}
		})
	}
}
