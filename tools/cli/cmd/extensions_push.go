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

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

var EXTENSIONSPUSH = command.Command{
	Name:       "push",
	ParentName: "EXT",
	ShortHelp:  "push extension",
	Example: `
  $ dice ext push -f --public
`,
	Flags: []command.Flag{
		command.BoolFlag{Short: "f", Name: "force", Doc: "override exist version", DefaultValue: false},
		command.BoolFlag{Short: "a", Name: "all", Doc: "override exist extension and version,must with -f", DefaultValue: false},
		command.StringFlag{Short: "d", Name: "dir", Doc: "extension dir", DefaultValue: ""},
		command.StringFlag{Short: "r", Name: "registry", Doc: "new registry", DefaultValue: ""},
	},
	Run: RunExtensionsPush,
}

func RunExtensionsPush(ctx *command.Context, force bool, all bool, dir, registry string) error {
	var request apistructs.ExtensionVersionCreateRequest
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if dir != "" {
		workDir = dir
	}

	specBytes, err := ioutil.ReadFile(path.Join(workDir, "spec.yml"))
	if err != nil {
		return err
	}
	specData := apistructs.Spec{}
	err = yaml.Unmarshal(specBytes, &specData)
	if err != nil {
		return errors.Wrap(err, "failed to parse spec")
	}

	request.Public = specData.Public
	request.ForceUpdate = force
	request.Name = specData.Name
	request.Version = specData.Version
	request.SpecYml = string(specBytes)
	request.All = all
	request.IsDefault = specData.IsDefault

	diceBytes, err := ioutil.ReadFile(path.Join(workDir, "dice.yml"))

	if registry != "" && err == nil {
		diceBytes, _, err = replaceDiceRegistry(diceBytes, specData.Type, registry)
		if err != nil {
			return fmt.Errorf("%s ext_dir:%s", err, workDir)
		}
	}

	if err == nil {
		request.DiceYml = string(diceBytes)
	}

	readmeBytes, err := ioutil.ReadFile(path.Join(workDir, "README.md"))
	if err == nil {
		request.Readme = string(readmeBytes)
	}
	swaggerBytes, err := ioutil.ReadFile(path.Join(workDir, "swagger.yml"))
	if err == nil {
		request.SwaggerYml = string(swaggerBytes)
	}

	err = pushExtension(ctx, request)
	if err != nil {
		return err
	}

	ctx.Succ(fmt.Sprintf("extension %s push success\n", specData.Name))

	return nil
}

func pushExtension(ctx *command.Context, request apistructs.ExtensionVersionCreateRequest) error {
	var resp apistructs.ExtensionVersionCreateResponse
	var b bytes.Buffer
	urlPath := "/api/extensions/" + request.Name
	response, err := ctx.Post().Path(urlPath).JSONBody(request).Do().Body(&b)
	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("extension push", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return fmt.Errorf(format.FormatErrMsg("extension push",
			fmt.Sprintf("failed to request, status-code: %d %s",
				response.StatusCode(), b.String()), false))
	}

	if err = json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf(format.FormatErrMsg("extension push",
			fmt.Sprintf("failed to unmarshal build extension response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return fmt.Errorf(format.FormatErrMsg("extension push",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}
	return nil
}
