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

package queryv1

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

func TestDynamicPoints(t *testing.T) {
	now := time.Now()
	t.Log(dynamicPoints(&Request{
		Start: now.Add(-time.Second*20).Unix() * 1000,
		End:   now.Unix() * 1000,
	}))
	t.Log(dynamicPoints(&Request{
		Start: now.Add(-time.Second*40).Unix() * 1000,
		End:   now.Unix() * 1000,
	}))
	t.Log(dynamicPoints(&Request{
		Start: now.Add(-time.Minute*20).Unix() * 1000,
		End:   now.Unix() * 1000,
	}))
	t.Log(dynamicPoints(&Request{
		Start: now.Add(-time.Hour*1).Unix() * 1000,
		End:   now.Unix() * 1000,
	}))
	t.Log(dynamicPoints(&Request{
		Start: now.AddDate(0, 0, -2).Unix() * 1000,
		End:   now.Unix() * 1000,
	}))
	t.Log(dynamicPoints(&Request{
		Start: now.AddDate(0, 0, -5).Unix() * 1000,
		End:   now.Unix() * 1000,
	}))
}

func TestMapToRawQuery(t *testing.T) {
	metricParams := make(map[string]string)
	metricParams["start"] = strconv.FormatInt(123, 10)
	metricParams["end"] = strconv.FormatInt(1234, 10)
	metricParams["filter_terminus_key"] = "asdfasdf"
	metricParams["group"] = "trace_id"
	metricParams["limit"] = strconv.FormatInt(11, 10)
	metricParams["sort"] = "max_start_time_min"
	metricParams["sum"] = "errors_sum"
	metricParams["min"] = "start_time_min"
	metricParams["max"] = "end_time_max"
	metricParams["last"] = "labels_distinct"
	metricParams["align"] = "false"

	statement, err := MapToRawQuery("test_metric", "agge", metricParams)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(statement)
}

func TestNormalizeColumn(t *testing.T) {
	tests := []struct {
		col  string
		typ  string
		want string
	}{
		{
			col:  "col",
			typ:  "typ",
			want: "typ.col",
		},
		{
			col:  "col.123",
			typ:  "typ",
			want: "col.123",
		},
	}

	for _, test := range tests {
		got := NormalizeColumn(test.col, test.typ)
		require.Equal(t, test.want, got)
	}

}

func TestNormalName(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{
			key:  ".kkk",
			want: "kkk",
		},
		{
			key:  "_kkk",
			want: "kkk",
		},
		{
			key:  "123123",
			want: "123123",
		},
		{
			key:  "",
			want: "",
		},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			got := NormalizeName(test.key)
			require.Equal(t, test.want, got)
		})
	}
}

func TestGetMapValue(t *testing.T) {
	tests := []struct {
		name string
		key  string
		data map[string]interface{}
		want interface{}
	}{
		{
			name: "single value no value",
			key:  "key",
			data: map[string]interface{}{"": "", "11": "value"},
			want: nil,
		},
		{
			name: "single value get value",
			key:  "key",
			data: map[string]interface{}{"key": "value"},
			want: "value",
		},
		{
			name: "nested",
			key:  "key.key1",
			data: map[string]interface{}{"key": map[string]interface{}{"key1": "111"}},
			want: "111",
		},
		{
			name: "nested two",
			key:  "key.key1.key2",
			data: map[string]interface{}{"key": map[string]interface{}{"key1": map[string]interface{}{"key2": "222"}}},
			want: "222",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getMapValue(test.key, test.data)
			require.Equal(t, test.want, got)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type testStruct struct {
		Column1  string
		Column2  string
		Silence  []string
		Duration time.Duration
	}

	tests := []struct {
		input   interface{}
		want    interface{}
		wantErr bool
		name    string
	}{
		{
			name: "unmarshal true",
			input: map[string]interface{}{
				"column1": "111",
				"column2": "222",
			},
			want: &testStruct{Column1: "111", Column2: "222"},
		},
		{
			name:    "error",
			input:   "1111",
			wantErr: true,
		},
		{
			name: "string to slice",
			input: map[string]interface{}{
				"column1": "111",
				"column2": "222",
				"Silence": "1,2,3,4,5",
			},
			want: &testStruct{Column1: "111", Column2: "222", Silence: []string{"1", "2", "3", "4", "5"}},
		},
		{
			name: "string to time duration, true",
			input: map[string]interface{}{
				"column1":  "111",
				"column2":  "222",
				"Duration": "3h",
			},
			want: &testStruct{Column1: "111", Column2: "222", Duration: time.Hour * 3},
		},

		{
			name: "string to time duration, false",
			input: map[string]interface{}{
				"column1":  "111",
				"column2":  "222",
				"Duration": "3G",
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &testStruct{}
			err := Unmarshal(test.input, output)
			if test.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.want, output)
		})
	}
}

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name string
		fn   string
		p    *Property
		want string
	}{
		{
			name: "no script",
			fn:   "function",
			p: &Property{
				Key:    "id",
				Script: "",
			},
			want: "function_id",
		},
		{
			name: "script",
			fn:   "function",
			p: &Property{
				Name:   "id",
				Script: "script",
			},
			want: "1fb3693e65dfdf89",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NormalizeID(test.fn, test.p)
			require.Equal(t, test.want, got)
		})
	}
}

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		name string
		keys string
		typ  string
		want string
	}{
		{
			name: "empty key",
			keys: "",
			want: "",
		},
		{
			name: "{key}",
			keys: "{key}",
			typ:  "string",
			want: "{key}",
		},
		{
			name: "key",
			keys: "key",
			typ:  "",
			want: "key",
		},
		{
			name: "key,type string",
			keys: "key",
			typ:  "string",
			want: "string.key",
		},
		{
			name: ".key",
			keys: ".key",
			want: "key",
		},
		{
			name: "_name",
			keys: "_name",
			want: "name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NormalizeKey(test.keys, test.typ)
			require.Equal(t, test.want, got)
		})
	}
}

