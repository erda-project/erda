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

package cr

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/pkg/limit_sync_group"
	"github.com/erda-project/erda/pkg/strutil"
)

type mrReviewer struct {
	repository *gitmodule.Repository
	mr         *apistructs.MergeRequestInfo
}

func newMRReviewer(repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo) CodeReviewer {
	return &mrReviewer{repository: repo, mr: mr}
}

func (r *mrReviewer) CodeReview() string {
	diff := getDiffFromMR(r.repository, r.mr)

	// mr has many changed files, we will review only the first ten files one by one. Then, combine the file-level suggestions.

	var changedFiles []FileCodeReviewer
	for _, diffFile := range diff.Files {
		if len(changedFiles) >= 10 {
			break
		}
		if len(diffFile.Sections) > 0 {
			fr := newFileReviewer(diffFile, convertToModelUser(r.mr.AuthorUser), r.mr)
			changedFiles = append(changedFiles, fr)
		}
	}

	// parallel do file-level cr
	var fileOrder []string
	fileSuggestions := make(map[string]string)
	wg := limit_sync_group.NewSemaphore(5) // parallel is 5
	var mu sync.Mutex

	wg.Add(len(changedFiles))
	for _, file := range changedFiles {
		fileOrder = append(fileOrder, file.GetFileName())
		go func(file FileCodeReviewer) {
			defer wg.Done()

			mu.Lock()
			fileSuggestions[file.GetFileName()] = file.CodeReview()
			mu.Unlock()
		}(file)
	}
	wg.Wait()

	// combine result
	var mrReviewResult string
	mrReviewResult = "# AI Code Review\n"
	for _, fileName := range fileOrder {
		mrReviewResult += "------\n## File: `" + fileName + "`\n\n" + fileSuggestions[fileName] + "\n"
	}

	return mrReviewResult
}

func getDiffFromMR(repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo) *gitmodule.Diff {
	fromCommit, err := repo.GetCommit(mr.SourceSha)
	if err != nil {
		logrus.Errorf("failed to get commit, sha: %s, err: %s", mr.SourceSha, err)
		return nil
	}
	toCommit, err := repo.GetCommit(mr.TargetSha)
	if err != nil {
		logrus.Errorf("failed to get commit, sha: %s, err: %s", mr.TargetSha, err)
		return nil
	}
	diff, err := repo.GetDiff(fromCommit, toCommit)
	if err != nil {
		logrus.Errorf("failed to get diff, from: %s, to: %s, err: %s", mr.SourceSha, mr.TargetSha, err)
		return nil
	}
	return diff
}

func getDiffFileFromMR(repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, fileName string) *gitmodule.DiffFile {
	diff := getDiffFromMR(repo, mr)
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

func convertToModelUser(userInfo *apistructs.UserInfoDto) *models.User {
	return &models.User{
		Id:       strutil.String(userInfo.UserID),
		Name:     userInfo.Username,
		NickName: userInfo.NickName,
		Email:    userInfo.Email,
	}
}
