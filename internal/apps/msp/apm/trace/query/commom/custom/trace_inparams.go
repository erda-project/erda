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

package custom

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
)

type Model struct {
	DurationMin int64       `json:"durationMin"`
	DurationMax int64       `json:"durationMax"`
	TenantId    string      `json:"tenantId"`
	StartTime   int64       `json:"startTime"` // ms
	EndTime     int64       `json:"endTime"`   // ms
	Status      string      `json:"status"`
	Limit       int64       `json:"limit"`
	Conditions  []Condition `json:"conditions"`
}

type Condition struct {
	TraceId     string `json:"traceId"`
	ServiceName string `json:"serviceName"`
	RpcMethod   string `json:"rpcMethod"`
	HttpPath    string `json:"httpPath"`
	Operator
}

func (m Model) ConvertCondition() []*pb.Condition {
	if len(m.Conditions) <= 0 {
		return []*pb.Condition{}
	}
	var result []*pb.Condition
	for _, condition := range m.Conditions {
		result = append(result, &pb.Condition{
			TraceID:     condition.TraceId,
			HttpPath:    condition.HttpPath,
			RpcMethod:   condition.RpcMethod,
			ServiceName: condition.ServiceName,
			Operator:    condition.OperatorText(),
		})
	}
	return result
}

func ConvertConditionByPbCondition(conditions []*pb.Condition) []Condition {
	if len(conditions) <= 0 {
		return []Condition{}
	}
	var result []Condition
	for _, condition := range conditions {
		result = append(result, Condition{
			TraceId:     condition.TraceID,
			HttpPath:    condition.HttpPath,
			RpcMethod:   condition.RpcMethod,
			ServiceName: condition.ServiceName,
			Operator:    Operator{condition.Operator},
		})
	}
	return result
}

type Operator struct {
	Operator string `json:"operator"`
}

type TraceInParams struct {
	InParamsPtr *Model
}

func (b *TraceInParams) CustomInParamsPtr() interface{} {
	b.InParamsPtr = &Model{}
	return b.InParamsPtr
}

func (b *TraceInParams) EncodeFromCustomInParams(customInParamsPtr interface{}, stdInParamsPtr *cptype.ExtraMap) {
	cputil.MustObjJSONTransfer(customInParamsPtr, stdInParamsPtr)
}

func (b *TraceInParams) DecodeToCustomInParams(stdInParamsPtr *cptype.ExtraMap, customInParamsPtr interface{}) {
	cputil.MustObjJSONTransfer(stdInParamsPtr, customInParamsPtr)
}

func (m Operator) IsNotEqualOperator() bool {
	return m.Operator == "!="
}

func (m Operator) OperatorText() string {
	if m.Operator == "" {
		return "="
	}
	return m.Operator
}
