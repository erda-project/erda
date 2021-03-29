package gittarutil

import (
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpclientutil"

	"github.com/pkg/errors"
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
