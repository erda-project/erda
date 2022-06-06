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

package aop

import (
	"context"

	aoptypes2 "github.com/erda-project/erda/modules/tools/pipeline/aop/aoptypes"
	spec2 "github.com/erda-project/erda/modules/tools/pipeline/spec"
)

// NewContextForPipeline 用于快速构造流水线 AOP 上下文
func NewContextForPipeline(p spec2.Pipeline, trigger aoptypes2.TuneTrigger, customKVs ...map[interface{}]interface{}) *aoptypes2.TuneContext {
	ctx := aoptypes2.TuneContext{
		Context: context.Background(),
		SDK:     globalSDK.Clone(),
	}
	ctx.SDK.TuneType = aoptypes2.TuneTypePipeline
	ctx.SDK.TuneTrigger = trigger
	ctx.SDK.Pipeline = p
	// 用户自定义上下文
	for _, kvs := range customKVs {
		for k, v := range kvs {
			ctx.PutKV(k, v)
		}
	}
	return &ctx
}

// NewContextForTask 用于快速构任务 AOP 上下文
func NewContextForTask(task spec2.PipelineTask, p spec2.Pipeline, trigger aoptypes2.TuneTrigger, customKVs ...map[interface{}]interface{}) *aoptypes2.TuneContext {
	// 先构造 pipeline 上下文
	ctx := NewContextForPipeline(p, trigger, customKVs...)
	// 修改 tune type
	ctx.SDK.TuneType = aoptypes2.TuneTypeTask
	// 注入特有 sdk 属性
	ctx.SDK.Task = task
	return ctx
}
