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

// +build !codeanalysis

package gitmodule

import (
	"strings"

	git "github.com/libgit2/git2go/v30"
)

const BRANCH_PREFIX = "refs/heads/"

// IsReferenceExist returns true if given reference exists in the repository.
func IsReferenceExist(repoPath, name string) bool {
	_, err := NewCommand("show-ref", "--verify", name).RunInDir(repoPath)
	return err == nil
}

// IsBranchExist returns true if given branch exists in the repository.
func IsBranchExist(repoPath, name string) bool {
	return IsReferenceExist(repoPath, BRANCH_PREFIX+name)
}

func (repo *Repository) IsBranchExist(name string) bool {
	return IsBranchExist(repo.DiskPath(), name)
}

// Branch represents a Git branch.
type Branch struct {
	Id        string  `json:"id"`
	Name      string  `json:"name"`
	Commit    *Commit `json:"commit"`
	IsDefault bool    `json:"isDefault"`
	IsProtect bool    `json:"isProtect"`
	IsMerged  bool    `json:"isMerged"` // 是否已合并到默认分支
}

type Branches []*Branch

func (branches Branches) Len() int {
	return len(branches)
}
func (branches Branches) Less(i, j int) (less bool) {
	defer func() {
		if r := recover(); r != nil {
			less = branches[i].Name < branches[j].Name
		}
	}()
	return !branches[i].Commit.Committer.When.Before(branches[j].Commit.Committer.When)
}

func (branches Branches) Swap(i, j int) {
	branches[i], branches[j] = branches[j], branches[i]
}

// GetBranches 获取分支名列表
func (repo *Repository) GetBranches() ([]string, error) {
	rawrepo, err := repo.GetRawRepo()
	if err != nil {
		return nil, err
	}
	iter, _ := rawrepo.NewBranchIterator(git.BranchLocal)
	branches := []string{}
	err = iter.ForEach(func(b *git.Branch, t git.BranchType) error {
		branch, err := b.Name()
		if err != nil {
			return err
		}

		branches = append(branches, branch)
		return nil
	})
	return branches, err
}

// GetDetailBranches 获取分支详情列表
func (repo *Repository) GetDetailBranches(onlyBranchNames bool, findBranch string) ([]*Branch, error) {
	rawrepo, err := repo.GetRawRepo()
	if err != nil {
		return nil, err
	}
	branchNum, err := repo.GetBranches()
	if len(branchNum) <= 0 {
		return []*Branch{}, nil
	}

	if onlyBranchNames {
		var branches []*Branch
		for _, branchName := range branchNum {
			if !strings.Contains(branchName, findBranch) {
				continue
			}
			branches = append(branches, &Branch{Name: branchName})
		}
		return branches, nil
	}

	defaultBranch, err := repo.GetDefaultBranch()
	if err != nil {
		return nil, err
	}
	mergedBranchMap, err := repo.GetMergedBranches(defaultBranch)
	if err != nil {
		return nil, err
	}

	iter, _ := rawrepo.NewBranchIterator(git.BranchLocal)
	branches := []*Branch{}
	rules, _ := repo.Bundle.GetAppBranchRules(uint64(repo.ApplicationId))
	err = iter.ForEach(func(b *git.Branch, t git.BranchType) error {
		branchName, err := b.Name()
		if err != nil {
			return err
		}
		if !strings.Contains(branchName, findBranch) {
			return nil
		}
		commit, err := repo.GetCommit(b.Target().String())
		if err != nil {
			return err
		}
		branch := &Branch{
			Id:        commit.ID,
			Name:      branchName,
			Commit:    commit,
			IsDefault: defaultBranch == branchName,
			IsProtect: repo.IsProtectBranchWithRules(branchName, rules),
			IsMerged:  mergedBranchMap[branchName] == 1,
		}

		branches = append(branches, branch)
		return nil
	})
	return branches, err
}

// DeleteBranch deletes a branch from repository.
func (repo *Repository) DeleteBranch(name string) error {
	rawRepo, err := repo.GetRawRepo()
	if err != nil {
		return err
	}
	branch, err := rawRepo.LookupBranch(name, git.BranchLocal)
	if err != nil {
		return err
	}

	return branch.Delete()
}

// CreateBranch create a branch from repository.
func (repo *Repository) CreateBranch(name string, ref string) error {
	commit, err := repo.GetCommitByAny(ref)
	if err != nil {
		return err
	}
	rawrepo, err := repo.GetRawRepo()
	if err != nil {
		return err
	}
	rawCommit, err := rawrepo.LookupCommit(commit.Git2Oid())
	if err != nil {
		return err
	}
	//不覆盖已有同名分支
	_, err = rawrepo.CreateBranch(name, rawCommit, false)
	return err
}

// AddRemote adds a new remote to repository.
func (repo *Repository) AddRemote(name, url string, fetch bool) error {
	cmd := NewCommand("remote", "add")
	if fetch {
		cmd.AddArguments("-f")
	}
	cmd.AddArguments(name, url)

	_, err := cmd.RunInDir(repo.DiskPath())
	return err
}

// RemoveRemote removes a remote from repository.
func (repo *Repository) RemoveRemote(name string) error {
	_, err := NewCommand("remote", "remove", name).RunInDir(repo.DiskPath())
	return err
}
