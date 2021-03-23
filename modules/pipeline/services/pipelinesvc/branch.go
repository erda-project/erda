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
