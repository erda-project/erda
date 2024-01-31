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

package actionagent

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/expression"
	"github.com/erda-project/erda/pkg/retry"
	"github.com/erda-project/erda/pkg/strutil"
)

const EncryptedValueMinLen = "ACTIONAGENT_ENCRYPTED_VAlUE_MIN_LEN"

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
		var err error
		getResp, err = agent.CallbackReporter.GetBootstrapInfo(agent.Arg.PipelineID, agent.Arg.PipelineTaskID)
		if err != nil {
			return err
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

	agent.Arg.Shell = bootstrapArg.Shell
	agent.Arg.Commands = bootstrapArg.Commands
	agent.Arg.Context = bootstrapArg.Context
	agent.Arg.PrivateEnvs = bootstrapArg.PrivateEnvs
	agent.Arg.EncryptSecretKeys = bootstrapArg.EncryptSecretKeys
	agent.Arg.DebugOnFailure = bootstrapArg.DebugOnFailure
	agent.Arg.DebugTimeout = bootstrapArg.DebugTimeout

	valueLen, err := strconv.Atoi(os.Getenv(EncryptedValueMinLen))
	if err != nil || valueLen < 6 {
		valueLen = 6
	}

	for _, v := range agent.Arg.EncryptSecretKeys {
		// the value's len >= EncryptedValueLen will be appended to BlackList
		if value, ok := agent.Arg.PrivateEnvs[strings.ToUpper(v)]; ok && len(value) >= valueLen {
			agent.TextBlackList = append(agent.TextBlackList, value)
		}
	}

	// use env to instead of ${{ envs.xxx }}
	agent.replaceEnvExpr()

	// set envs to current process, so `run` and other scripts can inherit
	for k, v := range agent.Arg.PrivateEnvs {
		if err = os.Setenv(k, v); err != nil {
			agent.AppendError(err)
			return
		}
		if k == apistructs.EnvOpenapiToken {
			agent.CallbackReporter.SetOpenApiToken(v)
			agent.EasyUse.OpenAPIToken = v
		}
	}
}

func (agent *Agent) replaceEnvExpr() {
	for k := range agent.Arg.PrivateEnvs {
		visited := make(map[string]struct{})
		replaceEnvElem(agent.Arg.PrivateEnvs, k, visited)
	}

	if agent.Arg.Commands != nil {
		commands := agent.convertCustomCommands()
		ifReplace := replaceCommandEnvExpr(commands, agent.Arg.PrivateEnvs)
		if ifReplace {
			agent.Arg.Commands = commands
		}
	}

}

// replaceEnvElem parse privateEnvs[elemKey]'s env_expr and returns the ture env and isLoopCall
func replaceEnvElem(privateEnvs map[string]string, elemKey string, visited map[string]struct{}) (string, bool) {
	elem := privateEnvs[elemKey]
	if !strings.HasPrefix(elemKey, EnvActionParamPrefix) {
		return elem, false
	}

	// avoid recursive infinite loop caused by circular reference，such as:
	// ACTION_A: ${{ ACTION_B }}
	// ACTION_B: ${{ ACTION_A }}
	// if it is loop call, return the origin value
	if _, ok := visited[elemKey]; ok {
		return elem, true
	}

	visited[elemKey] = struct{}{}

	isLoop := false

	replaced := strutil.ReplaceAllStringSubmatchFunc(expression.Re, elem, func(sub []string) string {
		inner := sub[1]
		inner = strings.Trim(inner, " ")

		// ss[0] = envs, ss[1] = variable
		ss := strings.SplitN(inner, ".", 2)
		if len(ss) == 1 {
			return elem
		}

		if ss[0] != expression.ENV {
			return elem
		}

		env, ok := privateEnvs[ss[1]]
		if !ok {
			env = os.Getenv(ss[1])
		}

		// check if the parsed environment is still an env_expr
		if expression.Re.MatchString(env) {
			env, isLoop = replaceEnvElem(privateEnvs, ss[1], visited)
			// if the upcoming recursion is a loop invocation
			// do not parse it and returns the current value instead
			if isLoop {
				return elem
			}
		}

		return env
	})

	privateEnvs[elemKey] = replaced
	return replaced, isLoop
}

func replaceCommandEnvExpr(commands []string, privateEnvs map[string]string) bool {
	ifReplace := false
	for index, command := range commands {
		replaced := strutil.ReplaceAllStringSubmatchFunc(expression.Re, command, func(sub []string) string {
			ifReplace = true
			inner := sub[1]
			inner = strings.Trim(inner, " ")

			// ss[0] = envs, ss[1] = variable
			ss := strings.SplitN(inner, ".", 2)
			if len(ss) == 1 {
				return command
			}

			if ss[0] != expression.ENV {
				return command
			}

			var env string
			var ok bool
			if privateEnvs != nil {
				env, ok = privateEnvs[ss[1]]
			}

			if !ok {
				env = os.Getenv(ss[1])
			}

			return env
		})
		commands[index] = replaced
	}

	return ifReplace
}
