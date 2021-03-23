package gittarutil

type Repo struct {
	GittarAddr string
	Repo       string
}

func NewRepo(gittarAddr, repo string) *Repo {
	return &Repo{GittarAddr: gittarAddr, Repo: repo}
}
