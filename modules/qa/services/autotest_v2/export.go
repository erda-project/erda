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
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/modules/qa/services/i18n"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/jsonparse"
	"github.com/erda-project/erda/pkg/strutil"
)

type AutoTestSpaceData struct {
	req       apistructs.AutoTestSpaceExportRequest
	Space     apistructs.AutoTestSpace
	SceneSets map[uint64][]apistructs.SceneSet
	Scenes    map[uint64][]apistructs.AutoTestScene
	Steps     map[uint64][]apistructs.AutoTestSceneStep
	Configs   []apistructs.AutoTestGlobalConfig
}

type AutoTestSpaceDataCreator interface {
	SetSpace() error
	SetSceneSets() error
	SetScenes() error
	SetSceneSteps() error
	SetConfigs() error
	GetSpaceData() *AutoTestSpaceData
	ConvertToExcel(w io.Writer) error
}

type AutoTestSpaceDirector struct {
	Creator AutoTestSpaceDataCreator
}

func (a *AutoTestSpaceDirector) New(m AutoTestSpaceDataCreator) {
	a.Creator = m
}

func (a *AutoTestSpaceDirector) Construct() error {
	if err := a.Creator.SetSpace(); err != nil {
		return err
	}

	if err := a.Creator.SetSceneSets(); err != nil {
		return err
	}

	if err := a.Creator.SetScenes(); err != nil {
		return err
	}

	if err := a.Creator.SetSceneSteps(); err != nil {
		return err
	}

	if err := a.Creator.SetConfigs(); err != nil {
		return err
	}

	return nil
}

type AutoTestSpaceDB struct {
	svc  *Service
	Data *AutoTestSpaceData
}

func (a *AutoTestSpaceDB) SetSpace() error {
	space, err := a.svc.GetSpace(a.Data.req.ID)
	//space, err := a.svc.db.GetAutoTestSpace(a.Data.SpaceID)
	if err != nil {
		return err
	}
	a.Data.Space = *space
	return nil
}

func (a *AutoTestSpaceDB) SetSceneSets() error {
	sceneSets, err := a.svc.db.SceneSetsBySpaceID(a.Data.Space.ID)
	if err != nil {
		return err
	}

	setMap := make(map[uint64]apistructs.SceneSet)
	for _, v := range sceneSets {
		setMap[v.PreID] = *mapping(&v)
	}

	a.Data.SceneSets = map[uint64][]apistructs.SceneSet{}
	for head := uint64(0); ; {
		s, ok := setMap[head]
		if !ok {
			break
		}
		head = s.ID
		a.Data.SceneSets[a.Data.Space.ID] = append(a.Data.SceneSets[a.Data.Space.ID], s)
	}
	return nil
}

func (a *AutoTestSpaceDB) SetScenes() error {
	var setIDs []uint64
	for _, v := range a.Data.SceneSets[a.Data.Space.ID] {
		setIDs = append(setIDs, v.ID)
	}
	scenes, err := a.svc.ListAutotestScenes(setIDs)
	if err != nil {
		return err
	}
	a.Data.Scenes = scenes
	return nil
}

