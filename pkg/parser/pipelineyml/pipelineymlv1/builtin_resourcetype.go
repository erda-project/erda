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

package pipelineymlv1

import (
	"fmt"
)

const DockerImageResType = "docker-image"

type builtinResType string

// builtinResType declaration
// 如有更新，请进入 gogenerate 目录重新运行 go generate 并提交
const (
	RES_TYPE_GIT            builtinResType = "git"
	RES_TYPE_BUILDPACK      builtinResType = "buildpack"
	RES_TYPE_BP_COMPILE     builtinResType = "bp-compile"
	RES_TYPE_BP_IMAGE       builtinResType = "bp-image"
	RES_TYPE_DICE           builtinResType = "dice"
	RES_TYPE_ABILITY        builtinResType = "ability"
	RES_TYPE_ADDON_REGISTRY builtinResType = "addon-registry"
	RES_TYPE_DICEHUB        builtinResType = "dicehub"
	RES_TYPE_IT             builtinResType = "it"
	RES_TYPE_SONAR          builtinResType = "sonar"
	RES_TYPE_FLINK          builtinResType = "flink"
	RES_TYPE_SPARK          builtinResType = "spark"
	RES_TYPE_UT             builtinResType = "ut"
)

func (y *PipelineYml) builtinResTypes() map[string]ResourceType {

	return addBuiltInResTypes(
		y.option.builtinResourceTypeDockerImagePrefix,
		y.option.builtinResourceTypeDockerImageTag,
		BuiltinResTypeNames...,
	)
}

var addBuiltInResTypes = func(imagePrefix, imageTag string, resNames ...string) map[string]ResourceType {
	resTypes := make(map[string]ResourceType)
	for _, resName := range resNames {
		resTypes[resName] = ResourceType{
			Name: resName,
			Type: DockerImageResType,
			Source: Source{
				"repository": fmt.Sprintf("%s/%s-action", imagePrefix, resName),
				"tag":        imageTag,
			},
		}
	}
	return resTypes
}
