// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sceneset

import (
	"fmt"
	"regexp"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	maxSize       int = 200
	nameMaxLength int = 50
	descMaxLength int = 255
)

func (svc *Service) CreateSceneSet(req apistructs.SceneSetRequest) (uint64, error) {
	if err := strutil.Validate(req.Name, strutil.MaxRuneCountValidator(nameMaxLength)); err != nil {
		return 0, err
	}
	if err := strutil.Validate(req.Description, strutil.MaxRuneCountValidator(descMaxLength)); err != nil {
		return 0, err
	}
	if ok, _ := regexp.MatchString("^[a-zA-Z\u4e00-\u9fa50-9_-]*$", req.Name); !ok {
		return 0, apierrors.ErrCreateAutoTestSceneSet.InvalidState("只可输入中文、英文、数字、中划线或下划线")
	}

	count, err := svc.db.CountSceneSetByName(req.Name, req.SpaceID)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, apierrors.ErrCreateAutoTestSceneSet.AlreadyExists()
	}

	sets, err := svc.GetSceneSetsBySpaceID(req.SpaceID)
	if err != nil {
		return 0, err
	}
	if len(sets) >= maxSize {
		return 0, fmt.Errorf("Reach max sceneset size!")
	}

	preID := uint64(0)
	if len(sets) > 0 {
		preID = sets[len(sets)-1].ID
	}

	sceneSet := dao.SceneSet{
		Name:        req.Name,
		Description: req.Description,
		SpaceID:     req.SpaceID,
		PreID:       preID,
		CreatorID:   req.UserID,
		UpdaterID:   req.UserID,
	}

	if err := svc.db.CreateSceneSet(&sceneSet); err != nil {
		return 0, apierrors.ErrCreateAutoTestSceneSet.InternalError(err)
	}
	return sceneSet.ID, nil
}

func (svc *Service) GetSceneSetsBySpaceID(spaceID uint64) ([]apistructs.SceneSet, error) {
	sceneSets, err := svc.db.SceneSetsBySpaceID(spaceID)
	if err != nil {
		return nil, apierrors.ErrListAutoTestSceneSet.InternalError(err)
	}

	setMap := make(map[uint64]apistructs.SceneSet)
	for _, v := range sceneSets {
		setMap[v.PreID] = *mapping(&v)
	}

	var res []apistructs.SceneSet
	for head := uint64(0); ; {
		s, ok := setMap[head]
		if !ok {
			break
		}
		head = s.ID
		res = append(res, s)
	}

	// res := make([]apistructs.SceneSet, 0, len(sceneSets))
	// for _, item := range sceneSets {
	// 	res = append(res, *mapping(&item))
	// }
	return res, nil
}

func (svc *Service) UpdateSceneSet(setID uint64, req apistructs.SceneSetRequest) (*apistructs.SceneSet, error) {
	if err := strutil.Validate(req.Name, strutil.MaxRuneCountValidator(nameMaxLength)); err != nil {
		return nil, err
	}
	if err := strutil.Validate(req.Description, strutil.MaxRuneCountValidator(descMaxLength)); err != nil {
		return nil, err
	}
	if ok, _ := regexp.MatchString("^[a-zA-Z\u4e00-\u9fa50-9_-]*$", req.Name); !ok {
		return nil, apierrors.ErrCreateAutoTestSceneSet.InvalidState("只可输入中文、英文、数字、中划线或下划线")
	}

	s, err := svc.db.GetSceneSet(setID)
	if err != nil {
		return nil, apierrors.ErrGetAutoTestSceneSet.InternalError(err)
	}

	count, err := svc.db.CountSceneSetByName(req.Name, req.SpaceID)
	if err != nil {
		return nil, err
	}
	if count > 1 || count == 1 && req.Name != s.Name {
		return nil, apierrors.ErrUpdateAutoTestSceneSet.AlreadyExists()
	}

	if len(req.Name) > 0 {
		s.Name = req.Name
	}
	s.Description = req.Description

	res, err := svc.db.UpdateSceneSet(s)
	if err != nil {
		return nil, apierrors.ErrDeleteAutoTestSceneSet.InternalError(err)
	}

	return mapping(res), nil
}

func (svc *Service) GetSceneSet(setID uint64) (*apistructs.SceneSet, error) {
	s, err := svc.db.GetSceneSet(setID)
	if err != nil {
		return nil, apierrors.ErrGetAutoTestSceneSet.InternalError(err)
	}
	return mapping(s), nil
}

func (svc *Service) DeleteSceneSet(req apistructs.SceneSetRequest) error {
	r, err := svc.db.CheckRelatedSceneSet(req.SetID)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneSet.InternalError(err)
	}
	if r {
		return fmt.Errorf("场景集合加入了测试计划, 无法删除")
	}

	s, err := svc.db.GetSceneSet(req.SetID)
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneSet.InternalError(err)
	}

	l, scenes, err := svc.GetScenes(apistructs.AutotestSceneRequest{SetID: req.SetID})
	if err != nil {
		return apierrors.ErrDeleteAutoTestSceneSet.InternalError(err)
	}

	ids := make([]uint64, l)
	for i, s := range scenes {
		ids[i] = s.ID
	}
	return svc.db.DeleteSceneSet(s, ids)
}

func (svc *Service) DragSceneSet(req apistructs.SceneSetRequest) error {
	if req.Position == 0 {
		return fmt.Errorf("Cannot drag sceneset into another!")
	}
	return svc.db.MoveSceneSet(req)
}

func (svc *Service) CopySceneSet(req apistructs.SceneSetRequest, isSpaceCopy bool) (uint64, error) {
	id := req.SetID
	set, err := svc.GetSceneSet(id)
	if err != nil {
		return 0, nil
	}

	newSet := &dao.SceneSet{
		Name:        set.Name,
		Description: set.Description,
		SpaceID:     req.SpaceID,
		PreID:       req.PreID,
		CreatorID:   req.UserID,
	}

	if err := svc.db.CreateSceneSet(newSet); err != nil {
		return 0, err
	}

	_, scenes, err := svc.GetScenes(apistructs.AutotestSceneRequest{SetID: id})
	preId := uint64(0)
	r := apistructs.AutotestSceneCopyRequest{
		PreID:   preId,
		SetID:   newSet.ID,
		SpaceID: req.SpaceID,
	}
	r.IdentityInfo = req.IdentityInfo

	var sceneIdMap = map[uint64]uint64{}
	for _, scene := range scenes {
		r.SceneID = scene.ID
		r.PreID = preId
		preId, err = svc.CopyScene(r, isSpaceCopy, sceneIdMap)
		if err != nil {
			return 0, err
		}
		sceneIdMap[scene.ID] = preId
	}
	return newSet.ID, nil
}

func mapping(s *dao.SceneSet) *apistructs.SceneSet {
	return &apistructs.SceneSet{
		ID:          s.ID,
		Name:        s.Name,
		SpaceID:     s.SpaceID,
		PreID:       s.PreID,
		Description: s.Description,
		CreatorID:   s.CreatorID,
		UpdaterID:   s.UpdaterID,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
