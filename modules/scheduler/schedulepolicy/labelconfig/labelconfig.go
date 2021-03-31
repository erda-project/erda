// Package labelconfig label 调度相关 key 名, 以及中间结构定义
package labelconfig

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const (
	// executor 名字列表
	// EXECUTOR_K8S k8s
	EXECUTOR_K8S = "K8S"
	// EXECUTOR_METRONOME metronome
	EXECUTOR_METRONOME = "METRONOME"
	// EXECUTOR_MARATHON marathon
	EXECUTOR_MARATHON = "MARATHON"
	// EXECUTOR_CHRONOS chronos
	EXECUTOR_CHRONOS = "CHRONOS"
	// EXECUTOR_EDAS edas
	EXECUTOR_EDAS = "EDAS"
	// EXECUTOR_EDAS edas
	EXECUTOR_EDASV2 = "EDASV2"
	// EXECUTOR_SPARK spark
	EXECUTOR_SPARK = "SPARK"
	// EXECUTOR_SPARK k8s spark
	EXECUTOR_K8SSPARK = "K8SSPARK"
	// EXECUTOR_FLINK flink
	EXECUTOR_FLINK = "FLINK"
	// EXECUTOR_K8SJOB k8sjob
	EXECUTOR_K8SJOB = "K8SJOB"

	// ENABLETAG 是否开启标签调度的
	ENABLETAG = "ENABLETAG"

	// DCOS_ATTRIBUTE dice 约定的限制条件的标识
	DCOS_ATTRIBUTE = "dice_tags"

	// ORG_KEY apistructs.ServiceGroup.Labels 中的 key
	// org label example: "DICE_ORG_NAME": "org-xxxx"
	ORG_KEY = "DICE_ORG_NAME"
	// ORG_VALUE_PREFIX org label 前缀
	ORG_VALUE_PREFIX = "org-"
	// ENABLE_ORG 是否打开 org 调度
	ENABLE_ORG = "ENABLE_ORG"

	// WORKSPACE_KEY apistructs.ServiceGroup.Labels 中的 key
	// workspace label example: "DICE_WORKSPACE" : "workspace-xxxx"
	WORKSPACE_KEY = "DICE_WORKSPACE"
	// WORKSPACE_VALUE_PREFIX workspace label 前缀
	WORKSPACE_VALUE_PREFIX = "workspace-"
	// ENABLE_WORKSPACE 是否打开 workspace 调度
	ENABLE_WORKSPACE = "ENABLE_WORKSPACE"

	// CPU_SUBSCRIBE_RATIO 配置中超卖比的 key
	CPU_SUBSCRIBE_RATIO = "CPU_SUBSCRIBE_RATIO"
	// CPU_NUM_QUOTA 配置 quota 值
	CPU_NUM_QUOTA = "CPU_NUM_QUOTA"

	// WORKSPACE_DEV dev 环境
	WORKSPACE_DEV = "dev"
	// WORKSPACE_TEST test 环境
	WORKSPACE_TEST = "test"
	// WORKSPACE_STAGING staging 环境
	WORKSPACE_STAGING = "staging"
	// WORKSPACE_PROD prod 环境
	WORKSPACE_PROD = "prod"

	// 对 Staging 和 Prod 的 job 可以配置目的工作区，比如:
	// "STAGING_JOB_DEST":"dev"
	// "PROD_JOB_DEST":"dev,test"

	// STAGING_JOB_DEST staging 环境的 JOB 可以运行的环境，如果没有配置，则只允许在原本所属环境（staging）
	STAGING_JOB_DEST = "STAGING_JOB_DEST"
	// PROD_JOB_DEST prod 环境的 JOB 可以运行的环境，如果没有配置，则只允许在原本所属环境（prod）
	PROD_JOB_DEST = "PROD_JOB_DEST"

	// SPECIFIC_HOSTS 指定节点调度的特定key
	SPECIFIC_HOSTS = "SPECIFIC_HOSTS"

	// HOST_UNIQUE 服务打散在不同节点
	// value: 是个json结构的字符串
	// e.g. [["service1-name", "service2-name"], ["service3-name"]]
	// value 代表打散的分组，比如上面的例子， service1 和 service2 互相打散，service3 自身为一组
	HOST_UNIQUE = "HOST_UNIQUE"

	// PLATFORM 服务是否属于平台组件
	PLATFORM = "PLATFORM"

	LOCATION_PREFIX = "LOCATION-"

	// K8SLabelPrefix K8S label prefix
	// node 和 pod 的 label 都用这个前缀
	K8SLabelPrefix = "dice/"
)

// LabelPipelineFunc labelpipeline 中的所有 filter 函数的类型
type LabelPipelineFunc func(*RawLabelRuleResult, *RawLabelRuleResult2, *LabelInfo)

// RawLabelRuleResult label 调度模块最终输出结果结构
type RawLabelRuleResult = apistructs.ScheduleInfo

type RawLabelRuleResult2 = apistructs.ScheduleInfo2

// LabelInfo label 调度模块输入结构
type LabelInfo struct {
	// job 或 runtime 的 label 字段
	Label map[string]string
	// label 所在的 job 或 runtime 对应的 executor name
	ExecutorName string
	// label 所在的 job 或 runtime 对应的 executor kind
	ExecutorKind string
	// ExecutorConfig 集群配置
	ExecutorConfig *executortypes.ExecutorWholeConfigs
	// label 所在的 job 或 runtime 对应的 executor optionsPlus
	OptionsPlus *conf.OptPlus
	// label 宿主（runtime 或者 job）的名字
	ObjName string
	// Selectors map[servicename]diceyml.Selectors
	Selectors map[string]diceyml.Selectors
}
