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
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
)

func (c *Client) CreateHPAEventInfo(hpaEvent dbclient.HPAEventInfo) error {
	if err := c.db.Save(&hpaEvent).Error; err != nil {
		return err
	}
	return nil
}
