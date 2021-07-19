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

package apistructs

import "strings"

// PipelineSource 从整体上区分流水线来源
//
// 例如 dice(普通 CI/CD)、qa(质量分析)、bigdata(大数据)、ops(运维链路)、api-test(API Test)
type PipelineSource string

func (s PipelineSource) String() string {
	return string(s)
}

var (
	PipelineSourceDefault      PipelineSource = "default"
	PipelineSourceDice         PipelineSource = "dice"             // 普通 Dice CI/CD
	PipelineSourceBigData      PipelineSource = "bigdata"          // 大数据任务
	PipelineSourceOps          PipelineSource = "ops"              // ops 链路
	PipelineSourceQA           PipelineSource = "qa"               // qa 链路
	PipelineSourceConfigSheet  PipelineSource = "config-sheet"     // 配置单
	PipelineSourceProject      PipelineSource = "project-pipeline" // 项目级流水线
	PipelineSourceProjectLocal PipelineSource = "local"            // gittar 流水线
	PipelineSourceAPITest      PipelineSource = "api-test"         // API Test
	PipelineSourceAutoTest     PipelineSource = "autotest"
	PipelineSourceAutoTestPlan PipelineSource = "autotest-plan"

	// cdp workflow
	PipelineSourceCDPDev     PipelineSource = "cdp-dev"
	PipelineSourceCDPTest    PipelineSource = "cdp-test"
	PipelineSourceCDPStaging PipelineSource = "cdp-staging"
	PipelineSourceCDPProd    PipelineSource = "cdp-prod"

	// cdp recommend
	PipelineSourceRecommendDev     PipelineSource = "recommend-dev"
	PipelineSourceRecommendTest    PipelineSource = "recommend-test"
	PipelineSourceRecommendStaging PipelineSource = "recommend-staging"
	PipelineSourceRecommendProd    PipelineSource = "recommend-prod"
)

// Valid 返回 PipelineSource 是否有效
func (s PipelineSource) Valid() bool {
	if s == PipelineSourceDefault ||
		s == PipelineSourceDice ||
		s == PipelineSourceOps {
		return true
	}
	if s.IsTest() {
		return true
	}
	if s.IsBigData() {
		return true
	}
	if s.IsConfigSheet() {
		return true
	}
	if s.IsProjectPipeline() {
		return true
	}
	return false
}

func (s PipelineSource) IsBigData() bool {
	if s == PipelineSourceBigData {
		return true
	}
	if strings.HasPrefix(s.String(), "cdp-") {
		return true
	}
	if strings.HasPrefix(s.String(), "recommend-") {
		return true
	}
	return false
}

func (s PipelineSource) IsTest() bool {
	switch s {
	case PipelineSourceQA, PipelineSourceAPITest, PipelineSourceAutoTest:
		return true
	default:
		return false
	}
}

func (s PipelineSource) IsConfigSheet() bool {
	switch s {
	case PipelineSourceConfigSheet:
		return true
	default:
		return false
	}
}

func (s PipelineSource) IsProjectPipeline() bool {
	switch s {
	case PipelineSourceProject:
		return true
	case PipelineSourceProjectLocal:
		return true
	default:
		return false
	}
}

// PipelineYmlSource 表示 pipelineYml 文件来源
type PipelineYmlSource string

var (
	PipelineYmlSourceContent PipelineYmlSource = "content" // pipeline.yml 直接由 api 调用时作为参数传入
	PipelineYmlSourceGittar  PipelineYmlSource = "gittar"  // pipeline.yml 从 gittar 中获取
)

// Valid 返回 PipelineYmlSource 是否有效
func (s PipelineYmlSource) Valid() bool {
	if s == PipelineYmlSourceContent || s == PipelineYmlSourceGittar {
		return true
	}
	return false
}

func (s PipelineYmlSource) String() string {
	return string(s)
}

// PipelineTriggerMode 流水线触发方式，手动 or 定时
type PipelineTriggerMode string

var (
	PipelineTriggerModeManual PipelineTriggerMode = "manual" // 手动触发
	PipelineTriggerModeCron   PipelineTriggerMode = "cron"   // 定时触发
)

// Valid 返回 PipelineTriggerMode 是否有效
func (m PipelineTriggerMode) Valid() bool {
	if m == PipelineTriggerModeManual || m == PipelineTriggerModeCron {
		return true
	}
	return false
}

func (m PipelineTriggerMode) String() string {
	return string(m)
}

// PipelineType 流水线运行类型，普通 or 重试失败节点
type PipelineType string

var (
	PipelineTypeNormal      PipelineType = "normal"       // normal
	PipelineTypeRerunFailed PipelineType = "rerun-failed" // retry failed action
	PipelineTypeRerun       PipelineType = "rerun"        // retry the whole pipeline
)

// Valid 返回 PipelineType 是否有效
func (t PipelineType) Valid() bool {
	if t == PipelineTypeNormal || t == PipelineTypeRerunFailed || t == PipelineTypeRerun {
		return true
	}
	return false
}

func (t PipelineType) String() string {
	return string(t)
}
