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

package issue

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

type mockTranslator struct{}

func (m *mockTranslator) Get(lang i18n.LanguageCodes, key, def string) string { return key }
func (m *mockTranslator) Text(lang i18n.LanguageCodes, key string) string     { return key }
func (m *mockTranslator) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return key
}

func TestIssue_handleIssueStreamChangeIteration(t *testing.T) {
	// mock db to mock iteration
	db := &dao.DBClient{}
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIteration", func(client *dao.DBClient, id uint64) (*dao.Iteration, error) {
		return &dao.Iteration{BaseModel: dbengine.BaseModel{ID: id}, Title: strutil.String(id)}, nil
	})
	svc := &Issue{db: db, tran: &mockTranslator{}}

	// from unassigned to concrete iteration
	streamType, params, err := svc.handleIssueStreamChangeIteration(nil, apistructs.UnassignedIterationID, 1)
	assert.NoError(t, err)
	assert.Equal(t, apistructs.ISTChangeIterationFromUnassigned, streamType)
	assert.Equal(t, "1", params.NewIteration)

	// from concrete iteration to unassigned
	streamType, params, err = svc.handleIssueStreamChangeIteration(nil, 2, apistructs.UnassignedIterationID)
	assert.NoError(t, err)
	assert.Equal(t, apistructs.ISTChangeIterationToUnassigned, streamType)
	assert.Equal(t, "2", params.CurrentIteration)

	// from concrete to concrete iteration
	streamType, params, err = svc.handleIssueStreamChangeIteration(nil, 3, 4)
	assert.NoError(t, err)
	assert.Equal(t, apistructs.ISTChangeIteration, streamType)
	assert.Equal(t, "3", params.CurrentIteration)
	assert.Equal(t, "4", params.NewIteration)
}
