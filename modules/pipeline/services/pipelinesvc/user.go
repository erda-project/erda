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

package pipelinesvc

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

// tryGetUser try to get user info from cmdb. If failed, return a basic user just with id.
// TODO later add cache here if need.
func (s *PipelineSvc) tryGetUser(userID string) *apistructs.PipelineUser {
	user, err := s.bdl.GetCurrentUser(userID)
	if err != nil {
		logrus.Warnf("failed to get user info, userID: %s, err: %v", userID, err)
		// return basic user just with id
		return &apistructs.PipelineUser{ID: userID}
	}
	if user == nil {
		logrus.Warnf("failed to get user info, userID: %s, err: %v", userID, fmt.Errorf("get empty user info"))
		// return basic user just with id
		return &apistructs.PipelineUser{ID: userID}
	}
	// return queried user
	return user.ConvertToPipelineUser()
}
