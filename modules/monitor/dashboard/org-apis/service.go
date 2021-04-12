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

	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

var (
	tsqlClusterStatus     = "SELECT health_status::field AS status, component_name::tag AS component_name FROM component_status WHERE cluster_name::tag = '%s' AND component_group::tag = 'cluster' ORDER BY timestamp DESC LIMIT 1"
	tsqlComponentStatuses = "SELECT health_status::field AS status, component_name::tag AS component_name FROM component_status WHERE cluster_name::tag = '%s' GROUP BY component_name::tag"
)

type queryServiceImpl interface {
	queryComponentStatus(componentType, clusterName string) (statuses []*statusDTO, err error)
}

type queryService struct {
	metricQ metricq.Queryer
}

func (mqs *queryService) queryComponentStatus(componentType, clusterName string) (statuses []*statusDTO, err error) {
	switch componentType {
	case "cluster":
		res, err := mqs.queryStatusWithTSQL(fmt.Sprintf(tsqlClusterStatus, clusterName))
		if err != nil {
			return nil, err
		}
		return res, nil
	case "component":
		res, err := mqs.queryStatusWithTSQL(fmt.Sprintf(tsqlComponentStatuses, clusterName))
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return
}

func (mqs *queryService) queryStatusWithTSQL(statement string) (statuses []*statusDTO, err error) {
	_, data, err := mqs.metricQ.QueryWithFormat(metricq.InfluxQL, statement, "dict", nil, nil, defaultDuration())
	if err != nil {
		return nil, errors.Wrap(err, "query inlfuxql failed")
	}
	if err := mapstructure.Decode(data, &statuses); err != nil {
		return nil, errors.Wrap(err, "unmarshal failed")
	}
	return statuses, nil
}

func defaultDuration() url.Values {
	options := url.Values{}
	options.Set("start", "before_1h")
	options.Set("end", "now")
	return options
}
