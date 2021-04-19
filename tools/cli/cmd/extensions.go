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
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

var EXT = command.Command{
	Name:      "ext",
	ShortHelp: "Extensions operation sets,including search, pull, push, retag",
	Example:   `dice ext`,
	Flags: []command.Flag{
		command.BoolFlag{Short: "a", Name: "all", Doc: "query all extensions", DefaultValue: false},
	},
	Run: RunExtensions,
}

func RunExtensions(ctx *command.Context, all bool) error {
	var resp apistructs.ExtensionQueryResponse
	var b bytes.Buffer
	urlPath := "/api/extensions"
	response, err := ctx.Get().Path(urlPath).
		Do().Body(&b)

	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("extension list", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return fmt.Errorf(format.FormatErrMsg("extension list",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err = json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf(format.FormatErrMsg("extension list",
			fmt.Sprintf("failed to unmarshal extension list response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return fmt.Errorf(format.FormatErrMsg("extension list",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	data := [][]string{}
	for _, ext := range resp.Data {
		data = append(data, []string{
			ext.Name,
			ext.Type,
			ext.Category,
			strconv.FormatBool(ext.Public),
			ext.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return table.NewTable().Header([]string{"id", "type", "category", "public", "updated_at"}).Data(data).Flush()
}

func getRegistryAuths() (map[string]DockerAuth, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	dockerConfigPath := path.Join(usr.HomeDir, ".docker")
	dockerConfigFilePath := path.Join(dockerConfigPath, "config.json")
	if _, err := os.Stat(dockerConfigPath); os.IsNotExist(err) {
		err := os.MkdirAll(dockerConfigPath, 0755)
		if err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(dockerConfigFilePath); os.IsNotExist(err) {
		err = ioutil.WriteFile(dockerConfigFilePath, []byte("{}"), 0755)
		if err != nil {
			return nil, err
		}
	}

	fileBytes, err := ioutil.ReadFile(dockerConfigFilePath)
	if err != nil {
		return nil, err
	}
	var dockerConfig DockerConfig
	err = json.Unmarshal(fileBytes, &dockerConfig)
	if err != nil {
		return nil, err
	}
	if dockerConfig.Auths == nil {
		dockerConfig.Auths = map[string]DockerAuth{}
	}
	return dockerConfig.Auths, nil
}

type DockerConfig struct {
	Auths map[string]DockerAuth `json:"auths"`
}

type DockerAuth struct {
	Auth string `json:"auth"`
}

func checkRegistryLogin(auths map[string]DockerAuth, images map[string]string) {
	for _, image := range images {
		registry := getImageRegistry(image)
		auth, ok := auths[registry]
		if !ok || auth.Auth == "" {
			fmt.Printf("%s login\n ", registry)
			username := inputNormal("username: ")
			password := inputPWD("password: ")
			auths[registry] = DockerAuth{
				Auth: base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
			}
		}
	}
}

func saveRegistryAuth(auths map[string]DockerAuth) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	dockerConfigPath := path.Join(usr.HomeDir, ".docker")
	dockerConfigFilePath := path.Join(dockerConfigPath, "config.json")
	var config map[string]interface{}

	fileBytes, err := ioutil.ReadFile(dockerConfigFilePath)
	if err != nil {
		return err
	}
	json.Unmarshal(fileBytes, &config)

	config["auths"] = auths
	resultBytes, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dockerConfigFilePath, resultBytes, 0755)
}

func getImageRegistry(image string) string {
	index := strings.Index(image, "/")
	if index > 0 {
		return image[0:index]
	}
	return image
}

func updateDockerAuth(images map[string]string) error {
	auths, err := getRegistryAuths()
	if err != nil {
		return err
	}
	checkRegistryLogin(auths, images)
	return saveRegistryAuth(auths)
}

func inputPWD(prompt string) string {
	cmd := exec.Command("stty", "-echo")
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	defer func() {
		cmd := exec.Command("stty", "echo")
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		fmt.Println("")
	}()
	return inputNormal(prompt)
}

func inputNormal(prompt string) string {
	fmt.Printf(prompt)
	r := bufio.NewReader(os.Stdin)
	input, err := r.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return input[:len(input)-1]
}
