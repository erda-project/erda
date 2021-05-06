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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

const MAX_SIZE int = 200

func (svc *Service) CreateSceneSet(req apistructs.SceneSetRequest) (uint64, error) {
	sets, err := svc.GetSceneSetsBySpaceID(req.SpaceID)
	if err != nil {
		return 0, err
	}

	if len(sets) == MAX_SIZE {
		return 0, apierrors.ErrCreateAutoTestSceneSet.InternalError(err)
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
	s, err := svc.db.GetSceneSet(setID)
	if err != nil {
		return nil, apierrors.ErrGetAutoTestSceneSet.InternalError(err)
	}

	if len(req.Name) > 0 {
		s.Name = req.Name
	}
	if len(req.Description) > 0 {
		s.Description = req.Description
	}

	res, err := svc.db.UpdateSceneSet(s)
	if err != nil {
		return nil, apierrors.ErrUpdateAutoTestSceneSet.InternalError(err)
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
