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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/terminal/color_str"
	"github.com/erda-project/erda/tools/cli/status"
	"github.com/erda-project/erda/tools/cli/utils"
)

var ctx = Context{
	Sessions: make(map[string]status.StatusInfo),
}

func GetContext() *Context {
	return &ctx
}

type Context struct {
	Sessions           map[string]status.StatusInfo
	CurrentHost        string
	Domain             *url.URL
	Openapi            *url.URL
	Hepaapi            *url.URL
	CurrentOrg         OrgInfo
	CurrentProject     ProjectInfo
	CurrentApplication ApplicationInfo
	Applications       []ApplicationInfo
	Debug              bool
	Token              string // uc token
	HttpClient         *httpclient.HTTPClient
}

func (ctx *Context) UseDomain() *Context {
	var ctx2 = *ctx
	ctx2.CurrentHost = ctx2.Domain.String()
	return &ctx2
}

func (ctx *Context) UseOpenapi() *Context {
	var ctx2 = *ctx
	ctx2.CurrentHost = ctx2.Openapi.String()
	return &ctx2
}

func (ctx *Context) UseHepaApi() *Context {
	var ctx2 = *ctx
	ctx2.CurrentHost = ctx2.Hepaapi.String()
	return &ctx2
}

func (ctx *Context) Get() *httpclient.Request {
	return ctx.wrapRequest(ctx.HttpClient.Get)
}

func (ctx *Context) Post() *httpclient.Request {
	return ctx.wrapRequest(ctx.HttpClient.Post)
}

func (ctx *Context) Put() *httpclient.Request {
	return ctx.wrapRequest(ctx.HttpClient.Put)
}

func (ctx *Context) Patch() *httpclient.Request {
	return ctx.wrapRequest(ctx.HttpClient.Patch)
}

func (ctx *Context) Delete() *httpclient.Request {
	return ctx.wrapRequest(ctx.HttpClient.Delete)
}

type MetadataResponse struct {
	OpenapiPublicUrl string                  `json:"openapi_public_url,omitempty"`
	Version          MetadataResponseVersion `json:"version,omitempty"`
}

type MetadataResponseVersion struct {
	Built       string `json:"built"`
	DiceVersion string `json:"dice_version"`
	GitCommit   string `json:"git_commit"`
	GoVersion   string `json:"go_version"`
}

func (ctx *Context) FetchOpenapi() error {
	domain, err := url.Parse(ctx.CurrentHost)
	if err != nil {
		return errors.Wrap(err, "invalid host")
	}
	ctx.Domain = domain

	var resp MetadataResponse
	response, err := ctx.Get().Path("/metadata.json").Do().RAW()
	if err != nil {
		return err
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err = json.Unmarshal(data, &resp); err != nil {
		return err
	}
	if response.StatusCode >= 300 || response.StatusCode < 200 {
		return errors.Errorf("Unexpected response returned from erda when querying openapi address: %s", response.Status)
	}
	u, err := url.Parse(resp.OpenapiPublicUrl)
	if err != nil {
		return err
	}
	ctx.Openapi = u
	ctx.Hepaapi = new(url.URL)
	*ctx.Hepaapi = *ctx.Openapi
	ctx.Hepaapi.Host = strings.Replace(ctx.Openapi.Host, "openapi.", "hepa.", 1)
	ctx.CurrentHost = ctx.Openapi.String()
	ctx.Info(`erda openapi info:
	domain: %s
	openapi: %s
	hepa: %s
	built: %s
	dice_version: %s
	git_commit: %s
	go_version: %s
	`, ctx.Domain.String(), ctx.Openapi.String(), ctx.Hepaapi.String(), resp.Version.Built, resp.Version.DiceVersion, resp.Version.GitCommit, resp.Version.GoVersion)
	return nil
}

func (ctx *Context) FetchOrgs() error {
	var resp BaseResponse
	request := ctx.UseDomain().Get().Path("/api/-/orgs")
	response, err := request.Do().JSON(&resp)
	if err != nil {
		return errors.Wrapf(err, "failed to Get %s", request.GetUrl())
	}
	if !response.IsOK() {
		return errors.Errorf("failed to Get %s: response is not ok: %s", request.GetUrl(), string(response.Body()))
	}
	var orgs OrgsResponseData
	if err := json.Unmarshal(resp.Data, &orgs); err != nil {
		return err
	}
	if len(orgs.List) == 0 {
		ctx.Info("no organization found for the user")
		return nil
	}
	inputS := "Choose your organization ID or name:\n"
	for i := 0; i < len(orgs.List); i++ {
		inputS += strconv.FormatUint(orgs.List[i].ID, 10) + ": " + orgs.List[i].Name + "\n"
	}
	orgIDOrName := utils.InputNormal(inputS)
	if orgIDOrName == "" {
		ctx.Info("You have not chosen any organization name: %s", orgIDOrName)
		return nil
	}
	orgIDOrName = strings.TrimSpace(orgIDOrName)
	for i := 0; i < len(orgs.List); i++ {
		if item := orgs.List[i]; strconv.FormatUint(item.ID, 10) == orgIDOrName || item.Name == orgIDOrName {
			ctx.CurrentOrg = item
			ctx.Info(`Your organization:
	OrgID: %v
	OrgName: %s
	Description: %s`,
				item.ID, item.Name, item.Desc)
			return nil
		}
	}
	return errors.Errorf("You have chosen an invalid organization: %s", strconv.Quote(orgIDOrName))
}

func (ctx *Context) wrapRequest(m func(host string, retry ...httpclient.RetryOption) *httpclient.Request) *httpclient.Request {
	req := m(ctx.CurrentHost)
	req.Header("USE-TOKEN", "true")
	if ctx.Token != "" {
		req.Header("Authorization", ctx.Token)
	}
	if v, ok := ctx.Sessions[ctx.CurrentHost]; ok && v.SessionID != "" {
		req.Cookie(&http.Cookie{Name: "OPENAPISESSION", Value: ctx.Sessions[ctx.CurrentHost].SessionID})
	}
	return req
}

func (ctx *Context) GetUserID() string {
	if sessionInfo, ok := ctx.Sessions[ctx.CurrentHost]; ok {
		return sessionInfo.ID
	}

	return ""
}

func (ctx *Context) Info(format string, a ...interface{}) {
	f := "[INFO] " + strings.TrimSuffix(format, "\n") + "\n"
	fmt.Printf(f, a...)
}

func (ctx *Context) Warn(format string, a ...interface{}) {
	f := "[WARN] " + strings.TrimSuffix(format, "\n") + "\n"
	fmt.Printf(f, a...)
}

func (ctx *Context) Error(format string, a ...interface{}) {
	f := "[ERROR] " + strings.TrimSuffix(format, "\n") + "\n"
	fmt.Printf(f, a...)
}

func (ctx *Context) Succ(format string, a ...interface{}) {
	f := color_str.Green("✔ ") + strings.TrimSuffix(format, "\n") + "\n"
	fmt.Printf(f, a...)
}

func (ctx *Context) Fail(format string, a ...interface{}) {
	f := color_str.Red("✗ ") + strings.TrimSuffix(format, "\n") + "\n"
	fmt.Printf(f, a...)
}
