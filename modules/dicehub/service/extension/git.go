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

package extension

import (
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/erda-project/erda/apistructs"
)

func (i *Extension) GetExtensionByGit(name, d string, file ...string) (*apistructs.ExtensionVersion, error) {
	files, err := getGitFileContents(d, file...)
	if err != nil {
		return nil, err
	}

	return &apistructs.ExtensionVersion{
		Name:      name,
		Version:   "1.0",
		Type:      "action",
		Spec:      files[0],
		Dice:      files[1],
		Swagger:   "",
		Readme:    files[2],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDefault: false,
		Public:    true,
	}, nil
}

func getGitFileContents(d string, file ...string) ([]string, error) {
	var resp []string
	// dirName is a random string
	dir, err := ioutil.TempDir(os.TempDir(), "*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	// git init
	command := exec.Command("sh", "-c", "git init")
	command.Dir = dir
	err = command.Run()
	if err != nil {
		return nil, err
	}

	// git remote
	remoteCmd := "git remote add -f origin " + d
	command = exec.Command("sh", "-c", remoteCmd)
	command.Dir = dir
	err = command.Run()
	if err != nil {
		return nil, err
	}

	// set git config
	command = exec.Command("sh", "-c", "git config core.sparsecheckout true")
	command.Dir = dir
	err = command.Run()
	if err != nil {
		return nil, err
	}

	// set sparse-checkout
	for _, v := range file {
		echoCmd := "echo " + v + " >> .git/info/sparse-checkout"
		command = exec.Command("sh", "-c", echoCmd)
		command.Dir = dir
		err = command.Run()
		if err != nil {
			return nil, err
		}
	}

	// git pull
	command = exec.Command("sh", "-c", "git pull origin master")
	command.Dir = dir
	err = command.Run()
	if err != nil {
		return nil, err
	}

	// read .yml
	for _, v := range file {
		f, err := os.Open(dir + "/" + v)
		if err != nil {
			resp = append(resp, "")
			continue
		}
		str, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		f.Close()
		resp = append(resp, string(str))
	}

	return resp, nil
}
