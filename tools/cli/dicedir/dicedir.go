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

package dicedir

import (
	"errors"
	"os"
	"os/user"
	"path"
	"path/filepath"
)

var (
	NotExist = errors.New("not exist")

	GlobalErdaDir = ".erda.d"
	GlobalConfig  = path.Join(GlobalErdaDir, "config")

	ProjectDiceDir     = ".dice"
	ProjectPipelineDir = ".dice/pipelines"

	ProjectErdaDir = ".erda"
)

func FindGlobalConfig() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	file := filepath.Join(u.HomeDir, GlobalConfig)
	f, err := os.Stat(file)
	if os.IsNotExist(err) {
		return file, NotExist
	}
	if f.IsDir() {
		return file, errors.New(file + " is a dirctory")
	}
	return file, nil
}

func FindGlobalDiceDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(u.HomeDir, GlobalErdaDir)
	f, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return "", NotExist
	}
	if !f.IsDir() {
		return "", NotExist
	}
	return dir, nil
}

func CreateGlobalDiceDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(u.HomeDir, GlobalErdaDir)
	if err := os.Mkdir(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func FindProjectDiceDir() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	var res string
	for {
		if existProjDiceDir(current) {
			res = mkProjDiceDirPath(current)
			return res, nil
		}
		origin := current
		current = filepath.Dir(current)
		if current == origin {
			return "", NotExist
		}
	}
}

func CreateProjectDiceDir() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	pdir := mkProjDiceDirPath(current)
	if err := os.Mkdir(pdir, 0755); err != nil {
		return "", err
	}
	return pdir, nil
}

func existProjDiceDir(path string) bool {
	f, err := os.Stat(mkProjDiceDirPath(path))
	if os.IsNotExist(err) {
		return false
	}
	return f.IsDir()
}

func mkProjDiceDirPath(path string) string {
	return filepath.Join(path, ProjectDiceDir)
}

func FindProjectErdaDir() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	var res string
	for {
		if existProjErdaDir(current) {
			res = mkProjErdaDirPath(current)
			return res, nil
		}
		origin := current
		current = filepath.Dir(current)
		if current == origin {
			return "", NotExist
		}
	}
}

func CreateProjectErdaDir() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	pdir := mkProjErdaDirPath(current)
	if err := os.Mkdir(pdir, 0755); err != nil && !os.IsExist(err) {
		return "", err
	}
	return pdir, nil
}

func existProjErdaDir(path string) bool {
	f, err := os.Stat(mkProjErdaDirPath(path))
	if os.IsNotExist(err) {
		return false
	}
	return f.IsDir()
}

func mkProjErdaDirPath(path string) string {
	return filepath.Join(path, ProjectErdaDir)
}
