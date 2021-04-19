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

package apitestsv2

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/assert"
)

// JudgeAsserts 执行断言测试
func (at *APITest) JudgeAsserts(outParams map[string]interface{}, asserts []apistructs.APIAssert) (bool, []*apistructs.APITestsAssertData) {
	var results []*apistructs.APITestsAssertData
	for _, ast := range asserts {
		// 出参里的值
		actualValue := outParams[ast.Arg]
		//var buffer bytes.Buffer
		//enc := json.NewEncoder(&buffer)
		//enc.SetEscapeHTML(false)
		//enc.Encode(actualValue)
		//actualValueString := buffer.String()
		//if len(actualValueString) > 0 {
		//	actualValueString = actualValueString[:len(actualValueString)-1]
		//}
		succ, err := assert.DoAssert(actualValue, ast.Operator, ast.Value)
		result := apistructs.APITestsAssertData{
			Arg:         ast.Arg,
			Operator:    ast.Operator,
			Value:       ast.Value,
			Success:     succ,
			ActualValue: actualValue,
			ErrorInfo: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}
		results = append(results, &result)
	}
	// 判断结果
	globalSuccess := true
	for _, result := range results {
		if !result.Success {
			globalSuccess = false
			break
		}
	}
	return globalSuccess, results
}
