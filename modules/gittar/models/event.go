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

package models

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
)

func (svc *Service) TriggerEvent(repository *gitmodule.Repository, eventName string, contentData interface{}) {
	err := svc.bundle.CreateEvent(&apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			ApplicationID: strconv.FormatInt(repository.ApplicationId, 10),
			ProjectID:     strconv.FormatInt(repository.ProjectId, 10),
			OrgID:         strconv.FormatInt(repository.OrgId, 10),
			Event:         eventName,
			Action:        eventName,
		},
		Sender:  "gittar",
		Content: contentData,
	})
	if err != nil {
		logrus.Errorf("failed to trigger event %s err:%s", eventName, err)
	}
}
