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
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/pkg/transport"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/registryhelper"
)

type (
	Interface interface {
		DeleteManifests(clusterName string, images []string) (err error)
	}
	Service struct {
		clusterSvc clusterpb.ClusterServiceServer
	}
)

func New(clusterSvc clusterpb.ClusterServiceServer) Interface {
	return &Service{
		clusterSvc: clusterSvc,
	}
}

// DeleteManifests deletes manifests from the cluster inner image registry
func (s *Service) DeleteManifests(clusterName string, images []string) (err error) {
	var l = logrus.WithField("func", "DeleteManifests").
		WithField("clusterName", clusterName).
		WithField("images", images)
	if len(images) == 0 {
		return nil
	}

	if s.clusterSvc == nil {
		return errors.New("cluster service is nil")
	}

	ctx := transport.WithHeader(context.TODO(), metadata.New(map[string]string{httputil.InternalHeader: "dop"}))
	clusterResp, err := s.clusterSvc.GetCluster(ctx, &clusterpb.GetClusterRequest{
		IdOrName: clusterName,
	})
	if err != nil {
		l.WithError(err).Errorln("failed to get cluster info")
		return err
	}

	clusterInfo := clusterResp.Data.GetCm()
	if clusterInfo == nil {
		l.WithError(err).Errorln("failed to get cluster configmap")
		return err
	}

	registryAddr, ok := clusterInfo[apistructs.REGISTRY_ADDR.String()]
	if !ok {
		l.WithError(err).Errorln("failed to get registry address")
		return errors.Wrap(err, "failed to get registry address")
	}

	req := registryhelper.RemoveManifestsRequest{
		RegistryAddr:   registryAddr,
		RegistryScheme: clusterInfo[apistructs.REGISTRY_SCHEME.String()],
		Images:         images,
		ClusterKey:     clusterName,
	}

	removeResp, err := registryhelper.RemoveManifests(req)
	if err != nil {
		return err
	}

	if len(removeResp.Failed) > 0 {
		return errors.Errorf("recycle image fail: %+v", removeResp.Failed)
	}

	return nil
}