func (a *AutoTestSpaceDB) SetSceneSteps() error {
	a.Data.Steps = map[uint64][]apistructs.AutoTestSceneStep{}
	for _, scenes := range a.Data.Scenes {
		for _, scene := range scenes {
			steps, err := a.svc.ListAutoTestSceneStep(scene.ID)
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
	autoTestGlobalConfigListRequest.UserID = a.Data.req.UserID
	configs, err := a.svc.bdl.ListAutoTestGlobalConfig(autoTestGlobalConfigListRequest)
	if err != nil {
		return err
	}
	a.Data.Configs = configs
	return nil
}

func (a *AutoTestSpaceDB) addSpaceToExcel(file *excel.XlsxFile) error {
	l := a.svc.bdl.GetLocale(a.Data.req.Locale)
	title := []excel.Cell{
		excel.NewCell(l.Get(i18n.I18nKeySpaceNum)),
		excel.NewCell(l.Get(i18n.I18nKeySpaceName)),
		excel.NewCell(l.Get(i18n.I18nKeySpaceProjectNum)),
		excel.NewCell(l.Get(i18n.I18nKeySpaceDescription)),
	}

	allLines := [][]excel.Cell{title}
	spaceLine := []excel.Cell{
		excel.NewCell(strutil.String(a.Data.Space.ID)),
		excel.NewCell(a.Data.Space.Name),
		excel.NewCell(strutil.String(a.Data.Space.ProjectID)),
		excel.NewCell("djfdfjklj"),
	}
	allLines = append(allLines, spaceLine)
	allLines = append(allLines, []excel.Cell{
		excel.NewCell("djflkejlk"),
		excel.NewCell(""),
		excel.NewCell(""),
		excel.NewCell("jelfjlejl"),
	})
	sheetName := l.Get(i18n.I18nKeySpaceSheetName)

	return excel.AddSheetByCell(file, allLines, sheetName)
}

func (a *AutoTestSpaceDB) addSceneSetToExcel(file *excel.XlsxFile) error {
	l := a.svc.bdl.GetLocale(a.Data.req.Locale)
	title := []excel.Cell{
		excel.NewCell(l.Get(i18n.I18nKeySceneSetNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneSetName)),
		excel.NewCell(l.Get(i18n.I18nKeySceneSetSpaceNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneSetPreNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneSetDescription)),
	}
	allLines := [][]excel.Cell{title}
	for _, sceneSets := range a.Data.SceneSets {
		sceneSetLines := [][]excel.Cell{}
		for _, sceneSet := range sceneSets {
			sceneSetLines = append(sceneSetLines, []excel.Cell{
				excel.NewCell(strutil.String(sceneSet.ID)),
				excel.NewCell(sceneSet.Name),
				excel.NewCell(strutil.String(sceneSet.SpaceID)),
				excel.NewCell(strutil.String(sceneSet.PreID)),
				excel.NewCell(sceneSet.Description),
			})
		}
		allLines = append(allLines, sceneSetLines...)
	}
	sheetName := l.Get(i18n.I18nKeySceneSetSheetName)

	return excel.AddSheetByCell(file, allLines, sheetName)
}

func (a *AutoTestSpaceDB) addSceneToExcel(file *excel.XlsxFile) error {
	l := a.svc.bdl.GetLocale(a.Data.req.Locale)

	sceneTitle := []excel.Cell{
		excel.NewCell(l.Get(i18n.I18nKeySceneNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneName)),
		excel.NewCell(l.Get(i18n.I18nKeySceneSpaceNum)),
		excel.NewCell(l.Get(i18n.I18nKeyScenePreNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneDescription)),
	}

	inputTitle := []excel.Cell{
		excel.NewCell(l.Get(i18n.I18nKeySceneInputNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneInputName)),
		excel.NewCell(l.Get(i18n.I18nKeySceneInputValue)),
		excel.NewCell(l.Get(i18n.I18nKeySceneInputTemp)),
		excel.NewCell(l.Get(i18n.I18nKeySceneInputSceneNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneInputSpaceNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneInputDescription)),
	}

	outputTitle := []excel.Cell{
		excel.NewCell(l.Get(i18n.I18nKeySceneOutputNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneOutputName)),
		excel.NewCell(l.Get(i18n.I18nKeySceneOutputValue)),
		excel.NewCell(l.Get(i18n.I18nKeySceneOutputSceneNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneOutputSpaceNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneOutputDescription)),
	}

	allSceneLines := [][]excel.Cell{sceneTitle}
	allInputLines := [][]excel.Cell{inputTitle}
	allOutputLines := [][]excel.Cell{outputTitle}
	for _, scenes := range a.Data.Scenes {
		sceneLines := [][]excel.Cell{}
		for _, scene := range scenes {
			sceneLines = append(sceneLines, []excel.Cell{
				excel.NewCell(strutil.String(scene.ID)),
				excel.NewCell(scene.Name),
				excel.NewCell(strutil.String(scene.SpaceID)),
				excel.NewCell(strutil.String(scene.PreID)),
				excel.NewCell(scene.Description),
			})

			inputLines := [][]excel.Cell{}
			for _, input := range scene.Inputs {
				inputLines = append(inputLines, []excel.Cell{
					excel.NewCell(strutil.String(input.ID)),
					excel.NewCell(input.Name),
					excel.NewCell(input.Value),
					excel.NewCell(input.Temp),
					excel.NewCell(strutil.String(input.SceneID)),
					excel.NewCell(strutil.String(input.SpaceID)),
					excel.NewCell(input.Description),
				})
			}
			allInputLines = append(allInputLines, inputLines...)

			outputLines := [][]excel.Cell{}
			for _, output := range scene.Output {
				outputLines = append(outputLines, []excel.Cell{
					excel.NewCell(strutil.String(output.ID)),
					excel.NewCell(output.Name),
					excel.NewCell(output.Value),
					excel.NewCell(strutil.String(output.SceneID)),
					excel.NewCell(strutil.String(output.SpaceID)),
					excel.NewCell(output.Description),
				})
			}
			allOutputLines = append(allOutputLines, outputLines...)
		}
		allSceneLines = append(allSceneLines, sceneLines...)
	}
	sceneSheetName := l.Get(i18n.I18nKeySceneSheetName)
	inputSheetName := l.Get(i18n.I18nKeySceneInputSheetName)
	outputSheetName := l.Get(i18n.I18nKeySceneOutputSheetName)
	if err := excel.AddSheetByCell(file, allSceneLines, sceneSheetName); err != nil {
		return err
	}
	if err := excel.AddSheetByCell(file, allInputLines, inputSheetName); err != nil {
		return err
	}

	return excel.AddSheetByCell(file, allOutputLines, outputSheetName)
}

func (a *AutoTestSpaceDB) addSceneStepToExcel(file *excel.XlsxFile) error {
	l := a.svc.bdl.GetLocale(a.Data.req.Locale)
	title := []excel.Cell{
		excel.NewCell(l.Get(i18n.I18nKeySceneStepNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneStepName)),
		excel.NewCell(l.Get(i18n.I18nKeySceneStepValue)),
		excel.NewCell(l.Get(i18n.I18nKeySceneStepType)),
		excel.NewCell(l.Get(i18n.I18nKeySceneStepPreNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneStepSceneNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneStepSpaceNum)),
		excel.NewCell(l.Get(i18n.I18nKeySceneStepPreType)),
		excel.NewCell(l.Get(i18n.I18nKeySceneStepApiSpecNum)),
	}

	allLines := [][]excel.Cell{title}
	for _, steps := range a.Data.Steps {
		stepLines := [][]excel.Cell{}
		for _, step := range steps {
			stepLines = append(stepLines, []excel.Cell{
				excel.NewCell(strutil.String(step.ID)),
				excel.NewCell(step.Name),
				excel.NewCell(step.Value),
				excel.NewCell(step.Type.String()),
				excel.NewCell(strutil.String(step.PreID)),
				excel.NewCell(strutil.String(step.SceneID)),
				excel.NewCell(strutil.String(step.SpaceID)),
				excel.NewCell(strutil.String(step.PreType)),
				excel.NewCell(strutil.String(step.APISpecID)),
			})
		}
		allLines = append(allLines, stepLines...)
	}
	sheetName := l.Get(i18n.I18nKeySceneStepSheetName)

	return excel.AddSheetByCell(file, allLines, sheetName)
}

func (a *AutoTestSpaceDB) addConfigsToExcel(file *excel.XlsxFile) error {
	l := a.svc.bdl.GetLocale(a.Data.req.Locale)
	configTitle := []excel.Cell{
		excel.NewCell(l.Get(i18n.I18nKeyConfigScopeName)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigScopeNum)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigNsName)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigDisplayName)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigDescription)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigDomain)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigHeader)),
	}
	globalConfigTitle := []excel.Cell{
		excel.NewCell(l.Get(i18n.I18nKeyConfigGlobalNsName)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigGlobalName)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigGlobalType)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigGlobalValue)),
		excel.NewCell(l.Get(i18n.I18nKeyConfigGlobalDescription)),
	}

	allConfigLines := [][]excel.Cell{configTitle}
	allGlobalConfigLines := [][]excel.Cell{globalConfigTitle}

	for _, config := range a.Data.Configs {
		allConfigLines = append(allConfigLines, []excel.Cell{
			excel.NewCell(config.Scope),
			excel.NewCell(config.ScopeID),
			excel.NewCell(config.Ns),
			excel.NewCell(config.DisplayName),
			excel.NewCell(config.Desc),
			excel.NewCell(config.APIConfig.Domain),
			excel.NewCell(jsonparse.JsonOneLine(config.APIConfig.Header)),
		})
		gConfigLines := [][]excel.Cell{}
		for _, gConfig := range config.APIConfig.Global {
			gConfigLines = append(gConfigLines, []excel.Cell{
				excel.NewCell(config.Ns),
				excel.NewCell(gConfig.Name),
				excel.NewCell(gConfig.Type),
				excel.NewCell(gConfig.Value),
				excel.NewCell(gConfig.Desc),
			})
		}
		allGlobalConfigLines = append(allGlobalConfigLines, gConfigLines...)
	}

	configSheetName := l.Get(i18n.I18nKeyConfigSheetName)
	gConfigSheetName := l.Get(i18n.I18nKeyConfigGlobalSheetName)
	if err := excel.AddSheetByCell(file, allConfigLines, configSheetName); err != nil {
		return err
	}
	return excel.AddSheetByCell(file, allGlobalConfigLines, gConfigSheetName)
}

func (a *AutoTestSpaceDB) ConvertToExcel(w io.Writer) error {
	l := a.svc.bdl.GetLocale(a.Data.req.Locale)
	file := excel.NewXLSXFile()
	if err := a.addSpaceToExcel(file); err != nil {
		return err
	}
	if err := a.addSceneSetToExcel(file); err != nil {
		return err
	}
	if err := a.addSceneToExcel(file); err != nil {
		return err
	}
	if err := a.addSceneStepToExcel(file); err != nil {
		return err
	}
	if err := a.addConfigsToExcel(file); err != nil {
		return err
	}
	excel.WriteFile(w, file, l.Get(i18n.I18nKeySpaceSheetName))
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
		req: req,
	},
		svc: svc,
	}

	creator := AutoTestSpaceDirector{}
	creator.New(&spaceDBData)
	if err := creator.Construct(); err != nil {
		return err
	}

	return spaceDBData.ConvertToExcel(w)
}
