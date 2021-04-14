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

	git "github.com/libgit2/git2go/v28"
)

const TAG_PREFIX = "refs/tags/"

// IsTagExist returns true if given tag exists in the repository.
func IsTagExist(repoPath, name string) bool {
	return IsReferenceExist(repoPath, TAG_PREFIX+name)
}

func (repo *Repository) IsTagExist(name string) bool {
	return IsTagExist(repo.DiskPath(), name)
}

// CreateTag 创建附注tag
func (repo *Repository) CreateTag(name, revision string, signature *Signature, message string) error {
	commit, err := repo.GetCommitByAny(revision)
	if err != nil {
		return err
	}
	rawRepo, err := repo.GetRawRepo()
	if err != nil {
		return err
	}

	tagger := &git.Signature{
		Name:  signature.Name,
		Email: signature.Email,
		When:  signature.When,
	}

	//TODO 直接用commit.rawCommit会报错the given target does not belong to this repository, CreateBranch一样的方式不会
	oid, err := git.NewOid(commit.ID)
	if err != nil {
		return err
	}
	rawCommit, err := rawRepo.LookupCommit(oid)
	if err != nil {
		return err
	}
	_, err = rawRepo.Tags.Create(name, rawCommit, tagger, message)
	return err
}

func (repo *Repository) getTag(id string) (*Tag, error) {
	// Get tag type
	tp, err := NewCommand("cat-file", "-t", id).RunInDir(repo.DiskPath())
	if err != nil {
		return nil, err
	}
	tp = strings.TrimSpace(tp)

	// Tag is a commit.
	if ObjectType(tp) == OBJECT_COMMIT {
		tag := &Tag{
			ID:     id,
			Object: id,
			Type:   string(OBJECT_COMMIT),
			repo:   repo,
		}

		return tag, nil
	}

	// Tag with message.
	data, err := NewCommand("cat-file", "-p", id).RunInDirBytes(repo.DiskPath())
	if err != nil {
		return nil, err
	}

	tag, err := parseTagData(data)
	if err != nil {
		return nil, err
	}

	tag.ID = id
	tag.repo = repo

	return tag, nil
}

// GetTag returns a Git tag by given name.
func (repo *Repository) GetTag(name string) (*Tag, error) {
	stdout, err := NewCommand("show-ref", "--tags", name).RunInDir(repo.DiskPath())
	if err != nil {
		return nil, err
	}

	id := strings.Split(stdout, " ")[0]
	if err != nil {
		return nil, err
	}

	tag, err := repo.getTag(id)
	if err != nil {
		return nil, err
	}
	tag.Name = name
	return tag, nil
}

// GetTags returns all tags of the repository.
func (repo *Repository) GetTags() ([]string, error) {
	rawrepo, err := repo.GetRawRepo()
	if err != nil {
		return nil, err
	}

	tags := []string{}
	err = rawrepo.Tags.Foreach(func(name string, id *git.Oid) error {
		name = strings.TrimPrefix(name, TAG_PREFIX)
		tags = append(tags, name)
		return nil
	})
	return tags, err
}

func (repo *Repository) GetDetailTags(findTags string) ([]*Tag, error) {
	rawrepo, err := repo.GetRawRepo()
	if err != nil {
		return nil, err
	}
	tags := []*Tag{}
	err = rawrepo.Tags.Foreach(func(name string, id *git.Oid) error {
		tagName := strings.TrimPrefix(name, TAG_PREFIX)
		if !strings.Contains(tagName, findTags) {
			return nil
		}
		ref, err := rawrepo.References.Lookup(name)

		object, err := rawrepo.Lookup(ref.Target())
		if err != nil {
			return err
		}

		rawCommit, err := object.AsCommit()
		if err != nil {
			//不是轻量级tag  尝试 附注tag
			rawTag, err := object.AsTag()
			if err != nil {
				return err
			} else {
				rawTagger := rawTag.Tagger()
				tag := &Tag{
					ID:     rawTag.TargetId().String(),
					Object: rawTag.Id().String(),
					Name:   tagName,
					Tagger: &Signature{
						Email: rawTagger.Email,
						Name:  rawTagger.Name,
						When:  rawTagger.When,
					},
					Message: rawTag.Message(),
				}
				tags = append(tags, tag)
			}
		} else {
			//轻量级tag, tag就是commit本身别名
			commit := NewCommitFromLibgit2(repo, rawCommit)
			tag := &Tag{
				ID:      id.String(),
				Object:  id.String(),
				Name:    tagName,
				Tagger:  commit.Committer,
				Message: commit.Message(),
			}
			tags = append(tags, tag)
		}

		return nil
	})
	return tags, err
}

// DeleteTag deletes a tag from the repository
func (repo *Repository) DeleteTag(name string) error {
	rawRepo, err := repo.GetRawRepo()
	if err != nil {
		return err
	}
	err = rawRepo.Tags.Remove(name)
	return err
}
