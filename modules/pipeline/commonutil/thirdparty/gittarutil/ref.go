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
