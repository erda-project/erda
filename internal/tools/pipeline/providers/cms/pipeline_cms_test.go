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

package cms

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cms/db"
)

func TestBatchGetConfigs(t *testing.T) {
	dbClient := &db.Client{}
	nsMap := map[uint64]string{
		0: "ns-0",
		1: "ns-1",
		2: "ns-2",
	}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "BatchGetCmsNamespaces", func(_ *db.Client, pipelineSource string, namespaces []string, ops ...mysqlxorm.SessionOption) ([]db.PipelineCmsNs, error) {
		res := make([]db.PipelineCmsNs, 0)
		for idx, ns := range namespaces {
			res = append(res, db.PipelineCmsNs{
				ID:             uint64(idx),
				PipelineSource: apistructs.PipelineSource(pipelineSource),
				Ns:             ns,
			})
		}
		return res, nil
	})
	defer pm1.Unpatch()
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "BatchGetCmsNsConfigs", func(_ *db.Client, cmsNsIDs []uint64, ops ...mysqlxorm.SessionOption) ([]db.PipelineCmsConfig, error) {
		res := make([]db.PipelineCmsConfig, 0)
		for idx, nsID := range cmsNsIDs {
			res = append(res, db.PipelineCmsConfig{
				ID:      uint64(idx),
				NsID:    nsID,
				Key:     fmt.Sprintf("key-%s", nsMap[nsID]),
				Value:   fmt.Sprintf("value-%s", nsMap[nsID]),
				Encrypt: &[]bool{false}[0],
			})
		}
		return res, nil
	})
	defer pm2.Unpatch()
	cm := pipelineCm{
		dbClient: dbClient,
	}
	configs, err := cm.BatchGetConfigs(context.Background(), &pb.CmsNsConfigsBatchGetRequest{
		PipelineSource: "dice",
		Namespaces: []string{
			"ns-0",
			"ns-1",
			"ns-2",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(configs))
	// check order
	assert.Equal(t, "key-ns-0", configs[0].Key)
	assert.Equal(t, "value-ns-2", configs[2].Value)
}
