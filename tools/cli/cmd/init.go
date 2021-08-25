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
	"path/filepath"
	"strings"

	. "github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/dicedir"
	"github.com/erda-project/erda/tools/cli/format"
)

var INIT = Command{
	Name:      "init",
	ShortHelp: "Init a dice.yml template",
	LongHelp:  "Make a .dice dir in current directory, then create a dice.yml template",
	Example: `
  $ dice init
`,
	Run:    RunInit,
	Hidden: true,
}

func RunInit(ctx *Context) error {
	if f, err := os.Stat(dicedir.ProjectDiceDir); err == nil && f.IsDir() {
		return fmt.Errorf(
			format.FormatErrMsg("init", "failed to reinitialize existing dice project", false))
	}

	pdir, err := dicedir.CreateProjectDiceDir()
	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("init", "failed to create project dice directory: "+err.Error(), false))
	}
	diceYmls := []string{
		filepath.Join(pdir, "dice.yml"),
	}
	for _, diceYml := range diceYmls {
		f, err := os.OpenFile(diceYml, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf(
				format.FormatErrMsg("init", "failed to create file "+diceYml, false))
		}
		_, err = f.WriteString(diceymlTemplate)
		if err != nil {
			return fmt.Errorf(
				format.FormatErrMsg("init", "failed to write to file "+diceYml, false))
		}
	}
	ctx.Succ("Init dice project success")
	return nil
}

var diceymlTemplate = strings.TrimSpace(`
# version：版本号
# 可选：NO
# 值：2.0
# 说明：version 字段目前只有 2.0 一个版本可选，如果没有配置 version: 2.0 将走老的 dice.yml 逻辑。
version: 2.0

# envs: 环境变量
# 可选：YES
# 说明：此处的 envs 设置全局环境变量，全局环境变量将被下文定义的所有 services 继承；环境变量采用 "key: value" 形式定义，
# 全局环境并不是必须的。
envs:
  DEBUG: PAMPAS_BLOG
  TERMINUS_TRACE_ENABLE: false

# services: 服务
# 可选：NO
# 说明：services 用于定义一个应用里的所有 service，一个正常应用至少需要定义一个 service。
services:

  # showcase-front 是一个 service 的名字，你需要根据自己的服务在这里填写正确的名字
  showcase-front:

    # image: Docker 镜像
    # 可选: YES
    # 说明：如果使用 Dice 的 CI 平台来构建镜像、一键部署，那么此处就不需要填镜像。
    image: nginx:latest

    # cmd: 服务启动命令
    # 可选：YES
    # 说明：
    cmd: echo hello && npm run start

    # ports: 服务监控端口
    # 可选：YES
    # 说明：可以配置多个端口
    ports:
      - 8080

    # expose: 服务导出
    # 可选：YES
    # 说明：服务需要被导出的时候，就需要写上 expose 描述，expose 只能导出 80 和 443 端口，也就是服务最终被用户访问的是
    # 80 或者 443 端口。一般需要被导出的服务都是终端服务，也就是直接面向用户访问的服务，后端服务一般来说不需要导出。
    # expose 端口的流量会被转发到 ports 里的第一个端口上。
    expose:
      - 80
      - 443

    # envs: 环境变量
    # 可选：YES
    # 说明：service 内的 envs 用于定义服务私有的环境变量，可以覆盖全局环境变量中的定义。
    envs:
      TERMINUS_TRACE_ENABLE: false
      TERMINUS_APP_NAME: showcase-front-app

    # hosts：host 绑定
    # 可选：YES
    # 说明：hosts 配置将被写入 /etc/hosts 文件。
    hosts:
      - 127.0.0.1 www.terminus.io

    # resources: 资源
    # 可选：NO
    # 说明：
    resources:
      cpu: 0.2
      # 单位 MB
      mem: 256
      # 单位 MB
      disk: 100

    # volumes: 
    # 可选：YES
    # 说明：
    volumes:
      - /home/admin/logs

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

      # labels: 部署标签
      # 可选：YES
      labels:
        a: b

    # depends_on: 服务依赖
    # 可选：YES
    # 说明：depends_on 用于描述一个应用里的服务相互依赖关系，被依赖的 service 必须被定义，最终所有 service 不能
    # 出现循环依赖。
    depends_on:
      - blog-service

    # health_check: 健康检查
    # 可选：NO
    # 说明：支持2种方式: http, exec 
    health_check: 
      # 可选：YES
      http:
        port: 80
        path: /status
        duration: 120
      # 可选：YES
      exec:
        cmd: curl http:127.0.0.1:7070/status
        duration: 120

  # blog-service 示范了最简洁的服务配置，也就是只配置了最必须的配置项。
  blog-service:
    resources:
      cpu: 0.2
      # 单位 MB
      mem: 256
      # 单位 MB
      disk: 100
    deployments:
      replicas: 1

# addons: 附加资源
# 可选：YES
# 说明：
addons:
    mysql: 
      # plan:
      # <addon类型>:<规格>
      # 可选规格:
      # 1. basic
      # 2. professional
      # 3. ultimate
      plan: mysql:basic
      # 各种 addon 有不同的 option 
      options:
        version: 5.7.23
        create_dbs: blog,boxxx2

    zk:
      plan: zookeeper:medium

# 以下字段描述各个环境各自配置
environments:
  development:
  test:
  staging:
  production:

`)
