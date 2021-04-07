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

package models

import "github.com/erda-project/erda/modules/gittar/pkg/gitmodule"

// PayloadPushEvent struct
// https://docs.gitlab.com/ce/user/project/integrations/webhooks.html#push-events

type PayloadPushEvent struct {
	ObjectKind        string                `json:"object_kind"`
	IsTag             bool                  `json:"is_tag"`
	Ref               string                `json:"ref"`
	After             string                `json:"after"`
	Before            string                `json:"before"`
	IsDelete          bool                  `json:"is_delete"`
	Commits           []PayloadCommit       `json:"commits"`
	TotalCommitsCount int                   `json:"total_commits_count"`
	Pusher            *User                 `json:"pusher"`
	Repository        *gitmodule.Repository `json:"repository"`
}
