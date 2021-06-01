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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/pkg/excel"
)

// AutoTestSpaceExcel convert excel data to space datas
// will achieve AutoTestSpaceDataCreator implement
type AutoTestSpaceExcel struct {
	sheets [][][]string
	Data   *AutoTestSpaceData
}

func convertToUint64PermitZero(idStr string) (uint64, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}
func convertStrIDToUint64(idStr string) (uint64, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return uint64(id), nil
}

// SetSpace get space data from sheets[0]
func (a *AutoTestSpaceExcel) SetSpace() error {
	var err error
	spaceSheet := a.sheets[0]
	if len(spaceSheet) != 2 {
		return fmt.Errorf("missing space data")
	}
	if len(spaceSheet[1]) != 4 {
		return fmt.Errorf("invalid space data row")
	}
	space := apistructs.AutoTestSpace{}
	space.ID, err = convertStrIDToUint64(spaceSheet[1][0])
	if err != nil {
		return err
	}
	space.Name = spaceSheet[1][1]
	space.ProjectID = int64(a.Data.ProjectID)
	space.Description = spaceSheet[1][3]
	a.Data.Space = &space
	return nil
}

// SetSceneSets get scene sets data from sheets[1]
func (a *AutoTestSpaceExcel) SetSceneSets() error {
	var err error
	sheet := a.sheets[1]
	if len(sheet) < 1 {
		return fmt.Errorf("missing scene set data")
	}
	a.Data.SceneSets = map[uint64][]apistructs.SceneSet{}
	for _, setRow := range sheet[1:] {
		if len(setRow) != 5 {
			return fmt.Errorf("invalid scene set data")
		}
		sceneSet := apistructs.SceneSet{}
		sceneSet.ID, err = convertStrIDToUint64(setRow[0])
		if err != nil {
			return err
		}
		sceneSet.Name = setRow[1]
		sceneSet.SpaceID, err = convertStrIDToUint64(setRow[2])
		if err != nil {
			return err
		}
		if sceneSet.SpaceID != a.Data.Space.ID {
			return fmt.Errorf("scene set id don`t match space id")
		}
		sceneSet.PreID, err = convertToUint64PermitZero(setRow[3])
		if err != nil {
			return err
		}
		sceneSet.Description = setRow[4]
		a.Data.SceneSets[a.Data.Space.ID] = append(a.Data.SceneSets[a.Data.Space.ID], sceneSet)
	}
	return nil
}

