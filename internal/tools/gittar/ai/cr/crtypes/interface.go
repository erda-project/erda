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

package crtypes

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

type CodeReviewer interface {
	CodeReview() string
}

type FileCodeReviewer interface {
	CodeReviewer
	GetFileName() string
}

type CreateFunc func(req models.AICodeReviewNoteRequest, repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, user *models.User) (CodeReviewer, error)

var Factory = map[models.AICodeReviewType]CreateFunc{}

func Register(t models.AICodeReviewType, f CreateFunc) {
	Factory[t] = f
}
