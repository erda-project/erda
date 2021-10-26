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
	"net/url"
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

func GetWorkspaceInfo() (org string, project string, app string, err error) {
	remoteCmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := remoteCmd.CombinedOutput()
	if err != nil {
		return "", "", "", err
	}

	re := regexp.MustCompile(`\r?\n`)
	newStr := re.ReplaceAllString(string(out), "")
	u, err := url.Parse(newStr)
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
