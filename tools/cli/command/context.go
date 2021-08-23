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

package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/terminal/color_str"
	"github.com/erda-project/erda/tools/cli/dicedir"
	"github.com/erda-project/erda/tools/cli/format"
	"github.com/erda-project/erda/tools/cli/status"
)

var ctx = Context{
	Sessions: make(map[string]status.StatusInfo),
}

func GetContext() *Context {
	return &ctx
}

type Context struct {
	Sessions           map[string]status.StatusInfo
	CurrentOpenApiHost string
	Debug              bool
	Token              string // uc token
	HttpClient         *httpclient.HTTPClient
}
type OrgInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func (c *Context) Get() *httpclient.Request {
	return c.wrapRequest(ctx.HttpClient.Get)
}

func (c *Context) Post() *httpclient.Request {
	return c.wrapRequest(ctx.HttpClient.Post)
}

func (c *Context) Put() *httpclient.Request {
	return c.wrapRequest(ctx.HttpClient.Put)
}

func (c *Context) Patch() *httpclient.Request {
	return c.wrapRequest(ctx.HttpClient.Patch)
}

func (c *Context) Delete() *httpclient.Request {
	return c.wrapRequest(ctx.HttpClient.Delete)
}

func (c *Context) wrapRequest(m func(host string, retry ...httpclient.RetryOption) *httpclient.Request) *httpclient.Request {
	req := m(c.CurrentOpenApiHost)
	req.Header("USE-TOKEN", "true")
	if c.Token != "" {
		req.Header("Authorization", c.Token)
	}
	if v, ok := c.Sessions[c.CurrentOpenApiHost]; ok && v.SessionID != "" {
		req.Cookie(&http.Cookie{Name: "OPENAPISESSION", Value: c.Sessions[c.CurrentOpenApiHost].SessionID})
	}
	return req
}

// 当前企业可能不存在，因为可能不在任何企业内，所以返回的OrgInfo可能为空
func (c *Context) CurrentOrg() (apistructs.OrgDTO, error) {
	var resp apistructs.OrgFetchResponse
	var b bytes.Buffer

	v, ok := c.Sessions[c.CurrentOpenApiHost]
	if !ok {
		return apistructs.OrgDTO{}, errors.Errorf("failed to find session for %s", c.CurrentOpenApiHost)
	}
	response, err := ctx.Get().Path(fmt.Sprintf("/api/orgs/%d", v.OrgID)).Do().JSON(&resp)
	if err != nil {
		return apistructs.OrgDTO{}, fmt.Errorf(
			format.FormatErrMsg("orgs", "failed to request ("+err.Error()+")", false))
	}

	// TODO: check 404 or any other known errors
	if !response.IsOK() {
		return apistructs.OrgDTO{}, fmt.Errorf(format.FormatErrMsg("orgs",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if !resp.Success {
		return apistructs.OrgDTO{}, fmt.Errorf(format.FormatErrMsg("orgs",
			fmt.Sprintf("error code(%s), error message(%s)", resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}

func (c *Context) AvailableOrgs() ([]apistructs.OrgDTO, error) {
	var resp apistructs.OrgSearchResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/orgs").Do().Body(&b)
	if err != nil {
		return nil, fmt.Errorf(
			format.FormatErrMsg("orgs", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return nil, fmt.Errorf(format.FormatErrMsg("orgs",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf(format.FormatErrMsg("orgs",
			fmt.Sprintf("failed to unmarshal organizations list response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return nil, fmt.Errorf(format.FormatErrMsg("orgs",
			fmt.Sprintf("error code(%s), error message(%s)", resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data.List, nil
}

// 项目 .dice 目录下的 dice.yml
func defaultYml(env string) (string, error) {
	pdir, err := dicedir.FindProjectDiceDir()
	if err != nil {
		return "", fmt.Errorf(format.FormatErrMsg("get default dice.yml",
			"find dice dir of current project error: "+err.Error(), false))
	}
	var envfilename string
	switch env {
	case "dev":
		envfilename = "dice_development.yml"
	case "test":
		envfilename = "dice_test.yml"
	case "staging":
		envfilename = "dice_staging.yml"
	case "prod":
		envfilename = "dice_production.yml"
	default:
		envfilename = "dice.yml"
	}
	ymlPath := filepath.Join(pdir, envfilename)

	return ymlPath, nil
}

func defaultYmlCheckExist(env string) (string, error) {
	path, err := defaultYml(env)
	if err != nil {
		return "", err
	}
	f, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if f.IsDir() {
		return "", fmt.Errorf(
			format.FormatErrMsg("check default dice.yml exist", "dice.yml can not be a dir", false))
	}
	return path, nil
}

func (c *Context) DiceYml(checkExist bool) (string, error) {
	if checkExist {
		return defaultYmlCheckExist("")
	}
	return defaultYml("")
}
func (c *Context) DevDiceYml(checkExist bool) (string, error) {
	if checkExist {
		return defaultYmlCheckExist("dev")
	}
	return defaultYml("dev")
}
func (c *Context) TestDiceYml(checkExist bool) (string, error) {
	if checkExist {
		return defaultYmlCheckExist("test")
	}
	return defaultYml("test")
}
func (c *Context) StagingDiceYml(checkExist bool) (string, error) {
	if checkExist {
		return defaultYmlCheckExist("staging")
	}
	return defaultYml("staging")
}
func (c *Context) ProdDiceYml(checkExist bool) (string, error) {
	if checkExist {
		return defaultYmlCheckExist("prod")
	}
	return defaultYml("prod")
}

func (c *Context) Succ(format string, a ...interface{}) {
	f := color_str.Green("✔ ") + strings.TrimSuffix(format, "\n") + "\n"
	fmt.Printf(f, a...)
}

func (c *Context) Fail(format string, a ...interface{}) {
	f := color_str.Red("✗ ") + strings.TrimSuffix(format, "\n") + "\n"
	fmt.Printf(f, a...)
}

type Release struct {
	ReleaseID   string    `json:"releaseId"`
	ReleaseName string    `json:"releaseName"`
	Labels      string    `json:"labels"`
	OrgId       int64     `json:"orgId"`       // 所属企业
	ClusterName string    `json:"clusterName"` // 所属集群
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ReleaseResponse struct {
	Success bool      `json:"success"`
	Data    []Release `json:"data"`
}
