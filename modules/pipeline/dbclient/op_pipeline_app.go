package dbclient

import (
	"sort"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// ListAppInvokedCombos list combos for pipeline sidebar.
// 所有数据从 labels 中可获取
func (client *Client) ListAppInvokedCombos(appID uint64, selected spec.PipelineCombosReq) (result []apistructs.PipelineInvokedCombo, err error) {

	sources := make([]apistructs.PipelineSource, 0, len(selected.Sources))
	for _, source := range selected.Sources {
		sources = append(sources, apistructs.PipelineSource(source))
	}

	mustMatchLabels := map[string][]string{
		apistructs.LabelAppID: {strconv.FormatUint(appID, 10)},
	}
	if len(selected.Branches) > 0 {
		mustMatchLabels[apistructs.LabelBranch] = selected.Branches
	}

	pipelineIDs, err := client.SelectPipelineIDsByLabels(
		apistructs.PipelineIDSelectByLabelRequest{
			PipelineSources:        sources,
			PipelineYmlNames:       selected.YmlNames,
			MustMatchLabels:        mustMatchLabels,
			AllowNoPipelineSources: true,
		},
	)
	if err != nil {
		return nil, err
	}

	// list pipelines by ids
	pipelines, err := client.ListPipelinesByIDs(pipelineIDs)
	if err != nil {
		return nil, err
	}

	// 将 pipelineYmlName 有关联的 combo 进行合并
	// 特殊处理 pipelineYmlName
	// pipeline.yml -> 1/PROD/master/pipeline.yml
	m := make(map[string]spec.Pipeline)
	for i := range pipelines {
		p := pipelines[i]
		generateV1UniqueYmlName := p.GenerateV1UniquePipelineYmlName(p.PipelineYmlName)
		exist, ok := m[generateV1UniqueYmlName]
		// 取流水线 ID 最大的
		if !ok || p.ID > exist.ID {
			m[p.GenerateV1UniquePipelineYmlName(p.PipelineYmlName)] = p
		}
	}
	for ymlName, p := range m {
		ymlNameMap := map[string]struct{}{
			ymlName:                                  {},
			p.PipelineYmlName:                        {},
			p.Extra.PipelineYmlNameV1:                {},
			p.DecodeV1UniquePipelineYmlName(ymlName): {},
		}
		// 保存需要聚合在一起的 ymlNames
		ymlNames := make([]string, 0)
		// 保存最短的 ymlName 用于 UI 展示
		shortYmlName := p.PipelineYmlName
		for name := range ymlNameMap {
			if name == "" {
				continue
			}
			if len(name) < len(shortYmlName) {
				shortYmlName = name
			}
			ymlNames = append(ymlNames, name)
		}
		result = append(result, apistructs.PipelineInvokedCombo{
			Branch: p.Labels[apistructs.LabelBranch], Source: string(p.PipelineSource), YmlName: shortYmlName, PagingYmlNames: ymlNames,
			PipelineID: p.ID, Commit: p.GetCommitID(), Status: string(p.Status),
			TimeCreated: p.TimeCreated, CancelUser: p.Extra.CancelUser,
			TriggerMode: string(p.TriggerMode),
		})
	}
	// 排序 ID DESC
	sort.Slice(result, func(i, j int) bool {
		return result[i].PipelineID > result[j].PipelineID
	})

	return result, nil
}
