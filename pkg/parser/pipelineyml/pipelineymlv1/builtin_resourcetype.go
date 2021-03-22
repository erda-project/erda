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
