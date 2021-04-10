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

package template

//func ChangeString() {
//	aa := "\ncron: \"*/10 * * * *\"\n\n\ncron_compensator:\n  enable: true\n  latest_first: true\n\n\nversion: \"1.0\"\nstages:\n  - stage:\n      - git-checkout:\n          alias: git-checkout\n          resources:\n            cpu: 0.2\n            mem: 512\n          params:\n            depth: 1\n\n  - stage:\n      - java-build:\n          alias: java-build\n          version: \"1.0\"\n          resources:\n            cpu: 0.2\n            mem: 512\n          params:\n            build_cmd:\n               - \"mvn clean install -Dmaven.test.skip=true\"\n              \n            jdk_version: 8\n            workdir: ${git-checkout}\n\n  - stage:\n      - release:\n          params:\n            dice_yml: ${git-checkout}/dice.yml\n            services:\n              java-demo:\n                # 图形界面基于选择，当然可以自选\n                image: openjdk:8-jre-alpine\n                copys:\n                  - ${java-build:OUTPUT:buildPath}/target/docker-java-app-example.jar:/target\n                  - ${java-build:OUTPUT:buildPath}/spot-agent/spot-agent.jar:/spot-agent\n                cmd: java ${java-build:OUTPUT:JAVA_OPTS} -jar /target/docker-java-app-example.jar\n\n  - stage:\n      - dice:\n          params:\n            release_id: ${release:OUTPUT:releaseID}\n"
//	str := "\ncron: \"*/10 * * * *\"\n\n\ncron_compensator:\n  enable: true\n  latest_first: true\n\n\nversion: \"1.0\"\nstages:\n  - stage:\n      - git-checkout:\n          alias: git-checkout\n          resources:\n            cpu: 0.2\n            mem: 512\n          params:\n            depth: 1\n\n  - stage:\n      - java-build:\n          alias: java-build\n          version: \"1.0\"\n          resources:\n            cpu: 0.2\n            mem: 512\n          params:\n            build_cmd:\n               - \"mvn clean install -Dmaven.test.skip=true\"\n              \n            jdk_version: 8\n            workdir: ${git-checkout}\n\n  - stage:\n      - release:\n          params:\n            dice_yml: ${git-checkout}/dice.yml\n            services:\n              java-demo:\n                # 图形界面基于选择，当然可以自选\n                image: openjdk:8-jre-alpine\n                copys:\n                  - ${java-build:OUTPUT:buildPath}/target/docker-java-app-example.jar:/target\n                  - ${java-build:OUTPUT:buildPath}/spot-agent/spot-agent.jar:/spot-agent\n                cmd: java ${java-build:OUTPUT:JAVA_OPTS} -jar /target/docker-java-app-example.jar\n\n  - stage:\n      - dice:\n          params:\n            release_id: ${release:OUTPUT:releaseID}\n"
//	str = strings.ReplaceAll(str, "\n", "\r\n")
//	ma := make(map[string]string)
//	ma["spec"] = str
//	aaa, _ := json.Marshal(ma)
//	fmt.Println(string(aaa))
//}

// func TestRenderTemplate(t *testing.T) {
//ChangeString()
// specYaml, err := ioutil.ReadFile("/Users/terminus/go/src/terminus.io/dice/extensions/template/custom-script-output/1.0/spec.yml")
// if err != nil {
// 	fmt.Println(err)
// 	t.Fail()
// 	return
// }
// ma := make(map[string]string)
// ma["spec"] = string(specYaml)

// str, err := json.Marshal(ma)
// if err != nil {
// 	fmt.Println(err)
// 	fmt.Println("----------")
// }
// fmt.Println(string(str))

//tests := []struct {
//	params map[string]interface{}
//}{
//	{
//		//params: map[string]interface{}{
//		//	"jdk_version":         "8",
//		//	"build_cmd":           []string{"mvn clean install -Dmaven.test.skip=true"},
//		//	"jar_name":            "docker-java-app-example.jar",
//		//	"pipeline_cron":       "*/10 * * * *",
//		//	"pipeline_scheduling": "new_to_old",
//		//},
//		params: map[string]interface{}{
//			"sleep_time":          1,
//			"sleep_time_one":      2,
//			"jar_name":            "docker-java-app-example.jar",
//			"pipeline_cron":       "*/10 * * * *",
//			"pipeline_scheduling": "new_to_old",
//		},
//	},
//}
//
//for _, v := range tests {
//	template, err := RenderTemplate(string(specYaml), v.params, "git-java-boot-release")
//	if err != nil {
//		fmt.Println(err)
//		t.Fail()
//	}
//
//	fmt.Println(template)
//}

// }
