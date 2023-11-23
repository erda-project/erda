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

package task

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/core/pipeline/task/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/pipeline/actionagent"
	"github.com/erda-project/erda/internal/tools/pipeline/conf"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/debug"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline_register"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/permission"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/pipeline"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
)

type taskService struct {
	p        *provider
	dbClient *dbclient.Client
	bdl      *bundle.Bundle

	pipelineSvc  pipeline.Interface
	permission   permission.Interface
	edgeRegister edgepipeline_register.Interface
}

func (s *taskService) PipelineTaskDetail(ctx context.Context, req *pb.PipelineTaskDetailRequest) (*pb.PipelineTaskDetailResponse, error) {
	p, err := s.pipelineSvc.Detail(ctx, req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineTaskDetail.InternalError(err)
	}
	task, err := s.TaskDetail(req.TaskID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineTaskDetail.InternalError(err)
	}
	if task.PipelineID != p.ID {
		return nil, apierrors.ErrGetPipelineTaskDetail.InvalidParameter("task not belong to pipeline")
	}

	identityInfo := apis.GetIdentityInfo(ctx)
	// check whether the user has GET permission under the corresponding branch of the application
	if err := s.permission.CheckBranch(identityInfo, p.Labels[apistructs.LabelAppID], p.Labels[apistructs.LabelBranch], apistructs.GetAction); err != nil {
		return nil, apierrors.ErrGetPipelineTaskDetail.AccessDenied()
	}
	return &pb.PipelineTaskDetailResponse{Data: task.Convert2PB()}, nil
}

func (s *taskService) PipelineTaskGetBootstrapInfo(ctx context.Context, req *pb.PipelineTaskGetBootstrapInfoRequest) (*pb.PipelineTaskGetBootstrapInfoResponse, error) {
	p, err := s.pipelineSvc.Detail(ctx, req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrGetTaskBootstrapInfo.InternalError(err)
	}
	task, err := s.TaskDetail(req.TaskID)
	if err != nil {
		return nil, apierrors.ErrGetTaskBootstrapInfo.InternalError(err)
	}
	if task.PipelineID != p.ID {
		return nil, apierrors.ErrGetTaskBootstrapInfo.InvalidParameter("task not belong to pipeline")
	}

	// only action-agent will use the token method to call this interface
	// and the verification has been completed by openapi checkToken

	// get openapi oauth2 token for callback platform
	_, err = s.GetOpenapiOAuth2TokenForActionInvokeOpenapi(task)
	if err != nil {
		return nil, apierrors.ErrGetTaskBootstrapInfo.InternalError(err)
	}
	breakpointConfig := debug.MergeBreakpoint(task.Extra.Breakpoint, p.Extra.Breakpoint)
	debugTimeout, err := debug.ParseDebugTimeout(breakpointConfig.Timeout)
	if err != nil {
		return nil, apierrors.ErrGetTaskBootstrapInfo.InvalidParameter("debug timeout")
	}
	// bootstrapInfoData
	bootstrapInfo := actionagent.AgentArg{
		Shell:             task.Extra.Action.Shell,
		Commands:          task.Extra.Action.Commands,
		Context:           task.Context,
		PrivateEnvs:       task.Extra.PrivateEnvs,
		EncryptSecretKeys: task.Extra.EncryptSecretKeys,
		DebugTimeout:      debugTimeout,
	}
	if breakpointConfig.On != nil {
		bootstrapInfo.DebugOnFailure = breakpointConfig.On.Failure
	}
	b, err := json.Marshal(&bootstrapInfo)
	if err != nil {
		return nil, apierrors.ErrGetTaskBootstrapInfo.InternalError(err)
	}

	var bootstrapInfoData pb.PipelineTaskGetBootstrapInfoResponseData
	bootstrapInfoData.Data = b

	return &pb.PipelineTaskGetBootstrapInfoResponse{
		Data: &bootstrapInfoData,
	}, nil
}

func (s *taskService) TaskDetail(taskID uint64) (*spec.PipelineTask, error) {
	task, err := s.dbClient.GetPipelineTask(taskID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineTaskDetail.InternalError(err)
	}
	return &task, nil
}

func (s *taskService) GetOpenapiOAuth2TokenForActionInvokeOpenapi(task *spec.PipelineTask) (*apistructs.OAuth2Token, error) {
	var tokenInfo *apistructs.OAuth2Token
	var err error
	req := apistructs.OAuth2TokenGetRequest{
		ClientID:     conf.OpenapiOAuth2TokenClientID(),
		ClientSecret: conf.OpenapiOAuth2TokenClientSecret(),
		Payload:      task.Extra.OpenapiOAuth2TokenPayload,
	}
	if s.edgeRegister.IsEdge() {
		tokenInfo, err = s.edgeRegister.GetOAuth2Token(req)
	} else {
		tokenInfo, err = s.bdl.GetOAuth2Token(req)
	}
	if err != nil {
		return nil, apierrors.ErrGetOpenapiOAuth2Token.InternalError(err)
	}
	if task.Extra.PrivateEnvs == nil {
		task.Extra.PrivateEnvs = make(map[string]string)
	}
	task.Extra.PrivateEnvs[apistructs.EnvOpenapiToken] = tokenInfo.AccessToken
	// store tokenInfo into task
	if err := s.dbClient.UpdatePipelineTaskExtra(task.ID, task.Extra); err != nil {
		logrus.Errorf("[alert] failed to update pipeline task extra to add %s, pipelineID: %d, taskID: %d, err: %v",
			apistructs.EnvOpenapiToken, task.PipelineID, task.ID, err)
	}
	return tokenInfo, nil
}
