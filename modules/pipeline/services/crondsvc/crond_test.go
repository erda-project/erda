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

package crondsvc

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/jsonstore"
)

// Result:
// 1000  个约 0.03s
// 10000 个约 1.7s
func TestReloadSpeed(t *testing.T) {
	d := cron.New()
	d.Start()
	for i := 0; i < 10; i++ {
		if err := d.AddFunc("*/1 * * * *", func() {
			fmt.Println("hello world")
		}); err != nil {
			panic(err)
		}
	}
	time.Sleep(time.Second * 2)
}

func TestMakePipelineCronName(t *testing.T) {
	cronID := uint64(123)
	cronName := makePipelineCronName(cronID)
	assert.Equal(t, "pipeline-cron[123]", cronName)
}

func TestMakeCleanBuildCacheJobName(t *testing.T) {
	expr := "golang"
	jobName := makeCleanBuildCacheJobName(expr)
	assert.Equal(t, "clean-build-cache-image-[golang]", jobName)
}

func TestCrondSnapshot(t *testing.T) {
	s := &CrondSvc{
		crond: &cron.Cron{},
	}
	cronSnapthots := s.crondSnapshot()
	assert.Equal(t, 2, len(cronSnapthots))
}

func TestParseCronIDFromWatchedKey(t *testing.T) {
	deleteKey := "/devops/pipeline/cron/delete-123"
	deleteCronID, err := parseCronIDFromWatchedKey(deleteKey)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(123), deleteCronID)

	addKey := "/devops/pipeline/cron/add-456"
	addCronID, err := parseCronIDFromWatchedKey(addKey)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(456), addCronID)
}

func TestDeletePipelineCrond(t *testing.T) {
	js := &jsonstore.JsonStoreImpl{}
	s := &CrondSvc{
		crond: &cron.Cron{},
		js:    js,
	}
	pm := monkey.PatchInstanceMethod(reflect.TypeOf(js), "Put", func(j *jsonstore.JsonStoreImpl, ctx context.Context, key string, object interface{}) error {
		return nil
	})
	defer pm.Unpatch()
	err := s.DeletePipelineCrond(10)
	assert.Equal(t, nil, err)
}

func TestAddIntoPipelineCrond(t *testing.T) {
	js := &jsonstore.JsonStoreImpl{}
	s := &CrondSvc{
		crond: &cron.Cron{},
		js:    js,
	}
	pm := monkey.PatchInstanceMethod(reflect.TypeOf(js), "Put", func(j *jsonstore.JsonStoreImpl, ctx context.Context, key string, object interface{}) error {
		return nil
	})
	defer pm.Unpatch()
	err := s.AddIntoPipelineCrond(10)
	assert.Equal(t, nil, err)
}

func TestListenCrond(t *testing.T) {
	s := &CrondSvc{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "ListenCrond", func(s *CrondSvc, ctx context.Context, pipelineCronFunc func(id uint64)) {
		return
	})
	defer pm1.Unpatch()
	t.Run("ListenCrond", func(t *testing.T) {
		s.ListenCrond(context.Background(), func(id uint64) {})
	})
}