// SetScenes get scenes, input, output data
// and insert input, output data to scenes
func (a *AutoTestSpaceExcel) SetScenes() error {
	var err error
	a.Data.Scenes = map[uint64][]apistructs.AutoTestScene{}

	inputSheet := a.sheets[3]
	if len(inputSheet) < 1 {
		return fmt.Errorf("missing input data")
	}
	inputList := []apistructs.AutoTestSceneInput{}
	for _, inputRow := range inputSheet[1:] {
		if len(inputRow) != 7 {
			return fmt.Errorf("invalid input data")
		}
		input := apistructs.AutoTestSceneInput{}
		input.ID, err = convertStrIDToUint64(inputRow[0])
		if err != nil {
			return err
		}
		input.Name = inputRow[1]
		input.Value = inputRow[2]
		input.Temp = inputRow[3]
		input.SceneID, err = convertStrIDToUint64(inputRow[4])
		if err != nil {
			return err
		}
		input.SpaceID, err = convertToUint64PermitZero(inputRow[5])
		if err != nil {
			return err
		}
		input.Description = inputRow[6]
		inputList = append(inputList, input)
	}

	outputSHeet := a.sheets[4]
	if len(outputSHeet) < 1 {
		return fmt.Errorf("missing output data")
	}
	outputList := []apistructs.AutoTestSceneOutput{}
	for _, outputRow := range outputSHeet[1:] {
		output := apistructs.AutoTestSceneOutput{}
		output.ID, err = convertStrIDToUint64(outputRow[0])
		if err != nil {
			return err
		}
		output.Name = outputRow[1]
		output.Value = outputRow[2]
		output.SceneID, err = convertStrIDToUint64(outputRow[3])
		if err != nil {
			return err
		}
		output.SpaceID, err = convertToUint64PermitZero(outputRow[4])
		if err != nil {
			return err
		}
		output.Description = outputRow[5]
		outputList = append(outputList, output)
	}

	sheet := a.sheets[2]
	if len(sheet) < 1 {
		return fmt.Errorf("missing scene data")
	}
	for _, sceneRow := range sheet[1:] {
		if len(sceneRow) != 7 {
			return fmt.Errorf("invalid scene data")
		}
		scene := apistructs.AutoTestScene{}
		scene.ID, err = convertStrIDToUint64(sceneRow[0])
		if err != nil {
			return err
		}
		scene.Name = sceneRow[1]
		scene.SetID, err = convertStrIDToUint64(sceneRow[2])
		if err != nil {
			return err
		}
		scene.SpaceID, err = convertStrIDToUint64(sceneRow[3])
		if err != nil {
			return err
		}
		if scene.SpaceID != a.Data.Space.ID {
			return fmt.Errorf("scene`s space id don not match")
		}
		scene.PreID, err = convertToUint64PermitZero(sceneRow[4])
		if err != nil {
			return err
		}
		scene.RefSetID, err = convertToUint64PermitZero(sceneRow[5])
		if err != nil {
			return err
		}
		scene.Description = sceneRow[6]
		for _, input := range inputList {
			if input.SceneID == scene.ID {
				scene.Inputs = append(scene.Inputs, input)
			}
		}
		for _, output := range outputList {
			if output.SceneID == scene.ID {
				scene.Output = append(scene.Output, output)
			}
		}
		if a.Data.Scenes[scene.SetID] == nil {
			a.Data.Scenes[scene.SetID] = []apistructs.AutoTestScene{}
		}
		a.Data.Scenes[scene.SetID] = append(a.Data.Scenes[scene.SetID], scene)
	}
	return nil
}

//SetSceneSteps get steps data and judge step`s children
func (a *AutoTestSpaceExcel) SetSceneSteps() error {
	var err error
	a.Data.Steps = map[uint64][]apistructs.AutoTestSceneStep{}
	sheet := a.sheets[5]

	if len(sheet) < 1 {
		return fmt.Errorf("missing scene step data")
	}

	for _, stepRow := range sheet[1:] {
		if len(stepRow) != 9 {
			return fmt.Errorf("invalid scene step data")
		}
		step := apistructs.AutoTestSceneStep{}
		step.ID, err = convertStrIDToUint64(stepRow[0])
		if err != nil {
			return err
		}
		step.Name = stepRow[1]
		step.Value = stepRow[2]
		step.Type = apistructs.StepAPIType(stepRow[3])
		step.PreID, err = convertToUint64PermitZero(stepRow[4])
		if err != nil {
			return err
		}
		step.SceneID, err = convertStrIDToUint64(stepRow[5])
		if err != nil {
			return err
		}
		step.SpaceID, err = convertStrIDToUint64(stepRow[6])
		if err != nil {
			return err
		}
		if step.SpaceID != a.Data.Space.ID {
			return fmt.Errorf("scene step space id don not match")
		}
		step.PreType = apistructs.PreType(stepRow[7])
		step.APISpecID, err = convertToUint64PermitZero(stepRow[8])
		if err != nil {
			return err
		}
		if a.Data.Steps[step.SceneID] == nil {
			a.Data.Steps[step.SceneID] = []apistructs.AutoTestSceneStep{}
		}
		a.Data.Steps[step.SceneID] = append(a.Data.Steps[step.SceneID], step)
	}

	for sceneID, scs := range a.Data.Steps {
		type idType struct {
			PreID   uint64
			PreType apistructs.PreType
		}
		stepMap := make(map[idType]*apistructs.AutoTestSceneStep)
		for i := range scs {
			stepMap[idType{scs[i].PreID, scs[i].PreType}] = &scs[i]
		}
		var steps []apistructs.AutoTestSceneStep
		for head := uint64(0); ; {
			s, ok := stepMap[idType{head, apistructs.PreTypeSerial}]
			if !ok {
				break
			}
			head = s.ID
			for head2 := s.ID; ; {
				s2, ok := stepMap[idType{head2, apistructs.PreTypeParallel}]
				if !ok {
					break
				}
				head2 = s2.ID
				s.Children = append(s.Children, *s2)
			}
			steps = append(steps, *s)
		}
		a.Data.Steps[sceneID] = steps
	}

	return nil
}

