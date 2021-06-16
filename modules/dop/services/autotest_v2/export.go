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

package autotestv2

import (
	"io"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

type AutoTestSpaceDB struct {
	Data *AutoTestSpaceData
}

func (a *AutoTestSpaceDB) CreateNewSpace() (*apistructs.AutoTestSpace, error) {
	var err error
	spaceName := a.Data.Space.Name
	if a.Data.IsCopy {
		spaceName, err = a.Data.svc.GenerateSpaceName(spaceName, int64(a.Data.ProjectID))
		if err != nil {
			return nil, err
		}
	}
	space, err := a.Data.svc.CreateSpace(apistructs.AutoTestSpaceCreateRequest{
		Name:         spaceName,
		ProjectID:    int64(a.Data.ProjectID),
		Description:  a.Data.Space.Description,
		IdentityInfo: a.Data.IdentityInfo,
	})
	a.Data.NewSpace = space
	return space, nil
}

func (a *AutoTestSpaceDB) SetSpace() error {
	space, err := a.Data.svc.GetSpace(a.Data.SpaceID)
	//space, err := a.svc.db.GetAutoTestSpace(a.Data.SpaceID)
	if err != nil {
		return err
	}
	a.Data.Space = space
	return nil
}

func (a *AutoTestSpaceDB) SetSceneSets() error {
	a.Data.SceneSets = map[uint64][]apistructs.SceneSet{}
	sceneSets, err := a.Data.svc.GetSceneSetsBySpaceID(a.Data.Space.ID)
	if err != nil {
		return err
	}
	a.Data.SceneSets[a.Data.Space.ID] = sceneSets
	return nil
}

func (a *AutoTestSpaceDB) SetScenes() error {
	a.Data.Scenes = map[uint64][]apistructs.AutoTestScene{}
	for _, v := range a.Data.SceneSets[a.Data.Space.ID] {
		_, scenes, err := a.Data.svc.ListAutotestScene(apistructs.AutotestSceneRequest{
			SetID:        v.ID,
			IdentityInfo: a.Data.IdentityInfo,
		})
		if err != nil {
			return err
		}
		sceneIDs := []uint64{}
		for _, v := range scenes {
			sceneIDs = append(sceneIDs, v.ID)
		}
		inputs, err := a.Data.svc.ListAutoTestSceneInputByScenes(sceneIDs)
		if err != nil {
			return err
		}
		inputMap := map[uint64][]apistructs.AutoTestSceneInput{}
		for _, input := range inputs {
			inputMap[input.SceneID] = append(inputMap[input.SceneID], input)
		}
		outputs, err := a.Data.svc.ListAutoTestSceneOutputByScenes(sceneIDs)
		if err != nil {
			return err
		}
		outputMap := map[uint64][]apistructs.AutoTestSceneOutput{}
		for _, output := range outputs {
			outputMap[output.SceneID] = append(outputMap[output.SceneID], output)
		}
		for i := 0; i < len(scenes); i++ {
			scenes[i].Inputs = inputMap[scenes[i].ID]
			scenes[i].Output = outputMap[scenes[i].ID]
		}
		a.Data.Scenes[v.ID] = scenes
	}
	return nil
}

func (a *AutoTestSpaceDB) SetSceneSteps() error {
	a.Data.Steps = map[uint64][]apistructs.AutoTestSceneStep{}
	for _, scenes := range a.Data.Scenes {
		for _, scene := range scenes {
			steps, err := a.Data.svc.ListAutoTestSceneStep(scene.ID)
			if err != nil {
				return err
			}
			a.Data.Steps[scene.ID] = steps
		}
	}
	return nil
}

func (a *AutoTestSpaceDB) SetConfigs() error {
	var autoTestGlobalConfigListRequest apistructs.AutoTestGlobalConfigListRequest
	autoTestGlobalConfigListRequest.ScopeID = strconv.Itoa(int(a.Data.Space.ProjectID))
	autoTestGlobalConfigListRequest.Scope = "project-autotest-testcase"
	autoTestGlobalConfigListRequest.UserID = a.Data.UserID
	configs, err := a.Data.svc.bdl.ListAutoTestGlobalConfig(autoTestGlobalConfigListRequest)
	if err != nil {
		return err
	}
	a.Data.Configs = configs
	return nil
}

func (a *AutoTestSpaceDB) GetSpaceData() *AutoTestSpaceData {
	return a.Data
}

func (svc *Service) Export(w io.Writer, req apistructs.AutoTestSpaceExportRequest) error {
	// check parameter
	if !req.FileType.Valid() {
		return apierrors.ErrExportAutoTestSpace.InvalidParameter("fileType")
	}

	spaceDBData := AutoTestSpaceDB{Data: &AutoTestSpaceData{
		svc:          svc,
		IdentityInfo: req.IdentityInfo,
		Locale:       req.Locale,
		SpaceID:      req.ID,
		IsCopy:       req.IsCopy,
	},
	}

	creator := AutoTestSpaceDirector{}
	creator.New(&spaceDBData)
	if err := creator.Construct(); err != nil {
		return err
	}
	spaceData := creator.Creator.GetSpaceData()

	return spaceData.ConvertToExcel(w)
}
