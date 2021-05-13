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

package orgapis

import (
	"fmt"
	"net/url"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
)

var (
	tsqlComponentStatus = `SELECT health_status::field AS health_status, weight::field AS weight, component_name::tag AS component_name FROM leaf_component_status WHERE cluster_name::tag = '%s' AND component_group::tag = '%s'`
)

var (
	componentNames = []string{"dice_addon", "dice_component", "kubernetes", "machine"}
)

type queryServiceImpl interface {
	queryStatus(clusterName string) (statuses []*statusDTO, err error)
}

type queryService struct {
	metricQ metricq.Queryer
}

func (mqs *queryService) queryStatus(clusterName string) (statuses []*statusDTO, err error) {
	for _, name := range componentNames {
		raws, err := mqs.queryStatusWithTSQL(fmt.Sprintf(tsqlComponentStatus, clusterName, name))
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, &statusDTO{
			Name:   name,
			Status: calculateStatus(raws, name),
		})
	}
	return
}

type intStatus uint8

func (is *intStatus) update(other uint8) {
	if intStatus(other) > *is {
		*is = *(*intStatus)(&other)
	}
}

func calculateStatus(raws []rawStatus, name string) uint8 {
	vipCnt, failureCnt := 0, 0
	// health, noraml
	status := intStatus(0)
	for _, item := range raws {
		if item.Weight == 1 {
			vipCnt++
		}

		if item.HealthStatus == 1 {
			status.update(1)
		}
		if item.HealthStatus == 2 {
			if item.Weight == 0 {
				status.update(1)
			}
			if item.Weight == 1 {
				switch name {
				case "machine":
					status.update(3)
				default:
					status.update(2)
				}
				failureCnt++
			}
		}
	}
	if vipCnt > 0 && failureCnt == vipCnt { // all critical component failure
		status.update(3)
	}
	return uint8(status)
}

type rawStatus struct {
	HealthStatus  uint8  `mapstructure:"health_status"`
	Weight        uint8  `mapstructure:"weight"`
	ComponentName string `mapstructure:"component_name"`
}

func (mqs *queryService) queryStatusWithTSQL(statement string) ([]rawStatus, error) {
	raws := []rawStatus{}
	_, data, err := mqs.metricQ.QueryWithFormat(metricq.InfluxQL, statement, "dict", nil, nil, defaultDuration())
	if err != nil {
		return nil, errors.Wrap(err, "query inlfuxql failed")
	}
	if err := mapstructure.Decode(data, &raws); err != nil {
		return nil, errors.Wrap(err, "unmarshal failed")
	}
	return raws, nil
}

func defaultDuration() url.Values {
	options := url.Values{}
	options.Set("start", "before_15m")
	options.Set("end", "now")
	return options
}
