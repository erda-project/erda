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

package common

import (
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func GetWorkspaceBranch() (string, error) {
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := branchCmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\r?\n`)
	branch := re.ReplaceAllString(string(out), "")

	return branch, nil
}

func GetOriginRepo() string {
	return GetRepo("origin")
}

func GetRepo(remote string) string {
	out, _ := exec.Command("git", "config", "--get", "remote."+remote+".url").CombinedOutput()
	return string(out)
}

func GetWorkspacePipelines(dir string) ([]string, error) {
	var ymls []string
	fs, err := ioutil.ReadDir(dir + "/pipelines")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	for _, path := range fs {
		if !path.IsDir() && (strings.HasSuffix(path.Name(), ".yml") ||
			strings.HasSuffix(path.Name(), ".yaml")) {
			ymls = append(ymls, path.Name())
		}
	}

	return ymls, nil
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func GetWorkspaceInfo(remoteName string) (org string, project string, app string, err error) {
	remoteCmd := exec.Command("git", "remote", "get-url", remoteName)
	out, err := remoteCmd.CombinedOutput()
	if err != nil {
		return "", "", "", err
	}

	re := regexp.MustCompile(`\r?\n`)
	newStr := re.ReplaceAllString(string(out), "")
	return GetWorkspaceInfoFromErdaRepo(newStr)
}

func GetWorkspaceInfoFromErdaRepo(erdaRepo string) (org, project, app string, err error) {
	u, err := url.Parse(erdaRepo)
	if err != nil {
		return "", "", "", err
	}
	// <org>/dop/<project>/<app>
	paths := strings.Split(u.Path, "/")
	if len(paths) != 5 {
		return "", "", "", errors.New("Invalid Erda Repo Path: " + u.Path)
	}

	return paths[1], paths[3], paths[4], nil
}
