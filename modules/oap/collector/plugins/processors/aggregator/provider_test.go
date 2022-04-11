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

package aggregator

import (
	"testing"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/oap/collector/common/pbconvert"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
)

func Test_provider_Init(t *testing.T) {
	type fields struct {
		Cfg *config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			fields: fields{
				Cfg: &config{
					Rules: []RuleConfig{
						{
							Func:      "rate",
							Args:      []interface{}{"counter_a"},
							TargetKey: "rate_a",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			fields: fields{
				Cfg: &config{
					Rules: []RuleConfig{
						{
							Func:      "rate",
							Args:      []interface{}{map[string]string{"a": "a"}},
							TargetKey: "rate_a",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				Cfg: tt.fields.Cfg,
			}

			if err := p.Init(nil); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_provider_add(t *testing.T) {
	type fields struct {
		Cfg *config
	}
	type args struct {
		in []odata.ObservableData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []odata.ObservableData
	}{
		{
			fields: fields{
				Cfg: &config{
					Rules: []RuleConfig{
						{
							Func:      "rate",
							Args:      []interface{}{"__dp__counter_a"},
							TargetKey: "__dp__rate_a",
						},
						{
							Func:      "*",
							Args:      []interface{}{"__dp__rate_a", 100},
							TargetKey: "__dp__rate_a",
						},
					},
				},
			},
			args: args{
				in: []odata.ObservableData{
					odata.NewMetric(&mpb.Metric{
						Name:         "name",
						TimeUnixNano: uint64(1649397600 * time.Second),
						DataPoints: map[string]*structpb.Value{
							"counter_a": pbconvert.ToValue(1),
						},
					}),
					odata.NewMetric(&mpb.Metric{
						Name:         "name",
						TimeUnixNano: uint64(1649397601 * time.Second),
						DataPoints: map[string]*structpb.Value{
							"counter_a": pbconvert.ToValue(2),
						},
					}),
					odata.NewMetric(&mpb.Metric{
						Name:         "name",
						TimeUnixNano: uint64(1649397602 * time.Second),
						DataPoints: map[string]*structpb.Value{
							"counter_a": pbconvert.ToValue(3),
						},
					}),
				},
			},
			want: []odata.ObservableData{
				odata.NewMetric(&mpb.Metric{
					Name:         "name",
					TimeUnixNano: uint64(1649397600 * time.Second),
					DataPoints: map[string]*structpb.Value{
						"counter_a": pbconvert.ToValue(1),
					},
				}),
				odata.NewMetric(&mpb.Metric{
					Name:         "name",
					TimeUnixNano: uint64(1649397601 * time.Second),
					DataPoints: map[string]*structpb.Value{
						"counter_a": pbconvert.ToValue(2),
						"rate_a":    pbconvert.ToValue(100),
					},
				}),
				odata.NewMetric(&mpb.Metric{
					Name:         "name",
					TimeUnixNano: uint64(1649397602 * time.Second),
					DataPoints: map[string]*structpb.Value{
						"counter_a": pbconvert.ToValue(3),
						"rate_a":    pbconvert.ToValue(100),
					},
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				Cfg: tt.fields.Cfg,
			}
			assert.Nil(t, p.Init(nil))
			for i := 0; i < len(tt.args.in); i++ {
				assert.Equal(t, tt.want[i].String(), p.add(tt.args.in[i]).String())
			}
		})
	}
}