func TestNormalizeRequest(t *testing.T) {
	clusterWhere := model.Filter{
		Key:      model.ClusterNameKey,
		Operator: "in",
		Value:    []interface{}{"cluster1", "cluster2"},
	}

	standardTime := time.Now()
	t.Logf("now, %v", standardTime.UnixNano())
	tests := []struct {
		name    string
		req     Request
		want    Request
		wantErr bool
		mockNow func() time.Time
	}{
		{
			name: "normal",
			req:  Request{},
			want: Request{
				Name:             "",
				Metrics:          []string{""},
				Start:            0,
				End:              1,
				TimeAlign:        "",
				Select:           nil,
				Where:            nil,
				GroupBy:          nil,
				OrderBy:          nil,
				Limit:            []int{20},
				Debug:            false,
				Aggregate:        nil,
				ExistKeys:        map[string]struct{}{},
				Columns:          map[string]*Column{},
				TimeKey:          model.TimestampKey,
				OriginalTimeUnit: tsql.Nanosecond,
				EndOffset:        0,
				Interval:         float64(time.Millisecond),
				Points:           0,
				AlignEnd:         false,
				ClusterNames:     nil,
				LegendMap:        nil,
				ChartType:        "",
				Trans:            false,
				TransGroup:       false,
				DefaultNullValue: nil,
			},
		},
		{
			name: "-1 ~ -1",
			mockNow: func() time.Time {
				return standardTime
			},
			req: Request{
				TimeAlign: TimeAlignNone,
				Start:     -1,
				End:       -1,
			},
			want: Request{
				Name:             "",
				Metrics:          []string{""},
				Start:            0,
				End:              1,
				TimeAlign:        TimeAlignNone,
				Select:           nil,
				Where:            nil,
				GroupBy:          nil,
				OrderBy:          nil,
				Limit:            []int{20},
				Debug:            false,
				Aggregate:        nil,
				ExistKeys:        map[string]struct{}{},
				Columns:          map[string]*Column{},
				TimeKey:          model.TimestampKey,
				OriginalTimeUnit: tsql.Nanosecond,
				EndOffset:        0,
				Interval:         float64(time.Millisecond),
				Points:           0,
				AlignEnd:         false,
				ClusterNames:     nil,
				LegendMap:        nil,
				ChartType:        "",
				Trans:            false,
				TransGroup:       false,
				DefaultNullValue: nil,
			},
		},
		{
			name: "end <= start",
			mockNow: func() time.Time {
				return standardTime
			},
			req: Request{
				TimeAlign: TimeAlignNone,
				Start:     10,
				End:       10,
			},
			want: Request{
				Name:             "",
				Metrics:          []string{""},
				Start:            0,
				End:              1,
				TimeAlign:        TimeAlignNone,
				Select:           nil,
				Where:            nil,
				GroupBy:          nil,
				OrderBy:          nil,
				Limit:            []int{20},
				Debug:            false,
				Aggregate:        nil,
				ExistKeys:        map[string]struct{}{},
				Columns:          map[string]*Column{},
				TimeKey:          model.TimestampKey,
				OriginalTimeUnit: tsql.Nanosecond,
				EndOffset:        0,
				Interval:         float64(time.Millisecond),
				Points:           0,
				AlignEnd:         false,
				ClusterNames:     nil,
				LegendMap:        nil,
				ChartType:        "",
				Trans:            false,
				TransGroup:       false,
				DefaultNullValue: nil,
			},
		},
		{
			name: "cluster where",
			mockNow: func() time.Time {
				return standardTime
			},
			req: Request{
				TimeAlign: TimeAlignNone,
				Start:     10,
				End:       10,
				Where: []*model.Filter{
					&clusterWhere,
				},
			},
			want: Request{
				Name:      "",
				Metrics:   []string{""},
				Start:     0,
				End:       1,
				TimeAlign: TimeAlignNone,
				Select:    nil,
				Where: []*model.Filter{
					&clusterWhere,
				},
				GroupBy:          nil,
				OrderBy:          nil,
				Limit:            []int{20},
				Debug:            false,
				Aggregate:        nil,
				ExistKeys:        map[string]struct{}{},
				Columns:          map[string]*Column{},
				TimeKey:          model.TimestampKey,
				OriginalTimeUnit: tsql.Nanosecond,
				EndOffset:        0,
				Interval:         float64(time.Millisecond),
				Points:           0,
				AlignEnd:         false,
				ClusterNames:     []string{"cluster1", "cluster2"},
				LegendMap:        nil,
				ChartType:        "",
				Trans:            false,
				TransGroup:       false,
				DefaultNullValue: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockNow != nil {
				nowFunction = test.mockNow
			}
			err := NormalizeRequest(&test.req)
			if test.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.want, test.req)
		})
	}

}
