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

package scene_before

import (
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func (p *Plugin) Name() string {
	return "scene_before"
}

func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {
	// source = autotest
	if ctx.SDK.Pipeline.PipelineSource == apistructs.PipelineSourceAutoTest && !ctx.SDK.Pipeline.IsSnippet {
		if strings.HasPrefix(ctx.SDK.Pipeline.PipelineYmlName, apistructs.PipelineSourceAutoTestPlan.String()+"-") {
			return nil
		}
		sceneID, err := strconv.ParseUint(ctx.SDK.Pipeline.PipelineYmlName, 10, 64)
		if err != nil {
			return err
		}
		var req apistructs.AutotestSceneRequest
		req.SceneID = sceneID
		req.UserID = ctx.SDK.Pipeline.PipelineExtra.Snapshot.PlatformSecrets["dice.user.id"]
		scene, err := ctx.SDK.Bundle.GetAutoTestScene(req)
		if err != nil {
			return err
		}
		req2 := apistructs.AutotestSceneSceneUpdateRequest{
			SceneID:     scene.ID,
			Description: scene.Description,
			Status:      apistructs.ProcessingSceneStatus,
			IsStatus:    true,
		}
		req2.UserID = req.UserID
		_, err = ctx.SDK.Bundle.UpdateAutoTestScene(req2)
		if err != nil {
			return err
		}
	}
	return nil
}

func New() *Plugin {
	var p Plugin
	return &p
}

type config struct {
	TuneType    aoptypes.TuneType      `file:"tune_type"`
	TuneTrigger []aoptypes.TuneTrigger `file:"tune_trigger" `
}

// +provider
type provider struct {
	Cfg *config
}

func (p *provider) Init(ctx servicehub.Context) error {
	for _, tuneTrigger := range p.Cfg.TuneTrigger {
		err := plugins_manage.RegisterTunePointToTuneGroup(p.Cfg.TuneType, tuneTrigger, New())
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func init() {
	servicehub.Register("erda.core.pipeline.aop.plugins.pipeline.scene-before", &servicehub.Spec{
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
