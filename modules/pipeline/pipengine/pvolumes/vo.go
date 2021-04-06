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

package pvolumes

import (
	"path/filepath"
)

const (
	VoLabelKeyContainerPath = "containerPath"
	VoLabelKeyContextPath   = "contextPath"
	VoLabelKeyStageOrder    = "stageOrder"
	VoLabelKeyVolumeID      = "ID"
	VoLabelKeyShareVolume   = "shareVolume"
	VoLabelKeyDiceFileUUID  = "diceFileUUID"
)

var (
	ContainerRootDir     = "/.pipeline/container"
	ContainerContextDir  = filepath.Join(ContainerRootDir, "context")
	ContainerMetadataDir = filepath.Join(ContainerRootDir, "metadata")
	ContainerUploadDir   = filepath.Join(ContainerRootDir, "uploaddir")

	ContainerVolumeMountRootDir = "/.pipeline/context"                  // task volume 的挂载目录的父目录
	ContainerDiceFilesDir       = "/.pipeline/container/cms/dice_files" // cms dice files 类型在运行时的挂载目录
)

// MakeTaskContainerWorkdir 生成 task 在容器内的 workdir 目录
func MakeTaskContainerWorkdir(taskName string) string {
	return filepath.Join(ContainerContextDir, taskName)
}

// MakeTaskContainerMetafilePath 生成 task 在容器内的 metadata 目录
func MakeTaskContainerMetafilePath(taskName string) string {
	return filepath.Join(ContainerMetadataDir, taskName, "metadata")
}

// MakeTaskContainerVolumeMountDir 生成 task 在容器内的 volume 挂载点
func MakeTaskContainerVolumeMountDir(taskName string) string {
	return filepath.Join(ContainerVolumeMountRootDir, taskName)
}

// MakeTaskContainerDiceFilesPath 生成 task 在容器内的文件配置路径
func MakeTaskContainerDiceFilesPath(fileName string) string {
	return filepath.Join(ContainerDiceFilesDir, fileName)
}
