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
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpclientutil"
)

type Ref struct {
	Name string `json:"name"`
}

func (r *Repo) Branches() ([]string, error) {
	var refs []Ref
	req := httpclient.New().Get(r.GittarAddr).Path("/" + r.Repo + "/branches")
	if err := httpclientutil.DoJson(req, &refs); err != nil {
		return nil, err
	}
	var branches []string
	for _, ref := range refs {
		branches = append(branches, ref.Name)
	}
	return branches, nil
}

func (r *Repo) Tags() ([]string, error) {
	var refs []Ref
	req := httpclient.New().Get(r.GittarAddr).Path("/" + r.Repo + "/tags")
	if err := httpclientutil.DoJson(req, &refs); err != nil {
		return nil, err
	}
	var tags []string
	for _, ref := range refs {
		tags = append(tags, ref.Name)
	}
	return tags, nil
}
