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

package dbclient

import (
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/retry"
	"github.com/erda-project/erda/pkg/strutil"
)

func (client *Client) GetLabel(id uint64) (label *spec.PipelineLabel, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "failed to get pipeline label by id: %v", id)
		}
	}()
	found, err := client.ID(id).Get(label)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("not found")
	}
	return label, nil
}

func (client *Client) CreatePipelineLabels(p *spec.Pipeline, ops ...SessionOption) (err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	defer func() { err = errors.Wrap(err, "failed to create pipeline label") }()
	labels := make([]spec.PipelineLabel, 0, len(p.Labels))
	for k, v := range p.Labels {
		label := spec.PipelineLabel{
			PipelineSource:  p.PipelineSource,
			PipelineYmlName: p.PipelineYmlName,
			PipelineID:      p.ID,
			Key:             k,
			Value:           v,
		}
		labels = append(labels, label)
	}
	_, err = session.InsertMulti(labels)
	return err
}

// ListLabelsByPipelineID 根据 pipelineID 获取 labels
func (client *Client) ListLabelsByPipelineID(pipelineID uint64, ops ...SessionOption) ([]spec.PipelineLabel, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var labels []spec.PipelineLabel
	if err := session.Find(&labels, spec.PipelineLabel{PipelineID: pipelineID}); err != nil {
		return nil, err
	}
	return labels, nil
}

func (client *Client) SelectPipelineIDsByLabels(req apistructs.PipelineIDSelectByLabelRequest, ops ...SessionOption) (pipelineIDs []uint64, err error) {
	defer func() {
		if err != nil {
			err = errors.Errorf("failed to get pipelineIDs match by labels, req: %+v, err: %v", req, err)
		}
	}()

	session := client.NewSession(ops...)
	defer session.Close()

	// 校验
	if len(req.PipelineSources) == 0 {
		req.PipelineSources = []apistructs.PipelineSource{apistructs.PipelineSourceDefault}
	}
	if req.AllowNoPipelineSources {
		req.PipelineSources = []apistructs.PipelineSource{}
	}
	if len(req.MustMatchLabels) > 0 && len(req.AnyMatchLabels) > 0 {
		return nil, errors.Errorf("please only set one of mustMatchLabels and anyMatchLabels")
	}
	if len(req.MustMatchLabels) == 0 && len(req.AnyMatchLabels) == 0 {
		return nil, errors.Errorf("neither mustMathLabels nor anyMathLabels set")
	}

	// SQL
	sqlSegments := []string{
		"SELECT `pipeline_id` FROM `pipeline_labels`",
		"FORCE INDEX (`idx_source_key_value_pipelineid`, `idx_source_ymlname_key_value_pipelineid`)",
		"WHERE 1=1",
	}
	sqlArgs := []interface{}{}

	// segment: pipeline_source
	if len(req.PipelineSources) > 0 {
		sqlSegments = append(sqlSegments, fmt.Sprintf("AND `pipeline_source` IN (%s)", questionMarks(len(req.PipelineSources))))
		for _, source := range req.PipelineSources {
			sqlArgs = append(sqlArgs, source.String())
		}
	}

	// segment: pipeline_yml_name
	if len(req.PipelineYmlNames) > 0 {
		sqlSegments = append(sqlSegments, fmt.Sprintf("AND `pipeline_yml_name` IN (%s)", questionMarks(len(req.PipelineYmlNames))))
		for _, name := range req.PipelineYmlNames {
			sqlArgs = append(sqlArgs, name)
		}
	}

	// 标签处理
	// 每组 kye/values 单独处理，然后在内存中进行组合
	var handleResult []uint64

	if len(req.MustMatchLabels) > 0 {
		var allNeedFilterPipelineIDs [][]uint64
		for key, values := range req.MustMatchLabels {
			if len(values) == 0 {
				continue
			}
			innerSegments := append(sqlSegments, fmt.Sprintf("AND `key` = ? AND `value` IN (%s)", questionMarks(len(values))))
			innerArgs := append(sqlArgs, key)
			for _, v := range values {
				innerArgs = append(innerArgs, v)
			}
			// execute
			var innerPipelineIDs []uint64
			err := session.Prepare().SQL(strutil.Join(innerSegments, " "), innerArgs...).Find(&innerPipelineIDs)
			if err != nil {
				return nil, err
			}
			allNeedFilterPipelineIDs = append(allNeedFilterPipelineIDs, innerPipelineIDs)
		}
		handleResult = filter(allNeedFilterPipelineIDs...)
	}

	if len(req.AnyMatchLabels) > 0 {
		var allNeedMergePipelineIDs [][]uint64
		for key, values := range req.AnyMatchLabels {
			if len(values) == 0 {
				continue
			}
			innerSegments := append(sqlSegments, fmt.Sprintf("AND `key` = ? AND `value` IN (%s)", questionMarks(len(values))))
			innerArgs := append(sqlArgs, key)
			for _, v := range values {
				innerArgs = append(innerArgs, v)
			}
			// execute
			var innerPipelineIDs []uint64
			err := session.SQL(strutil.Join(innerSegments, " "), innerArgs...).Find(&innerPipelineIDs)
			if err != nil {
				return nil, err
			}
			allNeedMergePipelineIDs = append(allNeedMergePipelineIDs, innerPipelineIDs)
		}
		handleResult = merge(allNeedMergePipelineIDs...)
	}

	pipelineIDs = handleResult

	// ORDER BY `pipeline_id` DESC / ASC
	if req.OrderByPipelineIDASC {
		// ASC
		sort.Slice(pipelineIDs, func(i, j int) bool { return pipelineIDs[i] < pipelineIDs[j] })
	} else {
		// DESC
		sort.Slice(pipelineIDs, func(i, j int) bool { return pipelineIDs[i] > pipelineIDs[j] })
	}

	return pipelineIDs, nil
}

