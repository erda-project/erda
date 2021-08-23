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
