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

package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/dicedir"
	"github.com/erda-project/erda/tools/cli/format"
)

var ERDAINIT = command.Command{
	Name:       "init",
	ParentName: "ERDA",
	ShortHelp:  "Init a erda.yml template",
	LongHelp:   "Make a .erda dir in current directory, then create a erda.yml template",
	Example:    "$ erda init",
	Flags: []command.Flag{
		command.FloatFlag{Short: "c", Name: "cpu",
			Doc:          "the quota of CPU for service",
			DefaultValue: 0.5},
		command.IntFlag{Short: "m", Name: "memory",
			Doc:          "the quota of Memory for service",
			DefaultValue: 1024},
	},
	Run:    ErdaInit,
	Hidden: true,
}

func ErdaInit(ctx *command.Context, cpuQuota float64, memQuota int) error {
	// TODO
	if _, err := os.Stat(path.Join(dicedir.ProjectErdaDir, "erda.yml")); err == nil {
		return fmt.Errorf(
			format.FormatErrMsg("init", "failed to reinitialize existing erda project", false))
	}

	pdir, err := dicedir.CreateProjectErdaDir()
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf(
			format.FormatErrMsg("init", "failed to create project dice directory: "+err.Error(), false))
	}

	p, err := common.ParseSpringBoot()
	if err != nil {
		return err
	}
	p["ServiceCPU"] = fmt.Sprintf("%f", cpuQuota)
	p["ServiceMemory"] = strconv.Itoa(memQuota)

	// TODO more erda-xxx.yml
	erdaYmls := []string{
		filepath.Join(pdir, "erda.yml"),
	}
	for _, erdaYml := range erdaYmls {
		f, err := os.OpenFile(erdaYml, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf(
				format.FormatErrMsg("init", "failed to create file "+erdaYml, false))
		}

		t := template.New("p")
		t.Parse(erdaymlTemplate)
		err = t.Execute(f, p)
		if err != nil {
			return fmt.Errorf(
				format.FormatErrMsg("init", "failed to write to file "+erdaYml, false))
		}
	}
	ctx.Succ("Init .erda/erda.yml success")
	return nil
}

var erdaymlTemplate = strings.TrimSpace(`
# version：版本号
# 可选：NO
# 值：2.0
# 说明：version 字段目前只有 2.0 一个版本可选，如果没有配置 version: 2.0 将走老版本逻辑。
version: 2.0

# envs: 环境变量
# 可选：YES
# 说明：此处的 envs 设置全局环境变量，全局环境变量将被下文定义的所有 services 继承；环境变量采用 "key: value" 形式定义，
# 全局环境并不是必须的。
envs:

# services: 服务
# 可选：NO
# 说明：services 用于定义一个应用里的所有 service，一个正常应用至少需要定义一个 service。
services:

  # 一个 service 的名字，你需要根据自己的服务在这里填写正确的名字
  {{.ServiceName}}:
    # ports: 服务监控端口
    # 可选：YES
    # 说明：可以配置多个端口
    ports:
      - {{.ServicePort}}

    # envs: 环境变量
    # 可选：YES
    # 说明：service 内的 envs 用于定义服务私有的环境变量，可以覆盖全局环境变量中的定义。
    envs:

    # resources: 资源
    # 可选：NO
    # 说明：
    resources:
      cpu: {{.ServiceCPU}}
      # 单位 MB
      mem: {{.ServiceMemory}}

    # deployments: 部署
    # 可选：NO
    # 说明：deployments 定义服务的部署配置
    deployments:

      # replicas: 实例数
      # 可选：NO
      # 说明：replicas 不能小于 1，replicas 定义了服务需要被部署几个容器实例。
      replicas: 1

      # policies: 部署策略
      # 可选：YES
      # 说明：可配置的值只有如下三种：shuffle, affinity, unique；shuffle 表示打散实例，避免部署在同一宿主机上，shuffle 也是
      # 默认值，一般都不需要配置此项；affinity 表示实例采用亲和部署，也就是部署到同机上；unique 表示唯一部署，也就是一个宿主机
      # 只能部署一个实例。
      policies: shuffle

    # health_check: 健康检查
    # 可选：NO
    # 说明：支持2种方式: http, exec
    # 建议配置应用的 健康检查 API
    # health_check:
    #   http:
    #     # port 配置健康检查 http get 请求的端口，此端口应该是 ports 配置中的一个
    #     port: 8080
    #     # path 配置健康检查 http get 的 URI 路径
    #     path: /health
    #     # duration 设定健康检查需要持续的检查的时间，单位是秒；这个时间值应该设置为比服务启动需要的时间更长一点
    #     duration: 120

# addons: 附加资源
# 可选：YES
addons:

# 以下字段描述各个环境各自配置
environments:
  development:
  test:
  staging:
  production:
`)
