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

package aop

import (
	"context"

	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// NewContextForPipeline 用于快速构造流水线 AOP 上下文
func NewContextForPipeline(p spec.Pipeline, trigger aoptypes.TuneTrigger, customKVs ...map[interface{}]interface{}) aoptypes.TuneContext {
	ctx := aoptypes.TuneContext{
		Context: context.Background(),
		SDK:     globalSDK.Clone(),
	}
	ctx.SDK.TuneType = aoptypes.TuneTypePipeline
	ctx.SDK.TuneTrigger = trigger
	ctx.SDK.Pipeline = p
	// 用户自定义上下文
	for _, kvs := range customKVs {
		for k, v := range kvs {
			ctx.PutKV(k, v)
		}
	}
	return ctx
}

// NewContextForTask 用于快速构任务 AOP 上下文
func NewContextForTask(task spec.PipelineTask, p spec.Pipeline, trigger aoptypes.TuneTrigger, customKVs ...map[interface{}]interface{}) aoptypes.TuneContext {
	// 先构造 pipeline 上下文
	ctx := NewContextForPipeline(p, trigger, customKVs...)
	// 修改 tune type
	ctx.SDK.TuneType = aoptypes.TuneTypeTask
	// 注入特有 sdk 属性
	ctx.SDK.Task = task
	return ctx
}
