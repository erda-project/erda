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

package dto

import (
	"strconv"

	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type DiceArgsDto struct {
	OrgId       string
	ProjectId   string
	Env         string
	DiceApp     string
	DiceService string
	PageNo      int64
	PageSize    int64
	SortField   string
	SortType    string
}

func (impl DiceArgsDto) GenSelectOptions() []orm.SelectOption {
	var result []orm.SelectOption
	needCluster := true
	// if impl.OrgId != "" {
	// 	result = append(result, orm.SelectOption{
	// 		Type:   orm.ExactMatch,
	// 		Column: "dice_org_id",
	// 		Value:  impl.OrgId,
	// 	})
	// } else {
	// 	needCluster = false
	// }
	if impl.ProjectId != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "dice_project_id",
			Value:  impl.ProjectId,
		})
	} else {
		needCluster = false
	}
	if impl.Env != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "dice_env",
			Value:  impl.Env,
		})
	} else {
		needCluster = false
	}
	if needCluster {
		azDb, err := db.NewGatewayAzInfoServiceImpl()
		if err != nil {
			log.Error("create az db failed")
			return nil
		}
		az, err := azDb.GetAz(&orm.GatewayAzInfo{
			Env:       impl.Env,
			OrgId:     impl.OrgId,
			ProjectId: impl.ProjectId,
		})
		if err != nil {
			log.Error("get az failed")
			return nil
		}
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "dice_cluster_name",
			Value:  az,
		})
	}
	if impl.DiceApp != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "dice_app",
			Value:  impl.DiceApp,
		})
	}
	if impl.DiceService != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "dice_service",
			Value:  impl.DiceService,
		})
	}
	if impl.SortField != "" && impl.SortType != "" {
		if impl.SortField == "apiPath" {
			impl.SortField = "api_path"
		} else if impl.SortField == "createAt" {
			impl.SortField = "create_time"
		}
		var option *orm.SelectOption = nil
		switch impl.SortType {
		case ST_UP:
			option = &orm.SelectOption{
				Type:   orm.AscOrder,
				Column: impl.SortField,
			}
		case ST_DOWN:
			option = &orm.SelectOption{
				Type:   orm.DescOrder,
				Column: impl.SortField,
			}
		default:
			log.Errorf("unknown sort type: %s", impl.SortType)
		}
		if option != nil {
			result = append(result, *option)
		}
	} else {
		// 默认按修改时间排序
		result = append(result, orm.SelectOption{
			Type:   orm.DescOrder,
			Column: "update_time",
		})
	}
	return result

}

func NewDiceArgsDto(c *gin.Context) DiceArgsDto {
	page, err := strconv.Atoi(c.DefaultQuery("pageNo", "1"))
	if err != nil {
		log.Warnf("atoi failed page[%s]", c.Query("pageNo"))
		page = 1
	}
	size, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil {
		log.Warnf("atoi failed size[%s]", c.Query("pageSize"))
		size = 20
	}
	dto := DiceArgsDto{
		OrgId:       c.Query("orgId"),
		ProjectId:   c.Query("projectId"),
		Env:         c.Query("env"),
		DiceApp:     c.Query("diceApp"),
		DiceService: c.Query("diceService"),
		PageNo:      int64(page),
		PageSize:    int64(size),
		SortField:   c.Query("sortField"),
		SortType:    c.Query("sortType"),
	}
	if dto.OrgId == "" {
		dto.OrgId = c.GetHeader("Org-ID")
	}
	if dto.ProjectId == "" {
		dto.ProjectId = c.Query("projectID")
	}
	if dto.Env == "" {
		dto.Env = c.Query("workspace")
	}
	return dto
}
