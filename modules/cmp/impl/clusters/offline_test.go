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

package clusters

import (
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/dbclient"
)

func TestOfflineEdgeCluster(t *testing.T) {
	var bdl *bundle.Bundle
	var db *dbclient.DBClient

	req := apistructs.OfflineEdgeClusterRequest{
		ClusterName: "fake-cluster",
		Force:       true,
	}

	// monkey patch Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "QueryClusterInfo", func(_ *bundle.Bundle, _ string) (apistructs.ClusterInfoData, error) {
		return apistructs.ClusterInfoData{apistructs.DICE_IS_EDGE: "true"}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "DereferenceCluster", func(_ *bundle.Bundle, _ uint64, _, _ string) (string, error) {
		return "", nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "DeleteCluster", func(_ *bundle.Bundle, _ string, _ ...http.Header) error {
		return nil
	})

	// monkey record delete func
	monkey.Patch(updateDeleteRecord, func(_ *dbclient.DBClient, _ dbclient.Record) (uint64, error) {
		return 0, nil
	})

	c := New(db, bdl, nil)

	// monkey patch Credential with core services
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "DeleteAccessKey", func(*Clusters, string) error {
		return nil
	})

	_, err := c.OfflineEdgeCluster(req, "", "")
	assert.NoError(t, err)
}
