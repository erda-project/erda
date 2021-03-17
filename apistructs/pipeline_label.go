package apistructs

type PipelineIDSelectByLabelRequest struct {
	PipelineSources  []PipelineSource `json:"pipelineSource"`
	PipelineYmlNames []string         `json:"pipelineYmlName"`

	// MUST match
	MustMatchLabels map[string][]string `json:"mustMatchLabels"`
	// ANY match
	AnyMatchLabels map[string][]string `json:"anyMatchLabels"`

	// AllowNoPipelineSources, default is false.
	// 默认查询必须带上 pipeline source，增加区分度
	AllowNoPipelineSources bool `json:"allowNoPipelineSources"`

	// OrderByPipelineIDASC 根据 pipeline_id 升序，默认为 false，即降序
	OrderByPipelineIDASC bool `json:"orderByPipelineIDDesc"`
}
