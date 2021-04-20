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

package apistructs

import (
	"strconv"
	"time"
)

const SceneSetsAutotestExecType = "sceneSets"
const SceneAutotestExecType = "scene"

type SceneSet struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	SpaceID     uint64    `json:"spaceID"`
	PreID       uint64    `json:"preID"`
	Description string    `json:"description"`
	CreatorID   string    `json:"creatorID"`
	UpdaterID   string    `json:"updatorID"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// type SceneSetCreateRequest struct {
// 	Name        string `json:"name"`
// 	SpaceID     uint64 `json:"spaceID"`
// 	Description string `json:"description"`
// 	PreID       uint64 `json:"preID"`
// }

type SceneSetRequest struct {
	Name        string `json:"name"`
	SpaceID     uint64 `json:"spaceID"`
	Description string `json:"description"`
	PreID       uint64 `json:"preID"`
	SetID       uint64 `json:"setID"`
	DropKey     int64  `json:"dropKey"`
	Position    int64  `json:"position,omitempty"` // 插入位置
	ProjectId   uint64 `json:"projectID"`
	IdentityInfo
}

// type SceneSetUpdateRequest struct {
// 	Name        string `json:"name"`
// 	Description string `json:"description"`
// }

type GetSceneSetResponse struct {
	Header
	Data SceneSet
}

type CreateSceneSetResponse struct {
	Header
	Id uint64
}

type GetSceneSetsResponse struct {
	Header
	Data []SceneSet
}

type UpdateSceneSetResponse struct {
	Header
	Data SceneSet
}

type DeleteSceneSetResponse struct {
	Header
	res string
}

func (req *SceneSetRequest) URLQueryString() map[string][]string {
	query := make(map[string][]string)
	if req.Name != "" {
		query["name"] = append(query["name"], req.Name)
	}
	if req.SpaceID != 0 {
		query["spaceID"] = []string{strconv.FormatInt(int64(req.SpaceID), 10)}
	}
	if req.Description != "" {
		query["description"] = append(query["description"], req.Description)
	}
	if req.PreID != 0 {
		query["preID"] = []string{strconv.FormatInt(int64(req.PreID), 10)}
	}
	if req.DropKey != 0 {
		query["dropKey"] = []string{strconv.FormatInt(int64(req.DropKey), 10)}
	}
	return query
}
