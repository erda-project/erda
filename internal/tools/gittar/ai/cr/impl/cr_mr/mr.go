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

package cr_mr

import (
	"fmt"
	"strings"
	"sync"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/impl/cr_mr_file"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/mrutil"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/pkg/limit_sync_group"
)

const (
	MaxDiffFileNum = 5
)

type mrReviewer struct {
	req  models.AICodeReviewNoteRequest
	repo *gitmodule.Repository
	user *models.User
	mr   *apistructs.MergeRequestInfo
}

func init() {
	models.RegisterCodeReviewer(models.AICodeReviewTypeMR, func(req models.AICodeReviewNoteRequest, repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, user *models.User) (models.CodeReviewer, error) {
		return &mrReviewer{req: req, repo: repo, user: user, mr: mr}, nil
	})
}

func (r *mrReviewer) CodeReview(i18n i18n.Translator, lang i18n.LanguageCodes) string {
	diff := mrutil.GetDiffFromMR(r.repo, r.mr)

	// mr has many changed files, we will review only the first ten files one by one. Then, combine the file-level suggestions.

	var changedFiles []models.FileCodeReviewer
	var reachMaxDiffFileNumLimit bool
	for _, diffFile := range diff.Files {
		if len(changedFiles) >= MaxDiffFileNum {
			reachMaxDiffFileNumLimit = true
			break
		}
		if len(diffFile.Sections) > 0 {
			fr := cr_mr_file.NewFileReviewer(diffFile, r.user, r.mr)
			changedFiles = append(changedFiles, fr)
		}
	}

	// parallel do file-level cr
	var fileOrder []string
	fileSuggestions := make(map[string]string)
	wg := limit_sync_group.NewSemaphore(MaxDiffFileNum) // parallel is 5
	var mu sync.Mutex

	wg.Add(len(changedFiles))
	for _, file := range changedFiles {
		fileOrder = append(fileOrder, file.GetFileName())
		go func(file models.FileCodeReviewer) {
			defer wg.Done()

			mu.Lock()
			fileSuggestion := file.CodeReview(i18n, lang)
			if strings.TrimSpace(fileSuggestion) == "" {
				fileSuggestion = i18n.Text(lang, models.I18nKeyMrAICrNoSuggestion)
			}
			fileSuggestions[file.GetFileName()] = fileSuggestion
			mu.Unlock()
		}(file)
	}
	wg.Wait()

	// combine result
	var mrReviewResults []string
	mrReviewResults = append(mrReviewResults, fmt.Sprintf("# %s", i18n.Text(lang, models.I18nKeyMrAICrTitle)))
	if reachMaxDiffFileNumLimit {
		mrReviewResults = append(mrReviewResults, fmt.Sprintf(i18n.Text(lang, models.I18nKeyTemplateMrAICrTipForEachMaxLimit), MaxDiffFileNum))
	}
	mrReviewResults = append(mrReviewResults, "")
	for _, fileName := range fileOrder {
		mrReviewResults = append(mrReviewResults, "------")
		mrReviewResults = append(mrReviewResults, fmt.Sprintf("## %s: `%s`", i18n.Text(lang, models.I18nKeyFile), fileName))
		mrReviewResults = append(mrReviewResults, "")
		mrReviewResults = append(mrReviewResults, fileSuggestions[fileName])
	}

	return strings.Join(mrReviewResults, "\n")
}
