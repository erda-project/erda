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

package extension

import (
	"testing"

	"github.com/alecthomas/assert"
)

func Test_ConvertConfigExtensionMenu(t *testing.T) {
	cfg := &config{
		ExtensionMenu: "{}",
	}
	test, err := cfg.ConvertConfigExtensionMenu()
	want := map[string][]string{}
	assert.NoError(t, err)
	assert.Equal(t, test, want)

	cfg.ExtensionMenu = "{\"流水线任务\":[\"source_code_management:代码管理\",\"build_management:构建管理\",\"deploy_management:部署管理\",\"version_management:版本管理\",\"test_management:测试管理\",\"data_management:数据治理\",\"custom_task:自定义任务\"],\"扩展服务\":[\"database:存储\",\"distributed_cooperation:分布式协作\",\"search:搜索\",\"message:消息\",\"content_management:内容管理\",\"security:安全\",\"traffic_load:流量负载\",\"monitoring&logging:监控&日志\",\"content:文本处理\",\"image_processing:图像处理\",\"document_processing:文件处理\",\"sound_processing:音频处理\",\"custom:自定义\",\"general_ability:通用能力\",\"new_retail:新零售能力\",\"srm:采供能力\",\"solution:解决方案\"]}"
	test, err = cfg.ConvertConfigExtensionMenu()
	want = map[string][]string{
		"流水线任务": {
			"source_code_management:代码管理",
			"build_management:构建管理",
			"deploy_management:部署管理",
			"version_management:版本管理",
			"test_management:测试管理",
			"data_management:数据治理",
			"custom_task:自定义任务",
		},
		"扩展服务": {
			"database:存储",
			"distributed_cooperation:分布式协作",
			"search:搜索",
			"message:消息",
			"content_management:内容管理",
			"security:安全",
			"traffic_load:流量负载",
			"monitoring&logging:监控&日志",
			"content:文本处理",
			"image_processing:图像处理",
			"document_processing:文件处理",
			"sound_processing:音频处理",
			"custom:自定义",
			"general_ability:通用能力",
			"new_retail:新零售能力",
			"srm:采供能力",
			"solution:解决方案",
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, test, want)
}
