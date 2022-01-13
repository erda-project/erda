// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deployment_order

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

func (d *DeploymentOrder) List(conditions *apistructs.DeploymentOrderListConditions, pageInfo *apistructs.PageInfo) (*apistructs.DeploymentOrderListData, error) {
	total, data, err := d.db.ListDeploymentOrder(conditions, pageInfo)
	if err != nil {
		logrus.Errorf("failed to list deployment order, err: %v", err)
		return nil, err
	}

	ret, err := d.convertDeploymentOrderToResponseItem(data)
	if err != nil {
		logrus.Errorf("failed to convert deployment order to response, err: %v", err)
		return nil, err
	}

	return &apistructs.DeploymentOrderListData{
		Total: total,
		List:  ret,
	}, nil
}

func (d *DeploymentOrder) convertDeploymentOrderToResponseItem(orders []dbclient.DeploymentOrder) ([]*apistructs.DeploymentOrderItem, error) {
	ret := make([]*apistructs.DeploymentOrderItem, 0)

	for _, order := range orders {
		appsStatus := make(apistructs.DeploymentOrderStatusMap, 0)
		if order.Status != "" {
			// parse status
			if err := json.Unmarshal([]byte(order.Status), &appsStatus); err != nil {
				return nil, fmt.Errorf("failed to unmarshal applications status, err: %v", err)
			}
		}

		applicationCount := 1

		releaseResp, err := d.bdl.GetRelease(order.ReleaseId)
		if err != nil {
			return nil, fmt.Errorf("failed to get release %s, err: %v", order.ReleaseId, err)
		}

		if releaseResp.IsProjectRelease {
			applicationCount = len(releaseResp.ApplicationReleaseList)
		}

		applicationStatus := strings.Join([]string{
			strconv.Itoa(parseApplicationStatus(appsStatus)),
			strconv.Itoa(applicationCount)}, "/")

		ret = append(ret, &apistructs.DeploymentOrderItem{
			ID:                order.ID,
			Name:              order.Name,
			ReleaseID:         order.ReleaseId,
			ReleaseVersion:    releaseResp.Version,
			Type:              order.Type,
			ApplicationStatus: applicationStatus,
			Workspace:         order.Workspace,
			Status:            parseDeploymentOrderStatus(appsStatus),
			Operator:          string(order.Operator),
			CreatedAt:         order.CreatedAt,
			UpdatedAt:         order.UpdatedAt,
			StartedAt:         parseStartedTime(order.StartedAt),
		})
	}

	return ret, nil
}

func parseApplicationStatus(status apistructs.DeploymentOrderStatusMap) int {
	var okCount int

	for _, s := range status {
		if s.DeploymentStatus == apistructs.DeploymentStatusOK {
			okCount++
		}
	}

	return okCount
}
