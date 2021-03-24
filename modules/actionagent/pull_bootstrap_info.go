package actionagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/retry"
)

func (agent *Agent) pullBootstrapInfo() {
	if !agent.Arg.PullBootstrapInfo {
		return
	}

	tokenForBootstrap := os.Getenv(apistructs.EnvOpenapiTokenForActionBootstrap)
	if tokenForBootstrap == "" {
		agent.AppendError(errors.Errorf("missing env %s", apistructs.EnvOpenapiTokenForActionBootstrap))
		return
	}
	agent.EasyUse.TokenForBootstrap = tokenForBootstrap

	var getResp apistructs.PipelineTaskGetBootstrapInfoResponse
	err := retry.DoWithInterval(func() error {
		var body bytes.Buffer
		r, err := httpclient.New(httpclient.WithCompleteRedirect()).
			Get(agent.EasyUse.OpenAPIAddr).
			Path(fmt.Sprintf("/api/pipelines/%d/tasks/%d/actions/get-bootstrap-info", agent.Arg.PipelineID, agent.Arg.PipelineTaskID)).
			Header("Authorization", tokenForBootstrap).
			Do().
			Body(&body)
		if err != nil {
			return err
		}
		if !r.IsOK() {
			return errors.Errorf("status-code: %d, resp body: %s", r.StatusCode(), body.String())
		}
		if err := json.NewDecoder(&body).Decode(&getResp); err != nil {
			return errors.Errorf("status-code: %d, failed to json unmarshal get-bootstrap-resp, err: %v", r.StatusCode(), err)
		}
		return nil
	}, 5, time.Second*5)
	if err != nil {
		agent.AppendError(errors.Errorf("failed to get bootstrap info, err: %v", err))
		return
	}

	var bootstrapArg AgentArg
	if err := json.Unmarshal(getResp.Data.Data, &bootstrapArg); err != nil {
		agent.AppendError(err)
		return
	}

	agent.Arg.Commands = bootstrapArg.Commands
	agent.Arg.Context = bootstrapArg.Context
	agent.Arg.PrivateEnvs = bootstrapArg.PrivateEnvs

	// set envs to current process, so `run` and other scripts can inherit
	for k, v := range agent.Arg.PrivateEnvs {
		if err = os.Setenv(k, v); err != nil {
			agent.AppendError(err)
			return
		}
		if k == apistructs.EnvOpenapiToken {
			agent.EasyUse.OpenAPIToken = v
		}
	}
}
