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
	"sort"
	"time"
)

type aggFunc string

const (
	rate        aggFunc = "rate"
	multiply100 aggFunc = "multiply:100"
)

type evaluationPoints func(points []Point) []Point

var Functions = map[aggFunc]evaluationPoints{
	rate:        evalRate,
	multiply100: evalMultiply(100),
}

func evalRate(points []Point) []Point {
	n := len(points)
	if n < 2 {
		return points
	}
	sort.Sort(sortedPoints(points))
	res := make([]Point, n-1)
	for i := 1; i < n; i++ {
		cur, pre := points[i], points[i-1]
		grow := cur.Value - pre.Value
		if cur.Value < pre.Value {
			grow = cur.Value
		}
		elapsed := float64(time.Duration(cur.TimestampNano-pre.TimestampNano) / time.Second)
		res[i-1] = Point{
			Value:         grow / elapsed,
			TimestampNano: cur.TimestampNano,
		}
	}
	return res
}

func evalMultiply(factor float64) evaluationPoints {
	return func(points []Point) []Point {
		sort.Sort(sortedPoints(points))
		res := make([]Point, len(points))
		for i := 0; i < len(points); i++ {
			res[i] = Point{
				Value:         points[i].Value * factor,
				TimestampNano: points[i].TimestampNano,
			}
		}
		return res
	}
}
