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

package gittarutil

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpclientutil"
)

type Commit struct {
	ID        string `json:"id"`
	Committer struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		When  string `json:"when"`
	} `json:"committer"`
	CommitMessage string `json:"commitMessage"`
}

func (r *Repo) GetCommit(ref string) (Commit, error) {
	var commit []Commit
	req := httpclient.New().Get(r.GittarAddr).
		Path("/"+r.Repo+"/commits/"+ref).
		Param("pageNo", "1").Param("pageSize", "1")

	if err := httpclientutil.DoJson(req, &commit); err != nil {
		return Commit{}, err
	}
	if len(commit) == 0 {
		return Commit{}, errors.New("no commit found")
	}
	return commit[0], nil
}
