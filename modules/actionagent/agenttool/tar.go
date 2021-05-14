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

package agenttool

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"

	"github.com/erda-project/erda/pkg/retry"
)

// Tar is the forth step of reource agent do
//
// 可选方案：
// a. tar 目标地址直接指向 NFS
// b. 在容器内制作 tar，移动至 NFS
// 经过测试，选用 a 速度更快
func Tar(tarFile, tarDir string) error {
	tarExecDir := filepath.Dir(tarFile)
	if err := os.Chdir(tarExecDir); err != nil {
		return err
	}
	if err := tar(tarFile, tarDir); err != nil {
		return err
	}
	return nil
}

func UnTar(tarFile, unTarDir string) error {
	err := retry.Do(func() error {
		return unTar(tarFile, unTarDir)
	}, 3)
	return err
}

func tar(tarAbsPath string, mainSrc string, otherSrcs ...string) error {
	if IsTarCommandExist() {
		args := []string{"-cf", tarAbsPath}
		args = append(args, getTarSrcArg(mainSrc)...)
		for _, src := range otherSrcs {
			args = append(args, getTarSrcArg(src)...)
		}
		cmd := exec.Command("tar", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			os.RemoveAll(tarAbsPath)
			return archiver.Tar.Make(tarAbsPath, append([]string{mainSrc}, otherSrcs...))
		}
		return err
	}
	return archiver.Tar.Make(tarAbsPath, append([]string{mainSrc}, otherSrcs...))
}

func unTar(tarAbsPath string, destDir string) error {
	if IsTarCommandExist() {
		os.Mkdir(destDir, os.ModePerm)
		cmd := exec.Command("tar", "-xf", tarAbsPath, "-C", destDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			os.RemoveAll(destDir)
			return archiver.Tar.Open(tarAbsPath, destDir)
		}
		return err
	}
	return archiver.Tar.Open(tarAbsPath, destDir)
}

func getTarSrcArg(sourcePath string) []string {
	parentIndex := strings.LastIndex(sourcePath, "/")
	if parentIndex == -1 {
		return []string{sourcePath}
	}
	//去除绝对路径中的父路径
	pathName := filepath.Base(sourcePath)
	return []string{"-C", sourcePath[0:parentIndex], pathName}
}
func IsTarCommandExist() bool {
	cmdName := "tar"
	pathEnv := os.Getenv("PATH")
	pathEnvList := strings.Split(pathEnv, ":")
	for _, pathItem := range pathEnvList {
		fullCmdName := path.Join(pathItem, cmdName)
		if _, err := os.Stat(fullCmdName); err == nil {
			return true
		}
	}
	return false
}
