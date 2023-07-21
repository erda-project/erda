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

package endpoints

import (
	"context"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/gorilla/schema"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/apps/dop/services/iteration"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

func TestPagingIterations(t *testing.T) {
	pm1 := monkey.Patch(user.GetIdentityInfo, func(r *http.Request) (apistructs.IdentityInfo, error) {
		return apistructs.IdentityInfo{UserID: "1", InternalClient: "bundle"}, nil
	})
	defer pm1.Unpatch()

	iterationSvc := &iteration.Iteration{}
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(iterationSvc), "Paging", func(itr *iteration.Iteration, req apistructs.IterationPagingRequest) ([]dao.Iteration, uint64, error) {
		return []dao.Iteration{{BaseModel: dbengine.BaseModel{ID: 1}, ProjectID: 1}}, 1, nil
	})
	defer pm2.Unpatch()

	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(iterationSvc), "SetIssueSummaries", func(itr *iteration.Iteration, projectID uint64, iterationMap map[int64]*apistructs.Iteration) error {
		return nil
	})
	defer pm3.Unpatch()

	ep := Endpoints{iteration: iterationSvc, queryStringDecoder: schema.NewDecoder()}
	r := &http.Request{Header: http.Header{}, URL: &url.URL{}}
	r.Header.Set("Org-ID", "1")
	q := r.URL.Query()
	q.Add("projectID", "1")
	q.Add("labels", "area")
	r.URL.RawQuery = q.Encode()
	_, err := ep.PagingIterations(context.Background(), r, map[string]string{"projectID": "1"})
	assert.NoError(t, err)
}

func Test_getIterationLabelDetails(t *testing.T) {
	bdl := bundle.New()
	db := &dao.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetLabelRelationsByRef", func(db *dao.DBClient, refType apistructs.ProjectLabelType, refID string) ([]dao.LabelRelation, error) {
		return []dao.LabelRelation{
			{
				LabelID: 1,
			},
		}, nil
	})
	defer pm1.Unpatch()

	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListLabelByIDs", func(bdl *bundle.Bundle, ids []uint64) ([]apistructs.ProjectLabel, error) {
		return []apistructs.ProjectLabel{
			{
				ID:   1,
				Name: "area",
			},
		}, nil
	})
	defer pm2.Unpatch()

	e := &Endpoints{bdl: bdl, db: db}
	labels, labelDetails := e.getIterationLabelDetails(1)
	assert.Equal(t, []string{"area"}, labels)
	assert.Equal(t, int64(1), labelDetails[0].ID)
}
