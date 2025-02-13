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
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/conf"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
)

func (k *Kubernetes) composeRegistryInfos(sg *apistructs.ServiceGroup) []apistructs.RegistryInfo {
	registryInfos := []apistructs.RegistryInfo{}

	for _, service := range sg.Services {
		if service.ImageUsername != "" {
			registryInfo := apistructs.RegistryInfo{}
			registryInfo.Host = strings.Split(service.Image, "/")[0]
			registryInfo.UserName = service.ImageUsername
			registryInfo.Password = service.ImagePassword
			registryInfos = append(registryInfos, registryInfo)
		}
	}
	return registryInfos
}

func (k *Kubernetes) setImagePullSecrets(namespace string) ([]apiv1.LocalObjectReference, error) {
	secrets := make([]apiv1.LocalObjectReference, 0, 1)
	secretName := conf.CustomRegCredSecret()

	_, err := k.secret.Get(namespace, secretName)
	if err == nil {
		secrets = append(secrets,
			apiv1.LocalObjectReference{
				Name: secretName,
			})
	} else {
		if !k8serror.NotFound(err) {
			return nil, fmt.Errorf("get secret %s in namespace %s err: %v", secretName, namespace, err)
		}
	}
	return secrets, nil
}
