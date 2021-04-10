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

package testplan

import (
	"github.com/erda-project/erda/apistructs"
	dao2 "github.com/erda-project/erda/modules/qa/dao"
)

func (t *TestPlan) ConvertMember(dbMem dao2.TestPlanMember) apistructs.TestPlanMember {
	return apistructs.TestPlanMember{
		ID:         uint64(dbMem.ID),
		TestPlanID: dbMem.TestPlanID,
		Role:       dbMem.Role,
		UserID:     dbMem.UserID,
		CreatedAt:  dbMem.CreatedAt,
		UpdatedAt:  dbMem.UpdatedAt,
	}
}

func (t *TestPlan) BatchConvertMembers(dbMems []dao2.TestPlanMember) []apistructs.TestPlanMember {
	var results []apistructs.TestPlanMember
	for _, dbMem := range dbMems {
		results = append(results, t.ConvertMember(dbMem))
	}
	return results
}
