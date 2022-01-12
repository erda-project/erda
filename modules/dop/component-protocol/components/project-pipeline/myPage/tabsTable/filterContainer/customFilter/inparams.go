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

package customFilter

import (
	"encoding/json"
	"strconv"
)

type InParams struct {
	OrgID uint64 `json:"orgID,omitempty"`

	FrontendProjectID string `json:"projectId,omitempty"`
	FrontendAppID     string `json:"appId,omitempty"`
	FrontendUrlQuery  string `json:"issueFilter__urlQuery,omitempty"`

	ProjectID uint64 `json:"-"`
	AppID     uint64 `json:"-"`
}

func (p *CustomFilter) setInParams() error {
	b, err := json.Marshal(p.InParamsPtr())
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &p.InParams); err != nil {
		return err
	}

	p.InParams.OrgID, err = strconv.ParseUint(p.sdk.Identity.OrgID, 10, 64)
	if err != nil {
		return err
	}

	if p.InParams.FrontendProjectID != "" {
		p.InParams.ProjectID, err = strconv.ParseUint(p.InParams.FrontendProjectID, 10, 64)
		if err != nil {
			return err
		}
	}
	if p.InParams.FrontendAppID != "" {
		p.InParams.AppID, err = strconv.ParseUint(p.InParams.FrontendAppID, 10, 64)
		if err != nil {
			return err
		}
	}
	return nil
}
