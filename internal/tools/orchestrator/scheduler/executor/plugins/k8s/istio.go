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

package k8s

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/istioctl/engines"
)

func getIstioEngine(clusterName string, info apistructs.ClusterInfoData) (istioctl.IstioEngine, error) {
	istioInfo := info.GetIstioInfo()
	if !istioInfo.Installed {
		return istioctl.EmptyEngine, nil
	}
	// TODO: Take asm's kubeconfig to create the corresponding engine
	if istioInfo.IsAliyunASM {
		return istioctl.EmptyEngine, nil
	}
	// TODO: Combine version to choose
	localEngine, err := engines.NewLocalEngine(clusterName)
	if err != nil {
		return istioctl.EmptyEngine, errors.Errorf("create local istio engine failed, cluster:%s, err:%v", clusterName, err)
	}
	return localEngine, nil
}
