// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package version

import (
	"fmt"
	"os"
)

var (
	// Version 版本号
	Version string
	// BuildTime 编译时间
	BuildTime string
	// GoVersion 编译时间
	GoVersion string
	// CommitID 版本库中的提交版本
	CommitID string
	// DockerImage Image地址
	DockerImage string
)

// String 返回版本信息
func String() string {
	return fmt.Sprintf("Version: %s\nBuildTime: %s\nGoVersion: %s\nCommitID: %s\nDockerImage: %s\n",
		Version, BuildTime, GoVersion, CommitID, DockerImage)
}

// Print 打印版本信息
func Print() {
	fmt.Print(String())
}

// PrintIfCommand 如果命令行有version参数，则打印并退出
func PrintIfCommand() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		Print()
		os.Exit(0)
	}
}
