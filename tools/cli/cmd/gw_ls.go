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
	"encoding/json"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/tools/cli/command"
)

var GwLs = command.Command{
	ParentName: "Gw",
	Name:       "ls",
	ShortHelp:  "Erda Gateway list all invalid endpoints",
	LongHelp:   "Erda Gateway list all invalid endpoints",
	Example:    "erda-cli gw ls --invalid-only --cluster erda-cloud -o erda-cloud.invalid-endpoints.json",
	Flags: []command.Flag{
		command.BoolFlag{
			Short:        "",
			Name:         "invalid-only",
			Doc:          "[Required] --invalid-only must be specified",
			DefaultValue: false,
		},
		command.StringFlag{
			Short:        "C",
			Name:         "cluster",
			Doc:          "[Required] the cluster name must be specified",
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "o",
			Name:         "output",
			Doc:          "[Optional] the output file should be specified",
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "",
			Name:         "hepa",
			Doc:          "[Optional] hepa address like https://hepa.erda.cloud",
			DefaultValue: "",
		},
	},
	Run: RunGwLs,
}

type BaseResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Err     BaseResponseErr `json:"err"`
}

type BaseResponseData struct {
	List  []command.OrgInfo `json:"list"`
	Total int               `json:"total"`
}

type BaseResponseErr struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Ctx  interface{} `json:"ctx"`
}

func RunGwLs(context *command.Context, invalidOnly bool, cluster, output, hepa string) error {
	if !invalidOnly {
		return errors.New("--invalid-only must be specified")
	}
	if cluster == "" {
		return errors.New("cluster name must be specified")
	}
	if output == "" {
		output = cluster + "." + time.Now().Format("2006-01-02_150405") + ".invalid-endpoints.json"
	}

	file, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var ctx = *context
	// to find an org
	if ctx.CurrentOrg.Name == "" {
		var orgResp BaseResponse
		response, err := ctx.Get().
			Path("/api/orgs").
			Do().
			JSON(&orgResp)
		if err != nil {
			return err
		}
		if !response.IsOK() {
			return errors.New(string(response.Body()))
		}

		var orgData BaseResponseData
		if err = json.Unmarshal(orgResp.Data, &orgData); err != nil {
			return err
		}
		if len(orgData.List) == 0 {
			return errors.Errorf("no org found for the user: %s", ctx.GetUserID())
		}

		ctx.CurrentOrg = orgData.List[0]
	}
	ctx.Info("Org-ID: %v, Org-Name: %v, User-ID: %v", ctx.CurrentOrg.ID, ctx.CurrentOrg.Name, ctx.GetUserID())

	// generate hepa host
	if hepa == "" {
		host, err := url.Parse(ctx.CurrentOpenApiHost)
		if err != nil {
			return err
		}
		host.Host = "hepa." + strings.TrimPrefix(host.Host, "openapi.")
		hepa = host.String()
	}
	ctx.CurrentOpenApiHost = hepa
	ctx.Info("HEPA host: %s", ctx.CurrentOpenApiHost)

	// get invalid endpoints
	response, err := ctx.Get().
		Path("/api/gateway/openapi/invalid-endpoints").
		Param("clusterName", cluster).
		Header("Org-ID", strconv.FormatUint(ctx.CurrentOrg.ID, 10)).
		Header("User-ID", ctx.GetUserID()).
		Header("Internal-Client", "erda-cli").
		Do().
		Body(file)
	if err != nil {
		return err
	}
	if !response.IsOK() {
		return errors.Errorf("success: %v, status: %v, message: %v", response.IsOK(), response.StatusCode(), string(response.Body()))
	}
	ctx.Info("result is writen into %s", output)
	return nil
}