// SetConfigs get api configs
func (a *AutoTestSpaceExcel) SetConfigs() error {
	sheet := a.sheets[6]
	a.Data.Configs = []apistructs.AutoTestGlobalConfig{}
	if len(sheet) < 1 {
		return fmt.Errorf("missing config data")
	}
	for _, configRow := range sheet[1:] {
		if len(configRow) != 7 {
			return fmt.Errorf("invalid config data")
		}
		config := apistructs.AutoTestGlobalConfig{}
		config.Scope = configRow[0]
		config.ScopeID = configRow[1]
		config.Ns = configRow[2]
		config.DisplayName = configRow[3]
		config.Desc = configRow[4]
		config.APIConfig = &apistructs.AutoTestAPIConfig{Global: map[string]apistructs.AutoTestConfigItem{}}
		config.APIConfig.Domain = configRow[5]
		headerStr := configRow[6]
		var header map[string]string
		if err := json.Unmarshal([]byte(headerStr), &header); err != nil {
			return err
		}
		config.APIConfig.Header = header
		a.Data.Configs = append(a.Data.Configs, config)
	}

	itemConfigSheet := a.sheets[7]
	if len(itemConfigSheet) < 1 {
		return fmt.Errorf("missing api config data")
	}
	for _, itemConfigRow := range itemConfigSheet[1:] {
		if len(itemConfigRow) != 5 {
			return fmt.Errorf("invalid api config item")
		}
		item := apistructs.AutoTestConfigItem{}
		item.Name = itemConfigRow[1]
		item.Value = itemConfigRow[3]
		item.Type = itemConfigRow[2]
		item.Desc = itemConfigRow[4]
		for _, config := range a.Data.Configs {
			if config.Ns == itemConfigRow[0] {
				config.APIConfig.Global[item.Name] = item
			}
		}
	}
	return nil
}

func (a *AutoTestSpaceExcel) GetSpaceData() *AutoTestSpaceData {
	return a.Data
}

func (svc *Service) Import(req apistructs.AutoTestSpaceImportRequest, r *http.Request) (*apistructs.AutoTestSpace, error) {
	if !req.FileType.Valid() {
		return nil, apierrors.ErrImportAutoTestSpace.InvalidParameter("fileType")
	}
	if req.ProjectID == 0 {
		return nil, apierrors.ErrImportAutoTestSpace.MissingParameter("projectID")
	}

	_, err := svc.bdl.GetProject(req.ProjectID)
	if err != nil {
		return nil, apierrors.ErrImportAutoTestSpace.InvalidParameter(fmt.Errorf("project not found, id: %d", req.ProjectID))
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	switch req.FileType {
	case apistructs.TestSpaceFileTypeExcel:
		sheets, err := excel.Decode(f)
		if err != nil {
			return nil, err
		}
		if len(sheets) != 8 {
			return nil, apierrors.ErrImportAutoTestSpace.InvalidParameter("sheet")
		}
		spaceExcelData := AutoTestSpaceExcel{
			sheets: sheets,
			Data: &AutoTestSpaceData{
				ProjectID:    req.ProjectID,
				IdentityInfo: req.IdentityInfo,
				svc:          svc,
			},
		}
		// make space data creator
		creator := AutoTestSpaceDirector{}
		creator.New(&spaceExcelData)
		if err := creator.Construct(); err != nil {
			return nil, err
		}
		data := creator.Creator.GetSpaceData()
		space, err := data.Copy()
		if err != nil {
			return nil, err
		}
		return space, nil
	default:
		return nil, apierrors.ErrImportAutoTestSpace.InvalidParameter("fileType")
	}
}