func (client *Client) DeletePipelineLabelsByPipelineID(pipelineID uint64, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	return retry.DoWithInterval(func() error {
		_, err := session.Delete(&spec.PipelineLabel{PipelineID: pipelineID})
		return err
	}, 3, time.Second)
}

// merge 合并 inputs 里的 id
func merge(inputs ...[]uint64) []uint64 {
	m := make(map[uint64]struct{})
	for _, input := range inputs {
		for _, id := range input {
			m[id] = struct{}{}
		}
	}
	var result []uint64
	for id := range m {
		result = append(result, id)
	}
	return result
}

// filter 保证 result 中的 key 在每组 input 中都存在
func filter(inputs ...[]uint64) []uint64 {
	if len(inputs) == 0 {
		return nil
	}
	// idCountMap: key: id, value: id 出现的次数
	idCountMap := make(map[uint64]int)
	for _, input := range inputs {
		for _, id := range input {
			idCountMap[id]++ // id 出现次数 +1
		}
	}
	for id, count := range idCountMap {
		// 若 id 在每组 input 中都存在，则出现的次数 count = input 组数
		if count < len(inputs) {
			delete(idCountMap, id)
		}
	}
	// idCountMap map -> slice
	var result []uint64
	for id := range idCountMap {
		result = append(result, id)
	}
	return result
}

// filterAndOrder filter 后，结果以 inputs[0] 的顺序为准
func filterAndOrder(inputs ...[]uint64) []uint64 {
	if len(inputs) == 0 {
		return nil
	}
	filterResult := filter(inputs...)

	// 以 inputs[0] 的顺序为准
	var orderedResult []uint64
	filterResultMap := make(map[uint64]struct{}, len(filterResult))
	for _, id := range filterResult {
		filterResultMap[id] = struct{}{}
	}
	for _, id := range inputs[0] {
		if _, ok := filterResultMap[id]; ok {
			orderedResult = append(orderedResult, id)
		}
	}
	return orderedResult
}

func paging(ids []uint64, pageNum, pageSize int) []uint64 {
	offset := (pageNum - 1) * pageSize
	limit := pageSize

	// 左边界
	if offset > len(ids) {
		return nil
	}
	// 右边界
	if (offset + limit) > len(ids) {
		return ids[offset:]
	}

	return ids[offset:(offset + limit)]
}

func questionMarks(length int) string {
	result := make([]string, length)
	for i := range result {
		result[i] = "?"
	}
	return strutil.Join(result, ",")
}

func (client *Client) ListPipelineLabelsByPipelineIDs(pipelineIDs []uint64, ops ...SessionOption) (map[uint64][]spec.PipelineLabel, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var labels []spec.PipelineLabel
	if err := session.In("pipeline_id", pipelineIDs).Find(&labels); err != nil {
		return nil, err
	}

	pipelineLabelsMap := make(map[uint64][]spec.PipelineLabel)
	for _, label := range labels {
		_, ok := pipelineLabelsMap[label.PipelineID]
		if !ok {
			pipelineLabelsMap[label.PipelineID] = make([]spec.PipelineLabel, 0)
		}
		pipelineLabelsMap[label.PipelineID] = append(pipelineLabelsMap[label.PipelineID], label)
	}

	return pipelineLabelsMap, nil
}
