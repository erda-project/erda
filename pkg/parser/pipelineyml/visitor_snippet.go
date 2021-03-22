package pipelineyml

import (
	"github.com/erda-project/erda/apistructs"
)

const (
	Snippet            = "snippet"
	SnippetLogo        = "http://terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/22/410935c6-e399-463a-b87b-0b774240d12e.png"
	SnippetDisplayName = "嵌套流水线"
	SnippetDesc        = "嵌套流水线可以声明嵌套的其他 pipeline.yml"
)

// HandleSnippetConfigLabel polish snippet config label
func HandleSnippetConfigLabel(snippetConfig *SnippetConfig, globalSnippetConfigLabels map[string]string) SnippetConfig {

	if snippetConfig.Labels == nil {
		snippetConfig.Labels = map[string]string{}
	}

	if globalSnippetConfigLabels != nil {
		for k, v := range globalSnippetConfigLabels {
			snippetConfig.Labels[k] = v
		}
	}

	// dice
	scopeID := snippetConfig.Labels[apistructs.LabelDiceSnippetScopeID]
	if scopeID == "" {
		scopeID = "0"
	}

	version := snippetConfig.Labels[apistructs.LabelChooseSnippetVersion]
	if version == "" {
		version = "latest"
	}

	return *snippetConfig
}

// GetParamDefaultValue get default param value by type
func GetParamDefaultValue(paramType string) interface{} {
	switch paramType {
	case apistructs.PipelineParamStringType, apistructs.PipelineParamIntType:
		return ""
	case apistructs.PipelineParamBoolType:
		return "false"
	}
	return ""
}
