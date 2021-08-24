package plugins_manage

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	. "github.com/erda-project/erda/modules/pipeline/providers/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/services/reportsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

type PluginsManage struct {
	// tuneGroup 保存所有 tune chain
	// 根据 类型、触发时机 初始化所有场景下的调用链
	tuneGroup   TuneGroup
	once        sync.Once
	initialized bool
	globalSDK   SDK
}

var pluginsManage = &PluginsManage{}

func (p *PluginsManage) Handle(ctx *TuneContext) error {
	if !p.initialized {
		return fmt.Errorf("AOP not initialized")
	}
	typ := ctx.SDK.TuneType
	trigger := ctx.SDK.TuneTrigger
	logrus.Debugf("AOP: type: %s, trigger: %s", typ, trigger)
	chain := p.tuneGroup.GetTuneChainByTypeAndTrigger(typ, trigger)
	if chain == nil || len(chain) == 0 {
		logrus.Debugf("AOP: type: %s, trigger: %s, tune chain is empty", typ, trigger)
		return nil
	}
	err := chain.Handle(ctx)
	if err != nil {
		logrus.Errorf("AOP: type: %s, trigger: %s, handle failed, err: %v", typ, trigger, err)
		return err
	}
	logrus.Debugf("AOP: type: %s, trigger: %s, handle success", typ, trigger)
	return nil
}

func InitPluginsManage() *PluginsManage {
	pluginsManage.once.Do(func() {
		db, err := dbclient.New()
		if err != nil {
			panic(err)
		}
		pluginsManage.initialized = true
		pluginsManage.globalSDK.Bundle = bundle.New(bundle.WithAllAvailableClients())
		pluginsManage.globalSDK.DBClient = db
		pluginsManage.globalSDK.Report = reportsvc.New(reportsvc.WithDBClient(db))
	})
	return pluginsManage
}

func (p *PluginsManage) GetTuneGroup() TuneGroup {
	if p.tuneGroup == nil {
		p.tuneGroup = make(TuneGroup)
	}

	// sort tunePoint by rank
	for tuneType := range p.tuneGroup {
		for tuneTrigger := range p.tuneGroup[tuneType] {
			sort.Sort(p.tuneGroup[tuneType][tuneTrigger])
		}
	}
	return p.tuneGroup
}

func RegisterTunePointToTuneGroup(tuneType TuneType, tuneTrigger TuneTrigger, tunePoint TunePoint) error {
	if pluginsManage.tuneGroup == nil {
		pluginsManage.tuneGroup = make(map[TuneType]map[TuneTrigger]TuneChain)
	}

	group, ok := pluginsManage.tuneGroup[tuneType]
	if !ok {
		group = make(map[TuneTrigger]TuneChain)
	}
	group[tuneTrigger] = append(group[tuneTrigger], tunePoint)
	pluginsManage.tuneGroup[tuneType] = group
	return nil
}

// NewContextForPipeline 用于快速构造流水线 AOP 上下文
func (p *PluginsManage) NewContextForPipeline(pi spec.Pipeline, trigger TuneTrigger, customKVs ...map[interface{}]interface{}) *TuneContext {
	ctx := TuneContext{
		Context: context.Background(),
		SDK:     p.globalSDK.Clone(),
	}
	ctx.SDK.TuneType = TuneTypePipeline
	ctx.SDK.TuneTrigger = trigger
	ctx.SDK.Pipeline = pi
	// 用户自定义上下文
	for _, kvs := range customKVs {
		for k, v := range kvs {
			ctx.PutKV(k, v)
		}
	}
	return &ctx
}

// NewContextForTask 用于快速构任务 AOP 上下文
func (p *PluginsManage) NewContextForTask(task spec.PipelineTask, pi spec.Pipeline, trigger TuneTrigger, customKVs ...map[interface{}]interface{}) *TuneContext {
	// 先构造 pipeline 上下文
	ctx := p.NewContextForPipeline(pi, trigger, customKVs...)
	// 修改 tune type
	ctx.SDK.TuneType = TuneTypeTask
	// 注入特有 sdk 属性
	ctx.SDK.Task = task
	return ctx
}
