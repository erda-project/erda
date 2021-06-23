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

package errorbox

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
)

// ErrorBox 错误日志操作封装
type ErrorBox struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 Member 对象配置选项
type Option func(*ErrorBox)

// New 新建 ErrorBox 实例
func New(options ...Option) *ErrorBox {
	eb := &ErrorBox{}
	for _, op := range options {
		op(eb)
	}
	return eb
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(eb *ErrorBox) {
		eb.db = db
	}
}

// WithBundle 配置 bdl
func WithBundle(bdl *bundle.Bundle) Option {
	return func(eb *ErrorBox) {
		eb.bdl = bdl
	}
}

// CreateOrUpdate 记录或更新错误日志
func (eb *ErrorBox) CreateOrUpdate(req apistructs.ErrorLogCreateRequest) error {
	occurrenceTime, err := req.FormartTime()
	if err != nil {
		return err
	}

	// 更新
	if req.DedupID != "" {
		errorLog, err := eb.db.GetErrorLogByDedupIndex(req.ResourceID, req.ResourceType, req.DedupID)
		if err != nil {
			return err
		}

		if errorLog != nil {
			errorLog.Level = req.Level
			errorLog.OccurrenceTime = *occurrenceTime
			errorLog.HumanLog = req.HumanLog
			errorLog.PrimevalLog = req.PrimevalLog
			if err := eb.db.UpdateErrorLog(errorLog); err != nil {
				return err
			}
			return nil
		}
	}

	// 创建
	if err := eb.db.CreateErrorLog(&model.ErrorLog{
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		Level:          req.Level,
		OccurrenceTime: *occurrenceTime,
		HumanLog:       req.HumanLog,
		PrimevalLog:    req.PrimevalLog,
		DedupID:        req.DedupID,
	}); err != nil {
		return err
	}

	return nil
}

// BatchCreateErrorLogs 批量创建错误日志
func (eb *ErrorBox) BatchCreateErrorLogs(reqs []apistructs.Audit) error {
	return nil
}

// List 通过参数过滤错误日志
func (eb *ErrorBox) List(param *apistructs.ErrorLogListRequest) ([]model.ErrorLog, error) {
	resourceIDs, resourceTypes, err := eb.aggregateResources(param.ResourceType, param.ResourceID)
	if err != nil {
		return nil, err
	}

	if param.StartTime != "" {
		startTime, err := param.GetFormartStartTime()
		if err != nil {
			return nil, err
		}
		return eb.db.ListErrorLogByResourcesAndStartTime(resourceTypes, resourceIDs, *startTime)
	}

	return eb.db.ListErrorLogByResources(resourceTypes, resourceIDs)
}

// aggregateResources 聚合目标资源下的所有子资源
func (eb *ErrorBox) aggregateResources(resourceType apistructs.ErrorResourceType, resourceID string) ([]string,
	[]apistructs.ErrorResourceType, error) {
	resourceTypes, resourceIDs := []apistructs.ErrorResourceType{resourceType}, []string{resourceID}

	switch resourceType {
	case apistructs.PipelineError:
		pipelineID, err := strconv.ParseUint(resourceID, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		runtimeIDs, err := eb.FindRuntimeByPipelineID(pipelineID)
		if err != nil {
			return nil, nil, err
		}
		if len(runtimeIDs) != 0 {
			resourceIDs = append(resourceIDs, runtimeIDs...)
			resourceTypes = append(resourceTypes, apistructs.RuntimeError)
		}

		for _, v := range runtimeIDs {
			runtimeID, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			addonIDs, err := eb.FindAddonByRuntimeID(runtimeID)
			if err != nil {
				return nil, nil, err
			}
			if len(addonIDs) != 0 {
				resourceIDs = append(resourceIDs, addonIDs...)
				resourceTypes = append(resourceTypes, apistructs.AddonError)
			}
		}
	case apistructs.RuntimeError:
		runtimeID, err := strconv.ParseUint(resourceID, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		addonIDs, err := eb.FindAddonByRuntimeID(runtimeID)
		if err != nil {
			return nil, nil, err
		}
		if len(addonIDs) != 0 {
			resourceIDs = append(resourceIDs, addonIDs...)
			resourceTypes = append(resourceTypes, apistructs.AddonError)
		}
		// case apistructs.AddonError:
	}

	return resourceIDs, resourceTypes, nil
}
