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

package testcase

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

// GetTestCase 获取测试用例详情
func (svc *Service) GetTestCase(tcID uint64) (*apistructs.TestCase, error) {
	model, err := svc.db.GetTestCaseByID(tcID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.ErrGetTestCase.NotFound()
		}
		return nil, apierrors.ErrGetTestCase.InternalError(err)
	}
	return svc.convertTestCase(*model)
}
