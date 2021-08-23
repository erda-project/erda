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

package addon

import (
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

const count = 20

func TestConcurrentReadWriteAppInfos(t *testing.T) {
	var keys = []string{"1", "2", "3", "4", "5"}

	var db *dbclient.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetAttachmentsByInstanceID",
		func(*dbclient.DBClient, string) (*[]dbclient.AddonAttachment, error) {
			var addons []dbclient.AddonAttachment
			for _, v := range keys {
				addons = append(addons, dbclient.AddonAttachment{
					ProjectID:     v,
					ApplicationID: v,
				})
			}
			return &addons, nil
		},
	)
	defer monkey.UnpatchAll()

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetAppsByProject",
		func(_ *bundle.Bundle, id uint64, _ uint64, _ string) (*apistructs.ApplicationListResponseData, error) {
			return &apistructs.ApplicationListResponseData{
				List: []apistructs.ApplicationDTO{
					{
						ID: id,
					},
				},
			}, nil
		},
	)

	var (
		wg         sync.WaitGroup
		orgID      uint64 = 1
		userID            = "1"
		instanceID        = "1"
	)
	wg.Add(count)
	for i := 0; i != count; i++ {
		go func() {
			a := Addon{}
			_, err := a.ListReferencesByInstanceID(orgID, userID, instanceID)
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for _, v := range keys {
		_, ok := AppInfos.Load(v)
		assert.Equal(t, true, ok)
	}
}

func TestDeleteAddonUsed(t *testing.T) {
	var db *dbclient.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetInstanceRouting",
		func(*dbclient.DBClient, string) (*dbclient.AddonInstanceRouting, error) {
			return &dbclient.AddonInstanceRouting{}, nil
		},
	)

	addon := Addon{}
	monkey.PatchInstanceMethod(reflect.TypeOf(&addon), "DeleteTenant",
		func(*Addon, string, string) error {
			return nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetAttachmentCountByRoutingInstanceID",
		func(*dbclient.DBClient, string) (int64, error) {
			return 1, nil
		},
	)
	defer monkey.UnpatchAll()

	err := addon.Delete("", "")
	if err.Error() != "addon is being referenced, can't delete" {
		t.Fatal("the err is not equal with expected")
	}
}
