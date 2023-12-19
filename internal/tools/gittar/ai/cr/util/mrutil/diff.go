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

package mrutil

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

func GetDiffFileFromMR(repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, fileName string) *gitmodule.DiffFile {
	diff := GetDiffFromMR(repo, mr)
	if diff == nil {
		return nil
	}
	for _, diffFile := range diff.Files {
		if diffFile.Name == fileName {
			return diffFile
		}
	}
	return nil
}

// GetDiffFromMR ref to: internal/tools/gittar/api/repo.go:411#Compare
func GetDiffFromMR(repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo) *gitmodule.Diff {
	fromCommit, err := repo.GetBranchCommit(mr.SourceBranch)
	if err != nil {
		logrus.Errorf("failed to get commit, sha: %s, err: %s", mr.SourceSha, err)
		return nil
	}
	toCommit, err := repo.GetBranchCommit(mr.TargetBranch)
	if err != nil {
		logrus.Errorf("failed to get commit, sha: %s, err: %s", mr.TargetSha, err)
		return nil
	}
	baseCommit, err := repo.GetMergeBase(fromCommit, toCommit)
	if err != nil {
		logrus.Errorf("failed to get merge base, from: %s, to: %s, err: %s", mr.SourceSha, mr.TargetSha, err)
		return nil
	}
	diff, err := repo.GetDiff(fromCommit, baseCommit)
	if err != nil {
		logrus.Errorf("failed to get diff, from: %s, to: %s, err: %s", mr.SourceSha, mr.TargetSha, err)
		return nil
	}
	return diff
}
