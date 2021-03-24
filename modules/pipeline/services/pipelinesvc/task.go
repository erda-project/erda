package pipelinesvc

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (s *PipelineSvc) TaskDetail(taskID uint64) (*spec.PipelineTask, error) {
	task, err := s.dbClient.GetPipelineTask(taskID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineTaskDetail.InternalError(err)
	}
	return &task, nil
}

func (s *PipelineSvc) GetOpenapiOAuth2TokenForActionInvokeOpenapi(task *spec.PipelineTask) (*apistructs.OpenapiOAuth2Token, error) {
	tokenInfo, err := s.bdl.GetOpenapiOAuth2Token(apistructs.OpenapiOAuth2TokenGetRequest{
		ClientID:     conf.OpenapiOAuth2TokenClientID(),
		ClientSecret: conf.OpenapiOAuth2TokenClientSecret(),
		Payload:      task.Extra.OpenapiOAuth2TokenPayload,
	})
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
