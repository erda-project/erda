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

package aliyun_resources

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/dbclient"
)

func ProcessFailedRecord(ctx Context, source string, addonID string, record *dbclient.Record,
	detail *apistructs.CreateCloudResourceRecord, err error) {
	if len(detail.Steps) == 0 {
		return
	}
	i := len(detail.Steps) - 1
	detail.Steps[i].Status = string(dbclient.StatusTypeFailed)
	detail.Steps[i].Reason = err.Error()
	content, err := json.Marshal(detail)
	if err != nil {
		logrus.Errorf("marshal record detail failed, error:%+v", err)
	}
	record.Status = dbclient.StatusTypeFailed
	record.Detail = string(content)
	if err := ctx.DB.RecordsWriter().Update(*record); err != nil {
		logrus.Errorf("failed to update record: %v", err)
	}

	if source == "addon" {
		if addonID != "" {
			_, err := ctx.Bdl.AddonConfigCallbackProvison(addonID, apistructs.AddonCreateCallBackResponse{IsSuccess: false})
			if err != nil {
				logrus.Errorf("add call back provision failed, error:%v", err)
			}
		} else {
			logrus.Errorf("addon with no addonID")
		}
	}
}
