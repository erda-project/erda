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

package cmd

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var GwCreate = command.Command{
	ParentName: "Gw",
	Name:       "create",
	ShortHelp:  "Creates a gateway package",
	LongHelp:   "Creates a gateway package",
	Example:    "erda-cli gw create hub --domain the-hub.erda.cloud",
	Args: []command.Arg{
		command.StringArg{}.Name("scene"),
	},
	Flags: []command.Flag{
		command.StringFlag{
			Name: "org",
			Doc:  Doc("the org name", "组织名称", Required),
		},
		command.IntFlag{
			Name: "project-id",
			Doc:  Doc("the project id", "项目 ID (随便填一个有权限的项目的 ID)", Required),
		},
		command.StringFlag{
			Name: "env",
			Doc:  Doc("The environment to which the traffic package belongs", "流量入口所属的环境", Required),
		},
		command.StringListFlag{
			Name: "domain",
			Doc:  Doc("the package's domain", "流量入口的域名", true),
		},
	},
	Run: RunGwCreate,
}

func RunGwCreate(ctx *command.Context, scene string, orgName string, projectID int, env string, domains []string) error {
	ctx.Info("RunGwCreate, scene: %s, orgName: %s, projectID: %v, env: %s, domains: %v",
		scene, orgName, projectID, env, domains)
	switch scene {
	case "hub":
	default:
		return errors.New("invalid scene, only support hub yet")
	}

	if orgName == "" {
		if err := ctx.FetchOrgs(); err != nil {
			return err
		}
	} else {
		ctx.CurrentOrg.Name = orgName
		_, _, err := common.GetOrgID(ctx, orgName)
		if err != nil {
			return err
		}
	}

	var (
		headers = http.Header{
			"Org-ID":          {strconv.FormatUint(ctx.CurrentOrg.ID, 10)},
			"Org-Name":        {ctx.CurrentOrg.Name},
			"Org":             {ctx.CurrentOrg.Name},
			"User-ID":         {ctx.GetUserID()},
			"Internal-Client": {"erda-cli"},
		}
		body = map[string]interface{}{
			"name":        "hub",
			"description": "-",
			"scene":       scene,
			"bindDomain":  domains,
		}
		request = ctx.UseHepaApi().Post().
			Path("/api/erda-demo/gateway/openapi/packages").
			Param("orgId", strconv.FormatUint(ctx.CurrentOrg.ID, 10)).
			Param("projectId", strconv.Itoa(projectID)).
			Param("env", strings.ToUpper(env)).
			Headers(headers).
			JSONBody(body)
	)
	response, err := request.Do().RAW()
	if err != nil {
		return err
	}
	defer func() {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			ctx.Error("failed to ReadAll: %v", err)
			return
		}
		defer response.Body.Close()
		ctx.Info(string(data))
	}()
	if response.StatusCode >= 300 || response.StatusCode < 200 {
		return errors.New("unexpected response")
	}
	return nil
}
