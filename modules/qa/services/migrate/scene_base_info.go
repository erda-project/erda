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

package migrate

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
)

// SceneBaseInfo 场景基础信息
type SceneBaseInfo struct {
	Space *dao.AutoTestSpace

	LastSceneSet *dao.SceneSet
	AllSceneSets []*dao.SceneSet

	AllSceneMap map[uint64][]*dao.AutoTestScene // key: sceneSetID, value: scenes, last in the end
}

func (si *SceneBaseInfo) GetPreSceneIDUnderSceneSet(sceneSetID uint64) uint64 {
	if scenes, ok := si.AllSceneMap[sceneSetID]; ok && len(scenes) > 0 {
		return scenes[len(scenes)-1].ID
	}
	return 0
}

func (si *SceneBaseInfo) GetPreSceneSetID() uint64 {
	if si.LastSceneSet != nil {
		return si.LastSceneSet.ID
	}
	return 0
}

// createSceneBaseInfo 创建场景所需的基础信息
func (svc *Service) createSceneBaseInfo(projectID uint64) (*SceneBaseInfo, error) {
	// 创建默认空间
	space, err := svc.db.CreateAutoTestSpace(&dao.AutoTestSpace{
		Name:          "Default Space",
		ProjectID:     int64(projectID),
		Description:   "平台升级自动创建的测试空间",
		CreatorID:     systemUserID,
		UpdaterID:     systemUserID,
		Status:        apistructs.TestSpaceOpen,
		SourceSpaceID: nil,
		DeletedAt:     nil,
	})
	if err != nil {
		return nil, printReturnErr(fmt.Errorf("failed to create default space, err: %v", err))
	}
	sceneBaseInfo := SceneBaseInfo{
		Space:       space,
		AllSceneMap: make(map[uint64][]*dao.AutoTestScene),
	}
	return &sceneBaseInfo, nil
}

func (si *SceneBaseInfo) appendSceneUnderSet(scene *dao.AutoTestScene) {
	si.AllSceneMap[scene.SetID] = append(si.AllSceneMap[scene.SetID], scene)
}
