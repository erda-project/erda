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

package converter

import (
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

func ToPipelineRunParamsWithValue(runParams []*basepb.PipelineRunParam) []*basepb.PipelineRunParamWithValue {
	var result []*basepb.PipelineRunParamWithValue
	for _, p := range runParams {
		result = append(result, &basepb.PipelineRunParamWithValue{
			Name:      p.Name,
			Value:     p.Value,
			TrueValue: nil,
		})
	}
	return result
}

func ToPipelineRunParams(runParamsWithValue []*basepb.PipelineRunParamWithValue) []*basepb.PipelineRunParam {
	var result []*basepb.PipelineRunParam
	for _, p := range runParamsWithValue {
		result = append(result, &basepb.PipelineRunParam{Name: p.Name, Value: p.Value})
	}
	return result
}
