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

package actionagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/retry"
)

const EncryptedValueLen = "ACTIONAGENT_ENCRYPTED_VAlUE_LEN"

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
	agent.Arg.EncryptSecretKeys = bootstrapArg.EncryptSecretKeys

	valueLen, err := strconv.Atoi(os.Getenv(EncryptedValueLen))
	if err != nil || valueLen == 0 {
		valueLen = 6
	}

	for _, v := range agent.Arg.EncryptSecretKeys {
		// the value's len >= EncryptedValueLen will be appended to BlackList
		if value, ok := agent.Arg.PrivateEnvs[strings.ToUpper(v)]; ok && len(value) >= valueLen {
			agent.BlackList = append(agent.BlackList, value)
		}
	}

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
