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

const (
	SnippetActionSourceType = "action"

	ActionTypeAPITest      = "api-test"
	ActionTypeSnippet      = "snippet"
	ActionTypeCustomScript = "custom-script"

	SnippetSourceLocal = "local"
)

const (
	PipelineParamStringType = "string"
	PipelineParamIntType    = "int"
	PipelineParamBoolType   = "boolean"
)

type PipelineYml struct {
	// 用于构造 pipeline yml
	Version         string                 `json:"version"`                   // 版本
	Envs            map[string]string      `json:"envs,omitempty"`            // 环境变量
	Cron            string                 `json:"cron,omitempty"`            // 定时配置
	CronCompensator *CronCompensator       `json:"cronCompensator,omitempty"` // 定时补偿配置
	Stages          [][]*PipelineYmlAction `json:"stages"`                    // 流水线
	FlatActions     []*PipelineYmlAction   `json:"flatActions"`               // 展平了的流水线

	Params []*PipelineParam `json:"params,omitempty"` // 流水线输入

	Outputs []*PipelineOutput `json:"outputs,omitempty"` // 流水线输出

	// --- 以下字段与构造 pipeline yml 无关 ---

	// 1.0 升级相关
	NeedUpgrade bool `json:"needUpgrade"` // pipeline yml 是否需要升级

	// YmlContent:
	// 1) 当 needUpgrade 为 true  时，ymlContent 返回升级后的 yml
	// 2) 当 needUpgrade 为 false 时：
	//    1) 用户传入的为 YAML(apistructs.PipelineYml) 时，ymlContent 返回 YAML(spec.PipelineYml)
	//    2) 用户传入的为 YAML(spec.PipelineYml) 时，返回优化后的 YAML(spec.PipelineYml)
	YmlContent string         `json:"ymlContent,omitempty"`
	On         *TriggerConfig `json:"on,omitempty"`
}

type TriggerConfig struct {
	Push  *PushTrigger  `yaml:"push,omitempty" json:"push,omitempty"`
	Merge *MergeTrigger `yaml:"merge,omitempty" json:"merge,omitempty"`
}

type PushTrigger struct {
	Branches []string `yaml:"branches,omitempty" json:"branches,omitempty"`
	Tags     []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

type MergeTrigger struct {
	Branches []string `yaml:"branches,omitempty" json:"branches,omitempty"`
}

type PipelineYmlAction struct {
	Alias         string                 `json:"alias,omitempty"`                                          // action 实例名
	Type          string                 `json:"type"`                                                     // action 类型，比如：git-checkout, release
	Description   string                 `json:"description,omitempty"`                                    // 描述
	Version       string                 `json:"version,omitempty"`                                        // action 版本
	Params        map[string]interface{} `json:"params,omitempty"`                                         // 参数
	Image         string                 `json:"image,omitempty"`                                          // 镜像
	Commands      []string               `json:"commands,omitempty"`                                       // 命令行
	Timeout       int64                  `json:"timeout,omitempty"`                                        // 超时设置，单位：秒
	Namespaces    []string               `json:"namespaces,omitempty"`                                     // Action 输出的命名空间
	Resources     Resources              `json:"resources,omitempty"`                                      // 资源
	DisplayName   string                 `json:"displayName,omitempty"`                                    // 中文名称
	LogoUrl       string                 `json:"logoUrl,omitempty"`                                        // logo
	Caches        []ActionCache          `json:"caches,omitempty"`                                         // 缓存
	SnippetConfig *SnippetConfig         `json:"snippet_config,omitempty" yaml:"snippet_config,omitempty"` // snippet 的配置
	If            string                 `json:"if,omitempty"`                                             // 条件执行
	Loop          *PipelineTaskLoop      `json:"loop,omitempty"`                                           // 循环执行
	SnippetStages *SnippetStages         `json:"snippetStages,omitempty"`                                  // snippetStages snippet 展开
}

type SnippetStages struct {
	Params  []*PipelineParam       `json:"params,omitempty"`  // 流水线输入
	Outputs []*PipelineOutput      `json:"outputs,omitempty"` // 流水线输出
	Stages  [][]*PipelineYmlAction `json:"stages,omitempty"`  // snippetStages snippet 展开
}

type SnippetConfig struct {
	Source string            `json:"source,omitempty" yaml:"source,omitempty"` // 来源 gittar dice test
	Name   string            `json:"name,omitempty" yaml:"name,omitempty"`     // 名称
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"` // 额外标签
}

type BatchSnippetConfigYml struct {
	Config SnippetConfig `json:"config"`
	Yml    string        `json:"yml"`
}

type PipelineParam struct {
	Name     string      `json:"name" yaml:"name,omitempty"`         // 名称
	Required bool        `json:"required" yaml:"required,omitempty"` // 是否必须
	Default  interface{} `json:"default" yaml:"default,omitempty"`   // 默认值
	Desc     string      `json:"desc" yaml:"desc,omitempty"`         // 描述
	Type     string      `json:"type" yaml:"type,omitempty"`         // 类型
}

type PipelineOutput struct {
	Name string `json:"name" yaml:"name,omitempty"` // 名称
	Desc string `json:"desc" yaml:"desc,omitempty"` // 描述
	Ref  string `json:"ref" yaml:"ref,omitempty"`   // 引用那个 action 的值
}

type PipelineOutputWithValue struct {
	PipelineOutput
	Value interface{} `json:"value,omitempty"` // 具体的值
}

type ActionCache struct {
	// 缓存生成的 key 或者是用户指定的 key
	// 用户指定的话 需要 {{basePath}}/路径/{{endPath}} 来自定义 key
	// 用户没有指定 key 有一定的生成规则, 具体生成规则看 prepare.go 的 setActionCacheStorageAndBinds 方法
	Key  string `yaml:"key,omitempty"`
	Path string `yaml:"path,omitempty"` // 指定那个目录被缓存, 只能是由 / 开始的绝对路径
}

type CronCompensator struct {
	Enable               bool `json:"enable"`
	LatestFirst          bool `json:"latestFirst"`
	StopIfLatterExecuted bool `json:"stopIfLatterExecuted"`
}

type PipelineYmlParseGraphRequest struct {
	PipelineYmlContent        string            `json:"pipelineYmlContent"`
	GlobalSnippetConfigLabels map[string]string `json:"globalSnippetConfigLabels"`
	SnippetConfig             *SnippetConfig    `json:"snippetConfig"`
}

type PipelineYmlParseGraphResponse struct {
	Header
	Data *PipelineYml `json:"data"`
}
