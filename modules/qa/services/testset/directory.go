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

package testset

import (
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
)

// GenerateTestSetName 生成测试集名，追加 (N)
func (svc *Service) GenerateTestSetName(projectID, parentTsID, testSetID uint64, requestName string) (string, error) {
	finalName := requestName
	for {
		// find by name
		exist, err := svc.db.GetTestSetByNameAndParentIDAndProjectID(projectID, parentTsID, apistructs.RecycledNo, finalName)
		if err != nil {
			return "", err
		}
		// not exist
		if exist == nil {
			return finalName, nil
		}
		// exist but is itself
		if exist.ID == testSetID {
			return finalName, nil
		}
		// exist a ain
		finalName, err = getTitleName(finalName)
		if err != nil {
			return "", err
		}
	}
}

// generateTestSetDirectory 根据父测试集生成目录
func generateTestSetDirectory(parentSet *dao.TestSet, name string) string {
	parentDirectory := "/"
	if parentSet != nil {
		parentDirectory = parentSet.Directory
	}
	return filepath.Join("/", parentDirectory, name)
}

// updateChildDirectory 更新子测试集路径
func (svc *Service) updateChildDirectory(projectID, testSetID uint64) error {
	ts, err := svc.ensureGetTestSet(projectID, testSetID)
	if err != nil {
		return err
	}
	// 更新子测试集路径名称
	testSets, err := svc.db.ListTestSets(apistructs.TestSetListRequest{
		Recycled:   apistructs.RecycledNo,
		ParentID:   &testSetID,
		ProjectID:  &projectID,
		TestSetIDs: nil,
	})
	if err != nil {
		return err
	}
	for _, item := range testSets {
		item.Directory = generateTestSetDirectory(ts, item.Name)
		item.UpdaterID = ts.UpdaterID
		if err := svc.db.UpdateTestSet(&item); err != nil {
			logrus.Errorf("update child testset error, parentID is %d, childID is %d", ts.ID, item.ID)
		}

		svc.updateChildDirectory(projectID, item.ID)
	}
	return nil
}
