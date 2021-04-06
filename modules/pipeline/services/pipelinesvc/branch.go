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

package pipelinesvc

import (
	"github.com/erda-project/erda/modules/pipeline/commonutil/thirdparty/gittarutil"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/discover"
)

func (s *PipelineSvc) AllValidBranchWorkspaces(appID uint64) (map[string]string, error) {

	branchWorkspaces := make(map[string]string, 0)

	app, err := s.bdl.GetApp(appID)
	if err != nil {
		return nil, apierrors.ErrGetBranchWorkspaceMap.InternalError(err)
	}

	repo := gittarutil.NewRepo(discover.Gittar(), app.GitRepoAbbrev)

	branches, err := repo.Branches()
	if err != nil {
		return nil, apierrors.ErrGetGittarRepo.InternalError(err)
	}
	rules, err := s.bdl.GetProjectBranchRules(app.ProjectID)
	if err != nil {
		return nil, apierrors.ErrGetGittarRepo.InternalError(err)
	}
	for _, branch := range branches {
		ws, err := diceworkspace.GetByGitReference(branch, rules)
		if err != nil {
			continue
		}
		branchWorkspaces[branch] = ws.String()
	}

	tags, err := repo.Tags()
	if err != nil {
		return nil, apierrors.ErrGetGittarRepo.InternalError(err)
	}
	for _, tag := range tags {
		ws, err := diceworkspace.GetByGitReference(tag, rules)
		if err != nil {
			continue
		}
		branchWorkspaces[tag] = ws.String()
	}

	return branchWorkspaces, nil
}
