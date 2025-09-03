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

func TestBatchGetConfigs_RealDataWithMultipleKeys(t *testing.T) {
	// test constants
	const (
		nsUser    = "user-1-org-1"
		nsDevelop = "pipeline-secrets-app-2-develop"
		nsDefault = "pipeline-secrets-app-2-default"

		keyBuildProfile = "BUILD_PROFILE"
		keyUserToken    = "USER_TOKEN"
		keyOrgName      = "ORG_NAME"
		keyEnvType      = "ENV_TYPE"
		keyApiHost      = "API_HOST"
		keyDatabaseURL  = "DATABASE_URL"
		keyRedisURL     = "REDIS_URL"

		valueBuildProfileDev  = "dev"
		valueBuildProfileTest = "test"
		valueUserToken        = "abc123"
		valueOrgName          = "terminus"
		valueEnvTypeDevelop   = "develop"
		valueEnvTypeDefault   = "default"
		valueApiHost          = "dev.api.com"
		valueDatabaseURL      = "mysql://default.db"
		valueRedisURL         = "redis://default.cache"
	)

	dbClient := &db.Client{}

	// real namespace data based on database screenshot
	realNsData := map[uint64]string{
		1: nsUser,
		2: nsDevelop,
		3: nsDefault,
	}

	// mock BatchGetCmsNamespaces to return in database order (by ID), not request order
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "BatchGetCmsNamespaces", func(_ *db.Client, pipelineSource string, namespaces []string, ops ...mysqlxorm.SessionOption) ([]db.PipelineCmsNs, error) {
		// return in database order [1,2,3], not request order [3,2,1]
		res := make([]db.PipelineCmsNs, 0)
		for id := uint64(1); id <= 3; id++ {
			if ns, exists := realNsData[id]; exists {
				// check if this namespace is in the request
				for _, reqNs := range namespaces {
					if reqNs == ns {
						res = append(res, db.PipelineCmsNs{
							ID:             id,
							PipelineSource: apistructs.PipelineSource(pipelineSource),
							Ns:             ns,
						})
						break
					}
				}
			}
		}
		return res, nil
	})
	defer pm1.Unpatch()

	// mock BatchGetCmsNsConfigs to return configs in mixed order
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "BatchGetCmsNsConfigs", func(_ *db.Client, cmsNsIDs []uint64, ops ...mysqlxorm.SessionOption) ([]db.PipelineCmsConfig, error) {
		res := make([]db.PipelineCmsConfig, 0)

		// ns 1: user-1-org-1 configs
		if contains(cmsNsIDs, 1) {
			res = append(res,
				db.PipelineCmsConfig{
					ID:      10,
					NsID:    1,
					Key:     keyUserToken,
					Value:   valueUserToken,
					Encrypt: &[]bool{false}[0],
				},
				db.PipelineCmsConfig{
					ID:      11,
					NsID:    1,
					Key:     keyOrgName,
					Value:   valueOrgName,
					Encrypt: &[]bool{false}[0],
				},
			)
		}

		// ns 2: pipeline-secrets-app-2-develop configs
		if contains(cmsNsIDs, 2) {
			res = append(res,
				db.PipelineCmsConfig{
					ID:      20,
					NsID:    2,
					Key:     keyBuildProfile,
					Value:   valueBuildProfileTest,
					Encrypt: &[]bool{false}[0],
				},
				db.PipelineCmsConfig{
					ID:      21,
					NsID:    2,
					Key:     keyEnvType,
					Value:   valueEnvTypeDevelop,
					Encrypt: &[]bool{false}[0],
				},
				db.PipelineCmsConfig{
					ID:      22,
					NsID:    2,
					Key:     keyApiHost,
					Value:   valueApiHost,
					Encrypt: &[]bool{false}[0],
				},
			)
		}

		// ns 3: pipeline-secrets-app-2-default configs
		if contains(cmsNsIDs, 3) {
			res = append(res,
				db.PipelineCmsConfig{
					ID:      30,
					NsID:    3,
					Key:     keyBuildProfile,
					Value:   valueBuildProfileDev,
					Encrypt: &[]bool{false}[0],
				},
				db.PipelineCmsConfig{
					ID:      31,
					NsID:    3,
					Key:     keyEnvType,
					Value:   valueEnvTypeDefault,
					Encrypt: &[]bool{false}[0],
				},
				db.PipelineCmsConfig{
					ID:      32,
					NsID:    3,
					Key:     keyDatabaseURL,
					Value:   valueDatabaseURL,
					Encrypt: &[]bool{false}[0],
				},
				db.PipelineCmsConfig{
					ID:      33,
					NsID:    3,
					Key:     keyRedisURL,
					Value:   valueRedisURL,
					Encrypt: &[]bool{false}[0],
				},
			)
		}

		return res, nil
	})
	defer pm2.Unpatch()

	cm := pipelineCm{
		dbClient: dbClient,
	}

	// request order: [3,2,1] (default, develop, user)
	configs, err := cm.BatchGetConfigs(context.Background(), &pb.CmsNsConfigsBatchGetRequest{
		PipelineSource: "dice",
		Namespaces: []string{
			nsDefault, // id 3
			nsDevelop, // id 2
			nsUser,    // id 1
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, 9, len(configs)) // 4 + 3 + 2 = 9 configs total

	// verify order: should be [3,2,1] namespace order
	// first 4 configs should be from ns 3 (pipeline-secrets-app-2-default)
	assert.Equal(t, nsDefault, configs[0].Ns.Ns)
	assert.Equal(t, nsDefault, configs[1].Ns.Ns)
	assert.Equal(t, nsDefault, configs[2].Ns.Ns)
	assert.Equal(t, nsDefault, configs[3].Ns.Ns)

	// next 3 configs should be from ns 2 (pipeline-secrets-app-2-develop)
	assert.Equal(t, nsDevelop, configs[4].Ns.Ns)
	assert.Equal(t, nsDevelop, configs[5].Ns.Ns)
	assert.Equal(t, nsDevelop, configs[6].Ns.Ns)

	// last 2 configs should be from ns 1 (user-1-org-1)
	assert.Equal(t, nsUser, configs[7].Ns.Ns)
	assert.Equal(t, nsUser, configs[8].Ns.Ns)

	// verify specific key values in correct positions
	buildProfileConfigs := make([]string, 0)
	for _, config := range configs {
		if config.Key == keyBuildProfile {
			buildProfileConfigs = append(buildProfileConfigs, config.Value)
		}
	}
	// BUILD_PROFILE should appear in order: ["dev", "test"]
	// (dev from ns 3 first, then test from ns 2)
	assert.Equal(t, []string{valueBuildProfileDev, valueBuildProfileTest}, buildProfileConfigs)

	// verify unique keys appear in correct namespaces
	var foundUserToken, foundDatabaseURL, foundApiHost bool
	for _, config := range configs {
		if config.Key == keyUserToken {
			assert.Equal(t, nsUser, config.Ns.Ns)
			assert.Equal(t, valueUserToken, config.Value)
			foundUserToken = true
		}
		if config.Key == keyDatabaseURL {
			assert.Equal(t, nsDefault, config.Ns.Ns)
			assert.Equal(t, valueDatabaseURL, config.Value)
			foundDatabaseURL = true
		}
		if config.Key == keyApiHost {
			assert.Equal(t, nsDevelop, config.Ns.Ns)
			assert.Equal(t, valueApiHost, config.Value)
			foundApiHost = true
		}
	}
	assert.True(t, foundUserToken, "USER_TOKEN should be found in user-1-org-1")
	assert.True(t, foundDatabaseURL, "DATABASE_URL should be found in pipeline-secrets-app-2-default")
	assert.True(t, foundApiHost, "API_HOST should be found in pipeline-secrets-app-2-develop")
}

// helper function to check if slice contains value
func contains(slice []uint64, val uint64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
