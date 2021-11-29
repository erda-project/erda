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
	"strings"
	"text/template"

	"github.com/erda-project/erda/tools/cli/dicedir"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var PIPELINEINIT = command.Command{
	Name:       "init",
	ParentName: "PIPELINE",
	ShortHelp:  "Init pipelines in .dice/pipelines directory (current repo)",
	Example:    "$ erda-cli pipeline init -f .dice/pipelines/pipeline.yml",
	Flags: []command.Flag{
		command.StringFlag{"f", "filename",
			"Specify the path of pipeline.yml file, default: .dice/pipelines/pipeline.yml",
			""},
	},
	Run: PipelineInit,
}

func PipelineInit(ctx *command.Context, ymlfile string) error {
	var filepath string
	if ymlfile == "" {
		filepath = path.Join(dicedir.ProjectPipelineDir, "pipeline.yml")
	} else {
		filepath = ymlfile
	}

	_, err := os.Stat(filepath)
	if os.IsExist(err) {
		fmt.Println(filepath, "is already exist")
		err = PipelineCheck(ctx, filepath)
		if err != nil {
			fmt.Println(filepath, "not valid!")
			return err
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(dicedir.ProjectPipelineDir, 0755)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	t := template.New("p")
	t.Parse(ymlTemplate)

	p, err := common.ParseSpringBoot()
	if err != nil {
		return err
	}

	y, err := getErdaYamls()
	if err != nil {
		return err
	}

	p["ErdaYamls"] = y

	err = t.Execute(f, p)
	if err != nil {
		return err
	}

	ctx.Succ(fmt.Sprintf("Init %s success", filepath))
	return nil
}

func getErdaYamls() (string, error) {
	if _, err := os.Stat(".erda/erda.yml"); err != nil {
		return "", err
	}

	yamls := []string{"dice_yml: ${git-checkout}/.erda/erda.yml"}
	for _, env := range []string{"dev", "test", "staging", "prod"} {
		if _, err := os.Stat(".erda/erda_" + env + ".yml"); err == nil {
			// TODO format is Critical!!
			s := fmt.Sprintf("        dice_%s_yml: ${git-checkout}/erda_%s.yml", env, env)
			yamls = append(yamls, s)
		}
	}

	return strings.Join(yamls, "\n"), nil
}

var ymlTemplate = `version: 1.1
stages:
- stage:
  - git-checkout:

- stage:
  - java:
      # 缓存对应目录，下次构建就可以加速
      caches:
        - path: /root/.m2/repository
      params:
        build_type: maven
        #打包时的工作目录，此路径一般为根 pom.xml 的路径。
        # ${git-checkout} 表示引用上一个 stage 流程里的输出结果，如有别名则使用别名表示
        workdir: ${git-checkout}
        # 打包产物，一般为 jar，填写相较于 workdir 的相对路径。文件必须存在，否则将会出错。
        target: ./target/{{.ServiceTargetName}}.jar
        # 运行 target（如 jar）所需的容器类型，比如这里我们打包的结果是 spring-boot 的 fat jar，故使用 spring-boot container
        container_type: spring-boot

- stage:
  - release:
      params:
        {{.ErdaYamls}}
        image:
          {{.ServiceName}}: ${java:OUTPUT:image}

- stage:
  - dice:
      params:
        release_id: ${release:OUTPUT:releaseID}
`
