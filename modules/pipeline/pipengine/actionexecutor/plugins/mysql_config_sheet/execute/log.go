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

package execute

import (
	"context"
	"reflect"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pkg/jsonparse"
	"github.com/erda-project/erda/modules/pipeline/pkg/log_collector"
)

func printOutParams(ctx context.Context, outParams map[string]interface{}, meta *Meta) {
	log := log_collector.Clog(ctx)

	if len(outParams) == 0 {
		return
	}
	log.Printf("Out Params:")
	defer addNewLine(ctx)

	// 按定义顺序返回
	for _, define := range meta.OutParams {
		k := define.Key
		v, ok := outParams[k]
		if !ok {
			continue
		}
		meta.OutParamsResult[k] = v
		log.Printf("  arg: %s", k)
		log.Printf("  source: %s", define.Source.String())
		if define.Expression != "" {
			log.Printf("  expr: %s", define.Expression)
		}
		log.Printf("  value: %s", jsonparse.JsonOneLine(ctx, v))
		var vtype string
		if v == nil {
			vtype = "nil"
		} else {
			vtype = reflect.TypeOf(v).String()
		}
		log.Printf("  type: %s", vtype)
		addLineDelimiter(ctx, "  ")
	}
}

func printAssertResults(ctx context.Context, success bool, results []*apistructs.APITestsAssertData) {
	log := log_collector.Clog(ctx)

	log.Printf("Assert Result: %t", success)
	defer addNewLine(ctx)

	log.Printf("Assert Detail:")
	for _, result := range results {
		log.Printf("  arg: %s", result.Arg)
		log.Printf("  operator: %s", result.Operator)
		log.Printf("  value: %s", result.Value)
		log.Printf("  actualValue: %s", jsonparse.JsonOneLine(ctx, result.ActualValue))
		log.Printf("  success: %t", result.Success)
		if result.ErrorInfo != "" {
			log.Printf("  errorInfo: %s", result.ErrorInfo)
		}
		addLineDelimiter(ctx, "  ")
	}
}


func addLineDelimiter(ctx context.Context, prefix ...string) {
	log := log_collector.Clog(ctx)

	var _prefix string
	if len(prefix) > 0 {
		_prefix = prefix[0]
	}
	log.Printf("%s==========", _prefix)
}

func addNewLine(ctx context.Context, num ...int) {
	log := log_collector.Clog(ctx)

	_num := 1
	if len(num) > 0 {
		_num = num[0]
	}
	if _num <= 0 {
		_num = 1
	}
	for i := 0; i < _num; i++ {
		log.Println()
	}
}
