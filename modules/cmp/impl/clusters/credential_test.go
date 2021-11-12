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
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	credentialpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/pkg/k8sclient"
)

var (
	bdl         *bundle.Bundle
	db          *dbclient.DBClient
	fakeCluster = "fake-cluster"
	fakeAkItem  = &credentialpb.AccessKeysItem{
		Id:          "5e34b95b-cd06-464c-8ee9-3aef696586c6",
		AccessKey:   "Q9x5k4MJ89h327yqoc9zvvoP",
		Status:      credentialpb.StatusEnum_ACTIVATE,
		SubjectType: credentialpb.SubjectTypeEnum_CLUSTER,
		Subject:     fakeCluster,
		Scope:       apistructs.CMPClusterScope,
		ScopeId:     fakeCluster,
	}
)

////go:generate mockgen -destination=./credential_ak_test.go -package clusters github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb AccessKeyServiceServer
func Test_GetOrCreateAccessKey_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockAccessKeyServiceServer(ctrl)

	akService.EXPECT().QueryAccessKeys(gomock.Any(), gomock.Any()).AnyTimes().Return(&credentialpb.QueryAccessKeysResponse{
		Data: []*credentialpb.AccessKeysItem{},
	}, nil)

	akService.EXPECT().CreateAccessKey(gomock.Any(), gomock.Any()).AnyTimes().Return(&credentialpb.CreateAccessKeyResponse{
		Data: fakeAkItem,
	}, nil)

	c := New(db, bdl, akService)

	monkey.PatchInstanceMethod(reflect.TypeOf(c), "CheckCluster", func(_ *Clusters, _ string) error {
		return nil
	})

	defer monkey.UnpatchAll()

	akResp, err := c.GetOrCreateAccessKey(fakeCluster)
	assert.NoError(t, err)
	assert.Equal(t, akResp, fakeAkItem)
}

func Test_GetOrCreateAccessKey_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockAccessKeyServiceServer(ctrl)

	akService.EXPECT().QueryAccessKeys(gomock.Any(), gomock.Any()).AnyTimes().Return(&credentialpb.QueryAccessKeysResponse{
		Data:  []*credentialpb.AccessKeysItem{fakeAkItem},
		Total: 1,
	}, nil)

	akService.EXPECT().CreateAccessKey(gomock.Any(), gomock.Any()).AnyTimes().Return(&credentialpb.CreateAccessKeyResponse{
		Data: fakeAkItem,
	}, nil)

	c := New(db, bdl, akService)
	akResp, err := c.GetOrCreateAccessKey(fakeCluster)
	assert.NoError(t, err)
	assert.Equal(t, akResp, fakeAkItem)
}

func Test_DeleteAccessKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockAccessKeyServiceServer(ctrl)

	akService.EXPECT().DeleteAccessKey(gomock.Any(), gomock.Any()).AnyTimes().Return(&credentialpb.DeleteAccessKeyResponse{}, nil)

	akService.EXPECT().QueryAccessKeys(gomock.Any(), gomock.Any()).AnyTimes().Return(&credentialpb.QueryAccessKeysResponse{
		Data:  []*credentialpb.AccessKeysItem{fakeAkItem},
		Total: 1,
	}, nil)

	c := New(db, bdl, akService)
	err := c.DeleteAccessKey(fakeCluster)
	assert.NoError(t, err)
}

func Test_ResetAccessKey_InCluster_Error(t *testing.T) {
	var (
		csErr            = errors.New("unit test, skip")
		emptyClusterName = ""
	)

	defer monkey.UnpatchAll()

	monkey.Patch(rest.InClusterConfig, func() (*rest.Config, error) {
		return nil, nil
	})

	monkey.Patch(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
		return nil, csErr
	})

	monkey.Patch(k8sclient.NewWithTimeOut, func(clusterName string, timeout time.Duration) (*k8sclient.K8sClient, error) {
		return nil, csErr
	})

	c := New(db, bdl, nil)
	_, err := c.ResetAccessKey(emptyClusterName)
	assert.Equal(t, err, fmt.Errorf("connect to cluster: %s error: %s", emptyClusterName, csErr.Error()))
}
