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

package apistructs

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// TestPlanV2 testplan
type TestPlanV2 struct {
	ID        uint64            `json:"id"`
	Name      string            `json:"name"`
	Desc      string            `json:"desc"`
	ProjectID uint64            `json:"project"`
	SpaceID   uint64            `json:"spaceID"`
	SpaceName string            `json:"spaceName"`
	Creator   string            `json:"creator"`
	Owners    []string          `json:"owners"`
	Updater   string            `json:"updater"`
	Steps     []*TestPlanV2Step `json:"steps"`
	CreateAt  *time.Time        `json:"createAt"`
	UpdateAt  *time.Time        `json:"updateAt"`
}

// TestPlanV2CreateRequest testplan v2 create request
type TestPlanV2CreateRequest struct {
	Name      string   `json:"name"`
	Desc      string   `json:"desc"`
	Owners    []string `json:"owners"`
	ProjectID uint64   `json:"projectID"`
	SpaceID   uint64   `json:"spaceID"`

	IdentityInfo
}

// Check check create request is valid
func (tp *TestPlanV2CreateRequest) Check() error {
	// req params check
	if tp.Name == "" {
		return errors.New("name is empty")
	}
	if tp.ProjectID == 0 {
		return errors.New("projectID is empty")
	}
	if tp.SpaceID == 0 {
		return errors.New("spaceID is empty")
	}
	if len(tp.Owners) == 0 {
		return errors.New("owners is empty")
	}

	return nil
}

// TestPlanV2CreateResponse testplan v2 create response
type TestPlanV2CreateResponse struct {
	Header
	Data uint64 `json:"data"`
}

// TestPlanV2UpdateRequest testplan v2 update request
type TestPlanV2UpdateRequest struct {
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	SpaceID    uint64   `json:"spaceID"`
	Owners     []string `json:"owners"`
	TestPlanID uint64   `json:"-"`

	IdentityInfo
}

// TestPlanV2UpdateResponse testplan v2 update response
type TestPlanV2UpdateResponse struct {
	Header
	Data uint64 `json:"data"`
}

// TestPlanV2PagingRequest testplan v2 query request
type TestPlanV2PagingRequest struct {
	Name      string   `schema:"name"`
	Owners    []string `schema:"owners"`
	Creator   string   `schema:"creator"`
	Updater   string   `schema:"updater"`
	SpaceID   uint64   `schema:"spaceID"`
	ProjectID uint64   `schema:"projectID"`

	// +optional default 1
	PageNo uint64 `schema:"pageNo"`
	// +optional default 20
	PageSize uint64 `schema:"pageSize"`
	// ids
	IDs []uint64

	IdentityInfo
}

func (tpr *TestPlanV2PagingRequest) UrlQueryString() map[string][]string {
	query := make(map[string][]string)
	query["pageNo"] = []string{strconv.FormatInt(int64(tpr.PageNo), 10)}
	query["pageSize"] = []string{strconv.FormatInt(int64(tpr.PageSize), 10)}
	query["projectID"] = []string{strconv.FormatInt(int64(tpr.ProjectID), 10)}
	if tpr.Name != "" {
		query["name"] = []string{tpr.Name}
	}
	if tpr.Creator != "" {
		query["creator"] = []string{tpr.Creator}
	}
	if tpr.Updater != "" {
		query["updater"] = []string{tpr.Updater}
	}
	if tpr.SpaceID != 0 {
		query["spaceID"] = []string{strconv.FormatInt(int64(tpr.SpaceID), 10)}
	}
	if len(tpr.Owners) != 0 {
		query["owners"] = tpr.Owners
	}

	return query
}

// TestPlanV2PagingResponse testplan query response
type TestPlanV2PagingResponse struct {
	Header
	UserInfoHeader
	Data TestPlanV2PagingResponseData `json:"data"`
}

// TestPlanV2GetResponse testplan get response
type TestPlanV2GetResponse struct {
	Header
	UserInfoHeader
	Data TestPlanV2 `json:"data"`
}

// TestPlanV2GetResponse testplan get response
type TestPlanV2StepGetResponse struct {
	Header
	Data TestPlanV2Step `json:"data"`
}

// TestPlanV2PagingResponseData testplan query response data
type TestPlanV2PagingResponseData struct {
	Total   int           `json:"total"`
	List    []*TestPlanV2 `json:"list"`
	UserIDs []string      `json:"userIDs,omitempty"`
}

// TestPlanV2Step step of test plan
type TestPlanV2Step struct {
	SceneSetID   uint64 `json:"sceneSetID"`
	SceneSetName string `json:"sceneSetName"`
	PreID        uint64 `json:"preID"`
	PlanID       uint64 `json:"planID"`
	ID           uint64 `json:"id"`
}

// TestPlanV2StepAddRequest Add a step in the test plan request
type TestPlanV2StepAddRequest struct {
	SceneSetID uint64 `json:"sceneSetID"`
	PreID      uint64 `json:"preID"`
	TestPlanID uint64 `json:"-"`

	IdentityInfo
}

type TestPlanV2StepAddResp struct {
	Header
	Data uint64 `json:"data"`
}

type TestPlanV2StepMoveResp struct {
	Header
	Data string `json:"data"`
}

// TestPlanV2StepDeleteRequest Delete a step in the test plan request
type TestPlanV2StepDeleteRequest struct {
	StepID     uint64 `json:"stepID"`
	TestPlanID uint64 `json:"-"`

	IdentityInfo
}

// TestPlanV2StepUpdateRequest Update a step in the test plan request
type TestPlanV2StepUpdateRequest struct {
	StepID      uint64 `json:"stepID"`
	PreID       uint64 `json:"preID"`
	ScenesSetId uint64 `json:"scenesSetId"`
	TestPlanID  uint64 `json:"-"`

	IdentityInfo
}

type TestPlanV2StepUpdateResp struct {
	Header
	Data string `json:"data"`
}
