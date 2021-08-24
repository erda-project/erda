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

package pipelinesvc

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/expression"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func (s *PipelineSvc) RunPipeline(req *apistructs.PipelineRunRequest) (*spec.Pipeline, error) {

	p, err := s.dbClient.GetPipeline(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrGetPipeline.InvalidParameter(err)
	}

	if req.UserID != "" {
		p.Extra.RunUser = s.tryGetUser(req.UserID)
	}
	if req.InternalClient != "" {
		p.Extra.InternalClient = req.InternalClient
	}

	reason, canManualRun := s.canManualRun(p)
	if !canManualRun {
		return nil, apierrors.ErrRunPipeline.InvalidState(reason)
	}
	if req.ForceRun {
		err := s.stopRunningPipelines(&p, req.IdentityInfo)
		if err != nil {
			return nil, err
		}
	}

	// 校验已运行的 pipeline
	if err := s.limitParallelRunningPipelines(&p); err != nil {
		return nil, err
	}

	// cms
	secrets, cmsDiceFiles, holdOnKeys, encryptSecretKeys, err := s.FetchSecrets(&p)
	if err != nil {
		return nil, apierrors.ErrRunPipeline.InternalError(err)
	}

	// replace global config use same random value
	for k, v := range secrets {
		secrets[k] = expression.ReplaceRandomParams(v)
	}

	// 校验私有配置转换出来的 envs
	secretsEnvs := make(map[string]string)
	for k, v := range secrets {
		newK := strings.Replace(strings.Replace(strings.ToUpper(k), ".", "_", -1), "-", "_", -1)
		secretsEnvs[newK] = v
	}
	errs := pipelineyml.CheckEnvs(secretsEnvs)
	if len(errs) > 0 {
		var errMsgs []string
		for _, checkErr := range errs {
			errMsgs = append(errMsgs, checkErr.Error())
		}
		return nil, apierrors.ErrCheckSecrets.InvalidParameter(strutil.Join(errMsgs, "\n", true))
	}

	// 获取平台级别配置
	platformSecrets, err := s.FetchPlatformSecrets(&p, holdOnKeys)
	if err != nil {
		return nil, apierrors.ErrRunPipeline.InternalError(err)
	}

	// Snapshot 快照用于记录
	p.Snapshot.PipelineYml = p.PipelineYml
	p.Snapshot.Secrets = secrets
	p.Snapshot.PlatformSecrets = platformSecrets
	p.Snapshot.CmsDiceFiles = cmsDiceFiles
	// pipeline 运行时的参数
	runParams, err := getRealRunParams(req.PipelineRunParams, p.PipelineYml)
	if err != nil {
		return nil, err
	}
	p.Snapshot.RunPipelineParams = runParams.ToPipelineRunParamsWithValue()
	p.Snapshot.EncryptSecretKeys = encryptSecretKeys

	now := time.Now()
	p.TimeBegin = &now

	// update pipeline base
	if err := s.dbClient.UpdatePipelineBase(p.ID, &p.PipelineBase); err != nil {
		return nil, apierrors.ErrUpdatePipeline.InternalError(err)
	}

	// update pipeline extra
	if err := s.dbClient.UpdatePipelineExtraByPipelineID(p.ID, &p.PipelineExtra); err != nil {
		return nil, apierrors.ErrRunPipeline.InternalError(err)
	}

	// aop
	_ = aop.Handle(aop.NewContextForPipeline(p, aoptypes.TuneTriggerPipelineBeforeExec))

	// send to pipengine reconciler
	s.engine.Send(p.ID)

	return &p, nil
}

func getRealRunParams(runParams []apistructs.PipelineRunParam, yml string) (result apistructs.PipelineRunParams, err error) {

	pipeline, err := pipelineyml.New([]byte(yml))
	if err != nil {
		return nil, apierrors.ErrRunPipeline.InternalError(err)
	}

	var runParamsMap = make(map[string]apistructs.PipelineRunParam)
	if runParams != nil {
		for _, runParam := range runParams {
			runParamsMap[runParam.Name] = runParam
		}
	}

	// 获取真实的运行时参数
	var realParamsMap = make(map[string]interface{})
	for _, param := range pipeline.Spec().Params {
		// 用户没有传 key, 且默认值不为空
		runValue, ok := runParamsMap[param.Name]

		if runValue.Value == nil && param.Default == nil && param.Required && ok {
			return nil, apierrors.ErrRunPipeline.InternalError(fmt.Errorf("pipeline param %s value is empty", param.Name))
		}

		if runValue.Value == nil && param.Default != nil {
			realParamsMap[param.Name] = param.Default
		}

		if runValue.Value == nil && param.Default == nil {
			realParamsMap[param.Name] = pipelineyml.GetParamDefaultValue(param.Type)
		}

		if runValue.Value != nil {
			realParamsMap[param.Name] = runValue.Value
		}

		if realParamsMap[param.Name] == nil {
			return nil, apierrors.ErrRunPipeline.InternalError(fmt.Errorf("pipeline param %s value is empty", param.Name))
		}
	}

	for key, v := range realParamsMap {
		result = append(result, apistructs.PipelineRunParam{
			Name:  key,
			Value: v,
		})
	}

	return result, nil
}

func (s *PipelineSvc) stopRunningPipelines(p *spec.Pipeline, identityInfo apistructs.IdentityInfo) error {
	var runningPipelineIDs []uint64
	err := s.dbClient.Table(&spec.PipelineBase{}).
		Select("id").In("status", apistructs.ReconcilerRunningStatuses()).
		Where("is_snippet = ?", false).
		Find(&runningPipelineIDs, &spec.PipelineBase{
			PipelineSource:  p.PipelineSource,
			PipelineYmlName: p.PipelineYmlName,
		})
	if err != nil {
		return apierrors.ErrParallelRunPipeline.InternalError(err)
	}
	for _, runningPipelineID := range runningPipelineIDs {
		if err := s.Cancel(&apistructs.PipelineCancelRequest{
			PipelineID:   runningPipelineID,
			IdentityInfo: identityInfo,
		}); err != nil {
			return err
		}
	}
	return nil
}

// limitParallelRunningPipelines 判断在 pipelineSource + pipelineYmlName 下只能有一个在运行
// 被嵌套的流水线跳过校验
func (s *PipelineSvc) limitParallelRunningPipelines(p *spec.Pipeline) error {
	// 流水线自身是嵌套流水线时，不做校验
	if p.IsSnippet {
		return nil
	}
	var runningPipelineIDs []uint64
	err := s.dbClient.Table(&spec.PipelineBase{}).
		Select("id").In("status", apistructs.ReconcilerRunningStatuses()).
		Where("is_snippet = ?", false).
		Find(&runningPipelineIDs, &spec.PipelineBase{
			PipelineSource:  p.PipelineSource,
			PipelineYmlName: p.PipelineYmlName,
		})
	if err != nil {
		return apierrors.ErrParallelRunPipeline.InternalError(err)
	}
	if len(runningPipelineIDs) > 0 {
		ctxMap := map[string]interface{}{
			apierrors.ErrParallelRunPipeline.Error(): fmt.Sprintf("%d", runningPipelineIDs[0]),
		}
		logrus.Infof("start run pipeline %d", runningPipelineIDs[0])
		return apierrors.ErrParallelRunPipeline.InvalidState("ErrParallelRunPipeline").SetCtx(ctxMap)
	}
	return nil
}

func makeCmsDiceFileEnvKey(diceFileUUID string) string {
	return fmt.Sprintf("PIPELINE_CMS_DICE_FILE_UUID_%s", diceFileUUID)
}
