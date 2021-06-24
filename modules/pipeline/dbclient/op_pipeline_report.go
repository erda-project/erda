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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (client *Client) CreatePipelineReport(report *spec.PipelineReport, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.InsertOne(report)
	return err
}

func (client *Client) UpdatePipelineReport(report *spec.PipelineReport, ops ...SessionOption) error {
	if report.ID == 0 {
		return fmt.Errorf("cannot update report, missing report id")
	}
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(report.ID).Update(report)
	return err
}

func (client *Client) PagingPipelineReportSets(req apistructs.PipelineReportSetPagingRequest, ops ...SessionOption) ([]apistructs.PipelineReportSet, int, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	// 预处理
	if len(req.Types) == 0 {
		req.Types = append(req.Types, apistructs.PipelineReportTypeBasic)
	}

	// 先查询流水线 ID 列表
	// 若 req.PipelineIDs 不为空，则直接使用不做查询
	if len(req.PipelineIDs) == 0 {
		// 报告类型，转换为 mustMatchLabels
		for _, typ := range req.Types {
			reportLabelKey, reportLabelValue := client.MakePipelineReportTypeLabelKey(typ)
			req.MustMatchLabelsQueryParams = append(req.MustMatchLabelsQueryParams, fmt.Sprintf("%s=%s", reportLabelKey, reportLabelValue))
		}
		pipelinePagingReq := &apistructs.PipelinePageListRequest{
			Sources:                    req.Sources,
			AllSources:                 false,
			StartTimeBeginTimestamp:    req.StartTimeBeginTimestamp,
			EndTimeBeginTimestamp:      req.EndTimeBeginTimestamp,
			StartTimeCreatedTimestamp:  req.StartTimeBeginTimestamp,
			EndTimeCreatedTimestamp:    req.EndTimeCreatedTimestamp,
			MustMatchLabelsQueryParams: req.MustMatchLabelsQueryParams,
			PageNum:                    req.PageNum,
			PageSize:                   req.PageSize,
			CountOnly:                  true,
		}
		if err := pipelinePagingReq.PostHandleQueryString(); err != nil {
			return nil, -1, fmt.Errorf("failed to paging pipeline ids, invalid req, err: %v", err)
		}
		_, pipelineIDs, _, _, err := client.PageListPipelines(*pipelinePagingReq, ops...)
		if err != nil {
			return nil, -1, fmt.Errorf("failed to paging pipeline ids, err: %v", err)
		}
		if len(pipelineIDs) == 0 {
			return nil, 0, nil
		}
		req.PipelineIDs = pipelineIDs
	}

	// 查询流水线报告
	sql := session.In("pipeline_id", req.PipelineIDs)
	if len(req.Types) > 0 {
		sql = sql.In("type", req.Types)
	}
	var reports []spec.PipelineReport
	if err := sql.Find(&reports); err != nil {
		return nil, -1, err
	}

	// 转换报告集
	reportSetMap := make(map[uint64]apistructs.PipelineReportSet)
	for _, report := range reports {
		set := reportSetMap[report.PipelineID]
		set.PipelineID = report.PipelineID
		set.Reports = append(set.Reports, client.ConvertPipelineReport(report))
		reportSetMap[report.PipelineID] = set
	}
	// 按照 pipelineID 倒序
	var sets []apistructs.PipelineReportSet
	for _, set := range reportSetMap {
		sets = append(sets, set)
	}
	sort.Slice(sets, func(i, j int) bool {
		// pipelineID 大的排在前面
		return sets[i].PipelineID > sets[j].PipelineID
	})
	total := len(sets)

	return sets, total, nil
}

func (client *Client) DeletePipelineReportsByPipelineID(pipelineID uint64, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(pipelineID).Delete(&spec.PipelineReport{})
	return err
}

func (client *Client) BatchListPipelineReportsByPipelineID(pipelineIDs []uint64, types []string, ops ...SessionOption) (map[uint64][]spec.PipelineReport, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var reports []spec.PipelineReport

	sql := session.In("pipeline_id", pipelineIDs)

	if len(types) > 0 {
		sql.In("type", types)
	}

	sql = sql.Desc("id")

	if err := sql.Find(&reports); err != nil {
		return nil, err
	}

	// handle to map format sort by pipelineID
	results := make(map[uint64][]spec.PipelineReport)
	for _, report := range reports {
		results[report.PipelineID] = append(results[report.PipelineID], report)
	}

	return results, nil
}

func (client *Client) MakePipelineReportTypeLabelKey(typ apistructs.PipelineReportType) (string, string) {
	return fmt.Sprintf("has-report-%s", string(typ)), strconv.FormatBool(true)
}

func (client *Client) ConvertPipelineReport(dbReport spec.PipelineReport) apistructs.PipelineReport {
	return apistructs.PipelineReport{
		ID:         dbReport.ID,
		PipelineID: dbReport.PipelineID,
		Type:       dbReport.Type,
		Meta:       dbReport.Meta,
		CreatorID:  dbReport.CreatorID,
		UpdaterID:  dbReport.UpdaterID,
		CreatedAt:  dbReport.CreatedAt,
		UpdatedAt:  dbReport.UpdatedAt,
	}
}
