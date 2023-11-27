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
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/crtypes"
	_ "github.com/erda-project/erda/internal/tools/gittar/ai/cr/impl"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

func NewCodeReviewer(req models.AICodeReviewNoteRequest, repo *gitmodule.Repository, user *models.User, mr *apistructs.MergeRequestInfo) (cr crtypes.CodeReviewer, err error) {
	f, ok := crtypes.Factory[req.Type]
	if !ok {
		return nil, fmt.Errorf("unknown code review type: %s", req.Type)
	}
	return f(req, repo, mr, user)
}
