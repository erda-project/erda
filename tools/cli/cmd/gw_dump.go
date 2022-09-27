// Copyright (c) 2022 Terminus, Inc.
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

package cmd

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

const (
	Required = true
	Optional = false
)

var GwDump = command.Command{
	ParentName: "Gw",
	Name:       "dump",
	ShortHelp:  "Dumps a package's endpoints\\n下载一个流量入口的全部路由",
	LongHelp:   "Dumps a package's endpoints\\n下载一个流量入口的全部路由",
	Example:    "erda-cli gw dump --pkg-id xxx -o xxx.json",
	Flags: []command.Flag{
		command.StringFlag{
			Name: "pkg-id",
			Doc:  Doc("the package's id", "流量入口的 id", Required),
		},
		command.StringFlag{
			Short: "o",
			Name:  "output",
			Doc:   Doc("output json filename", "输出文件", Optional),
		},
		command.StringFlag{
			Name: "org",
			Doc:  Doc("the org name", "组织名称", Required),
		},
	},
	Run: RunGwDump,
}

func RunGwDump(ctx *command.Context, pkgId, output, orgName string) error {
	ctx.Info("RunGwDump")
	if pkgId == "" {
		return errors.New("invalid --pkg-id")
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
		uri     = "/api/gateway/openapi/packages/" + pkgId + "/apis"
		headers = http.Header{
			"Org-ID":          {strconv.FormatUint(ctx.CurrentOrg.ID, 10)},
			"Org-Name":        {ctx.CurrentOrg.Name},
			"Org":             {ctx.CurrentOrg.Name},
			"User-ID":         {ctx.GetUserID()},
			"Internal-Client": {"erda-cli"},
		}
		request = ctx.UseHepaApi().Get().
			Path(uri).
			Param("pageNo", "1").
			Headers(headers)
		resp ListPackageAPIsResponse
	)

	response, err := request.Param("pageSize", "1").Do().JSON(&resp)
	if err != nil {
		return err
	}
	if response.StatusCode() < 200 || response.StatusCode() >= 300 {
		ctx.Error("response.StatusCode: %v, response.Body: %s", response.StatusCode(), string(response.Body()))
		return errors.Errorf("unexpected response from hepa openapi: %s", request.GetUrl())
	}
	if !resp.Success {
		return errors.New("response fails")
	}
	if resp.Data.Total == 0 {
		ctx.Info("no api to output")
		return nil
	}

	var out = os.Stdout
	if output != "" {
		file, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			ctx.Error("failed to Openfile, filename: %s, err: %v", output, err)
			return err
		}
		defer file.Close()
		out = file
	}

	if resp.Data.Total == 1 {
		_, err := out.Write(response.Body())
		if err != nil {
			ctx.Error("failed to out.Write, err: %v", err)
		}
		return err
	}

	response, err = request.SetParam("pageSize", strconv.Itoa(resp.Data.Total)).Do().Body(out)
	if err != nil {
		return err
	}
	if response.StatusCode() < 200 || response.StatusCode() >= 300 {
		return errors.New("unexpected response from erda openapi")
	}
	if !resp.Success {
		return errors.New("response fails")
	}

	return nil
}

type ListPackageAPIsResponse struct {
	Success bool                        `json:"success"`
	Data    ListPackageAPIsResponseData `json:"data"`
}

type ListPackageAPIsResponseData struct {
	List  []ListPackageAPIsResponseItem `json:"list"`
	Total int                           `json:"total"`
}

type ListPackageAPIsResponseItem struct {
	AllowPassAuth       bool        `json:"allowPassAuth"`
	ApiId               string      `json:"apiId"`
	ApiPath             string      `json:"apiPath"`
	CreateAt            string      `json:"createAt"`
	Description         string      `json:"description"`
	DiceApp             string      `json:"diceApp"`
	DiceService         string      `json:"diceService"`
	Hosts               interface{} `json:"hosts"`
	Mutable             bool        `json:"mutable"`
	Origin              string      `json:"origin"`
	RedirectAddr        string      `json:"redirectAddr"`
	RedirectApp         string      `json:"redirectApp"`
	RedirectPath        string      `json:"redirectPath"`
	RedirectRuntimeId   string      `json:"redirectRuntimeId"`
	RedirectRuntimeName string      `json:"redirectRuntimeName"`
	RedirectService     string      `json:"redirectService"`
	RedirectType        string      `json:"redirectType"`
}

func Doc(en, zh string, required bool) string {
	var s = "[Optional] "
	if required {
		s = "[Required] "
	}
	return s + en + "\\n" + strings.Repeat(" ", len(s)) + zh
}
