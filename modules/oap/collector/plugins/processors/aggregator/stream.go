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
	"errors"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
)

var (
	ErrInvalidDataType     = errors.New("invalid data type")
	ErrUnsupportedDataType = errors.New("unsupported data type")
)

// Series is a stream of data points belonging to a metric.
type Series struct {
	points []Point

	// information about original sample data
	sampleName string
	dataType   model.DataType
	tags       map[string]string
	fieldKey   string

	// information about evaluation
	functions []evaluationPoints
	alias     string
}

func (se *Series) Eval() (model.ObservableData, error) {
	points := se.points
	for _, fn := range se.functions {
		points = fn(points)
	}

	np := len(points)
	switch se.dataType {
	case model.MetricDataType:
		res := &model.Metrics{Metrics: make([]*mpb.Metric, np)}
		field := se.fieldKey
		if se.alias != "" {
			field = se.alias
		}

		for i := 0; i < np; i++ {
			res.Metrics[i] = &mpb.Metric{
				Name:         se.sampleName,
				TimeUnixNano: uint64(points[i].TimestampNano),
				Attributes:   se.tags,
				DataPoints: map[string]*structpb.Value{
					field: structpb.NewNumberValue(points[i].Value),
				},
			}
		}
		return res, nil
	}
	return nil, fmt.Errorf("datatype is %s: %w", se.dataType, ErrUnsupportedDataType)
}

type sortedPoints []Point

func (s sortedPoints) Len() int {
	return len(s)
}

func (s sortedPoints) Less(i, j int) bool {
	return s[i].TimestampNano < s[j].TimestampNano
}

func (s sortedPoints) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Point represents a single data point for a given timestamp.
type Point struct {
	Value         float64
	TimestampNano int64
}

type Appender struct {
	rw    sync.RWMutex
	cache map[uint64]Series
}

func NewAppender() *Appender {
	return &Appender{
		cache: make(map[uint64]Series),
	}
}

// TODO drop item whose timestamp less than smallest?
func (a *Appender) AddItem(item *model.DataItem, fieldKey string, r rule) error {
	val, ok := item.Fields[fieldKey]
	if !ok {
		return nil
	}
	if item.Type != model.MetricDataType {
		return ErrUnsupportedDataType
	}

	id := item.HashDataItem(fieldKey)
	a.rw.Lock()
	defer a.rw.Unlock()
	if vv, ok := a.cache[id]; !ok {
		fns, err := convertToFunctions(r.Functions)
		if err != nil {
			return fmt.Errorf("convert functions: %w", err)
		}
		a.cache[id] = Series{
			dataType:   item.Type,
			tags:       item.Tags,
			sampleName: item.Name,
			fieldKey:   fieldKey,
			functions:  fns,
			alias:      r.Alias,
			points: []Point{
				{Value: val.GetNumberValue(), TimestampNano: int64(item.TimestampNano)},
			},
		}
	} else {
		if vv.dataType != item.Type {
			return ErrInvalidDataType
		}

		vv.points = append(vv.points, Point{
			Value:         val.GetNumberValue(),
			TimestampNano: int64(item.TimestampNano),
		})
		a.cache[id] = vv
	}
	return nil
}

func convertToFunctions(fns []string) ([]evaluationPoints, error) {
	res := make([]evaluationPoints, len(fns))
	for idx, item := range fns {
		fn, ok := Functions[aggFunc(item)]
		if !ok {
			return nil, fmt.Errorf("unsupported function: %s", item)
		}
		res[idx] = fn
	}
	return res, nil
}

func (a *Appender) createCacheSnapshot() map[uint64]Series {
	a.rw.Lock()
	snap := a.cache
	a.cache = make(map[uint64]Series)
	a.rw.Unlock()
	return snap
}
