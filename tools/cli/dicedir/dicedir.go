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

package dicedir

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
)

var (
	NotExist = errors.New("not exist")

	GlobalDiceDir  = ".dice.d"
	ProjectDiceDir = ".dice"
)

func FindGlobalDiceDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(u.HomeDir, GlobalDiceDir)
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
	dir := filepath.Join(u.HomeDir, GlobalDiceDir)
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
