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

package debug

import (
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

// MergeBreakpoint merge pipeline and task breakpoint config
// pipeline breakpoint config is globally used for all tasks
// if task breakpoint config is not empty, it will override pipeline breakpoint config
func MergeBreakpoint(taskConfig, pipelineConfig *basepb.Breakpoint) basepb.Breakpoint {
	breakpoint := basepb.Breakpoint{}
	if pipelineConfig != nil {
		breakpoint.On = pipelineConfig.On
		if pipelineConfig.Timeout != nil {
			breakpoint.Timeout = pipelineConfig.Timeout
		}
	}
	if taskConfig != nil {
		breakpoint.On = taskConfig.On
		if taskConfig.Timeout != nil {
			breakpoint.Timeout = taskConfig.Timeout
		}
	}
	return breakpoint
}

func ParseDebugTimeout(value *structpb.Value) (*time.Duration, error) {
	if value == nil {
		return nil, nil
	}
	val := value.AsInterface()
	switch val.(type) {
	case string:
		d, err := time.ParseDuration(val.(string))
		if err != nil {
			return nil, err
		}
		return &d, nil
	case float64:
		d := time.Duration(val.(float64)) * time.Second
		return &d, nil
	case int:
		d := time.Duration(val.(int)) * time.Second
		return &d, nil
	case int32:
		d := time.Duration(val.(int32)) * time.Second
		return &d, nil
	case int64:
		d := time.Duration(val.(int64)) * time.Second
		return &d, nil
	default:
		return nil, fmt.Errorf("invalid defebug timeout value: %v, type: %v", val, reflect.TypeOf(val).Name())
	}
}
