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

package instanceinfo

import (
	"fmt"

	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
)

func (c *Client) CreateHPAEventInfo(hpaEvent dbclient.HPAEventInfo) error {
	if err := c.db.Save(&hpaEvent).Error; err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateVPARecommendation(vpaRecommendation dbclient.RuntimeVPAContainerRecommendation) error {
	if err := c.db.Save(&vpaRecommendation).Error; err != nil {
		return err
	}
	return nil
}

func (c *Client) GetLatestVPARecommendation(runtimeId uint64, ruleId, serviceName, containerName string) (dbclient.RuntimeVPAContainerRecommendation, error) {
	vcr := dbclient.RuntimeVPAContainerRecommendation{}

	if err := c.db.Where("runtime_id = ? AND rule_id = ? AND service_name = ? AND container_name = ?",
		runtimeId, ruleId, serviceName, containerName).Order("updated_at desc").Limit(1).Find(&vcr).Error; err != nil {
		return vcr, err
	}
	return vcr, nil
}

func (c *Client) DeletedOldVPARecommendations(interval int) error {
	query := fmt.Sprintf("date_sub(curdate(), interval %d day)  >= date(updated_at)", interval)
	if err := c.db.Where(query).Delete(&dbclient.RuntimeVPAContainerRecommendation{}).Error; err != nil {
		return err
	}
	return nil
}
