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

package types

// 测试结果解析 parser
type TestParserType string

const (
	// 使用 maven-surefire-plugin 生成的 TEST-xxx.xml 格式进行解析
	Default TestParserType = "DEFAULT"
	// 使用 testng 插件 org.testng.reporters.XMLReporter 生成的格式进行解析, 需要在 pom.xml 中配置该 reporter
	NGTest TestParserType = "NGTEST"
	// 使用 junit 生成的 xml 格式进行解析
	JUnit TestParserType = "JUNIT"
)

func (t TestParserType) TPValue() string {
	return string(t)
}
