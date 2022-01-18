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

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
)

func TestSeries_Eval(t *testing.T) {
	now := time.Now()
	type fields struct {
		points     []Point
		fieldKey   string
		sampleName string
		functions  []evaluationPoints
		alias      string
	}
	tests := []struct {
		name    string
		fields  fields
		want    model.ObservableData
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				points: []Point{
					{
						Value:         100,
						TimestampNano: now.UnixNano(),
					},
					{
						Value:         110,
						TimestampNano: now.Add(time.Second).UnixNano(),
					},
					{
						Value:         120,
						TimestampNano: now.Add(2 * time.Second).UnixNano(),
					},
					{
						Value:         130,
						TimestampNano: now.Add(3 * time.Second).UnixNano(),
					},
				},
				sampleName: "sample",
				fieldKey:   "field-1",
				functions: []evaluationPoints{
					Functions[rate], Functions[multiply100],
				},
				alias: "field-2",
			},
			want: &model.Metrics{Metrics: []*mpb.Metric{
				{
					Name:         "sample",
					TimeUnixNano: uint64(now.Add(time.Second).UnixNano()),
					Attributes: map[string]string{
						"cluster_name": "dev",
					},
					DataPoints: map[string]*structpb.Value{
						"field-2": structpb.NewNumberValue(10 * 100),
					},
				},
				{
					Name:         "sample",
					TimeUnixNano: uint64(now.Add(2 * time.Second).UnixNano()),
					Attributes: map[string]string{
						"cluster_name": "dev",
					},
					DataPoints: map[string]*structpb.Value{
						"field-2": structpb.NewNumberValue(10 * 100),
					},
				},
				{
					Name:         "sample",
					TimeUnixNano: uint64(now.Add(3 * time.Second).UnixNano()),
					Attributes: map[string]string{
						"cluster_name": "dev",
					},
					DataPoints: map[string]*structpb.Value{
						"field-2": structpb.NewNumberValue(10 * 100),
					},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := &Series{
				points:     tt.fields.points,
				sampleName: tt.fields.sampleName,
				dataType:   model.MetricDataType,
				fieldKey:   tt.fields.fieldKey,
				tags: map[string]string{
					"cluster_name": "dev",
				},
				functions: tt.fields.functions,
				alias:     tt.fields.alias,
			}
			got, err := se.Eval()
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}
