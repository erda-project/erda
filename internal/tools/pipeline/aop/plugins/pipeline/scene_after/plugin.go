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

package scene_after

import (
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/aop"
	aoptypes2 "github.com/erda-project/erda/internal/tools/pipeline/aop/aoptypes"
)

// +provider
type provider struct {
	aoptypes2.PipelineBaseTunePoint
}

func (p *provider) Name() string { return "scene-after" }

func (p *provider) Handle(ctx *aoptypes2.TuneContext) error {
	// source = autotest
	if ctx.SDK.Pipeline.PipelineSource == apistructs.PipelineSourceAutoTest && !ctx.SDK.Pipeline.IsSnippet {
		if strings.HasPrefix(ctx.SDK.Pipeline.PipelineYmlName, apistructs.PipelineSourceAutoTestPlan.String()+"-") {
			return nil
		}
		sceneID, err := strconv.ParseUint(ctx.SDK.Pipeline.PipelineYmlName, 10, 64)
		if err != nil {
			return err
		}
		var status apistructs.SceneStatus
		if ctx.SDK.Pipeline.Status.IsSuccessStatus() {
			status = apistructs.SuccessSceneStatus
		}
		if ctx.SDK.Pipeline.Status.IsFailedStatus() {
			status = apistructs.ErrorSceneStatus
		}
		if ctx.SDK.Pipeline.Status.IsReconcilerRunningStatus() {
			status = apistructs.ProcessingSceneStatus
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
			Status:      status,
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

func (p *provider) Init(ctx servicehub.Context) error {
	err := aop.RegisterTunePoint(p)
	if err != nil {
		panic(err)
	}
	return nil
}

func init() {
	servicehub.Register(aop.NewProviderNameByPluginName(&provider{}), &servicehub.Spec{
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
