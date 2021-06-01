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

package endpoints

import (
	"math/rand"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/addon"
)

const count = 20

func TestConcurrentWriteAddonInfos(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e"}

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "QueryExtensions",
		func(*bundle.Bundle, apistructs.ExtensionQueryRequest) ([]apistructs.Extension, error) {
			var ext []apistructs.Extension
			for i, v := range keys {
				ext = append(ext, apistructs.Extension{ID: uint64(i), Name: v})
			}
			return ext, nil
		},
	)
	defer monkey.UnpatchAll()
	var (
		wg sync.WaitGroup
	)
	wg.Add(count)
	for i := 0; i != count; i++ {
		go func() {
			e := Endpoints{
			}
			_, err := e.SyncAddons()
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for _, v := range keys {
		_, ok := addon.AddonInfos.Load(v)
		assert.Equal(t, true, ok)
	}
}

func TestConcurrentReadWriteAddonInfos(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e"}

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "QueryExtensions",
		func(*bundle.Bundle, apistructs.ExtensionQueryRequest) ([]apistructs.Extension, error) {
			var ext []apistructs.Extension
			for i, v := range keys {
				ext = append(ext, apistructs.Extension{ID: uint64(i), Name: v})
			}
			return ext, nil
		},
	)
	defer monkey.UnpatchAll()
	var (
		wg sync.WaitGroup
	)
	wg.Add(count * 2)
	for i := 0; i != count; i++ {
		go func() {
			e := Endpoints{
			}
			_, err := e.SyncAddons()
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	for i := 0; i != count; i++ {
		go func() {
			addon.AddonInfos.Load(rand.Intn(len(keys)))
			wg.Done()
		}()
	}

	wg.Wait()
	for _, v := range keys {
		_, ok := addon.AddonInfos.Load(v)
		assert.Equal(t, true, ok)
	}
}

func TestConcurrentWriteProjectInfos(t *testing.T) {
	keys := []string{"1", "2", "3", "4", "5"}

	var db *dbclient.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetDistinctProjectInfo",
		func(*dbclient.DBClient) (*[]string, error) {
			return &keys, nil
		},
	)
	defer monkey.UnpatchAll()

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProject",
		func(_ *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{
				ID: id,
			}, nil
		},
	)

	var (
		wg sync.WaitGroup
	)
	wg.Add(count)
	for i := 0; i != count; i++ {
		go func() {
			e := Endpoints{
			}
			_, err := e.SyncProjects()
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for _, v := range keys {
		_, ok := addon.ProjectInfos.Load(v)
		assert.Equal(t, true, ok)
	}
}

func TestConcurrentReadWriteProjectInfos(t *testing.T) {
	keys := []string{"1", "2", "3", "4", "5"}

	var db *dbclient.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetDistinctProjectInfo",
		func(*dbclient.DBClient) (*[]string, error) {
			return &keys, nil
		},
	)
	defer monkey.UnpatchAll()

	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProject",
		func(_ *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{
				ID: id,
			}, nil
		},
	)

	var (
		wg sync.WaitGroup
	)
	wg.Add(count * 2)
	for i := 0; i != count; i++ {
		go func() {
			e := Endpoints{
			}
			_, err := e.SyncProjects()
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	for i := 0; i != count; i++ {
		go func() {
			addon.ProjectInfos.Load(rand.Intn(len(keys)))
			wg.Done()
		}()
	}
	wg.Wait()
	for _, v := range keys {
		_, ok := addon.ProjectInfos.Load(v)
		assert.Equal(t, true, ok)
	}
}
