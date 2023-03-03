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

package registry

import (
	"testing"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/registryhelper"
)

////go:generate mockgen -destination=./cluster_service_test.go -package registry github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb ClusterServiceServer
func Test_DeleteManifests(t *testing.T) {
	type args struct {
		clusterName string
		images      []string
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
		args args
	}{
		{
			name: "case1",
			args: args{
				clusterName: "hello",
				images: []string{
					"addon-registry.default.svc.cluster.local:5000/busybox:v0.1",
				},
			},
		},
	}

	clusterService := NewMockClusterServiceServer(ctrl)
	clusterService.EXPECT().GetCluster(gomock.Any(), gomock.Any()).AnyTimes().Return(&pb.GetClusterResponse{
		Data: &pb.ClusterInfo{
			Cm: map[string]string{
				apistructs.REGISTRY_ADDR.String():   "addon-registry.default.svc.cluster.local:5000",
				apistructs.REGISTRY_SCHEME.String(): "https",
			},
		},
	}, nil)

	r := New(clusterService)

	for _, tt := range tests {
		monkey.Patch(registryhelper.RemoveManifests, func(req registryhelper.RemoveManifestsRequest) (
			*registryhelper.RemoveManifestsResponse, error) {
			return &registryhelper.RemoveManifestsResponse{
				Succeed: tt.args.images,
			}, nil
		})
		t.Run(tt.name, func(t *testing.T) {
			err := r.DeleteManifests(tt.args.clusterName, tt.args.images)
			assert.NoError(t, err)
		})
		monkey.Unpatch(registryhelper.RemoveManifests)
	}
}
