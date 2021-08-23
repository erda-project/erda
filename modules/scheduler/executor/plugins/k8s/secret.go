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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

// NewImageSecret create new image pull secret
// 1, create imagePullSecret of this namespace
// 2, put this secret into serviceaccount of the namespace
func (k *Kubernetes) NewImageSecret(namespace string) error {
	// When the cluster is initialized, a secret to pull the mirror will be created in the default namespace
	s, err := k.secret.Get("default", AliyunRegistry)
	if err != nil {
		return err
	}

	mysecret := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: namespace,
		},
		Data:       s.Data,
		StringData: s.StringData,
		Type:       s.Type,
	}

	if err := k.secret.Create(mysecret); err != nil {
		return err
	}

	return k.updateDefaultServiceAccountForImageSecret(namespace, s.Name)
}

// NewImageSecret create mew image pull secret
// 1, create imagePullSecret of this namespace
// 2, Add the secret of the image that needs to be authenticated to the secret of the namespace
// 3, put this secret into serviceaccount of the namespace
func (k *Kubernetes) NewRuntimeImageSecret(namespace string, sg *apistructs.ServiceGroup) error {
	// When the cluster is initialized, a secret to pull the mirror will be created in the default namespace
	s, err := k.secret.Get("default", AliyunRegistry)
	if err != nil {
		return err
	}

	var dockerConfigJson apistructs.RegistryAuthJson
	if err := json.Unmarshal(s.Data[".dockerconfigjson"], &dockerConfigJson); err != nil {
		return err
	}

	//Append the runtime secret with the username and password to the secret
	for _, service := range sg.Services {
		if service.ImageUsername != "" {
			u := strings.Split(service.Image, "/")[0]
			authString := base64.StdEncoding.EncodeToString([]byte(service.ImageUsername + ":" + service.ImagePassword))
			dockerConfigJson.Auths[u] = apistructs.RegistryUserInfo{Auth: authString}
		}
	}

	var sData []byte
	if sData, err = json.Marshal(dockerConfigJson); err != nil {
		return err
	}

	mysecret := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: namespace,
		},
		Data: map[string][]byte{".dockerconfigjson": sData},
		Type: s.Type,
	}

	if err := k.secret.Create(mysecret); err != nil {
		return err
	}

	return k.updateDefaultServiceAccountForImageSecret(namespace, s.Name)
}

// CopyErdaSecrets Copy the secret under orignns namespace to dstns
func (k *Kubernetes) CopyErdaSecrets(originns, dstns string) ([]apiv1.Secret, error) {
	secrets, err := k.secret.List(originns)
	if err != nil {
		return nil, err
	}
	result := []apiv1.Secret{}
	for _, secret := range secrets.Items {
		// ignore default token
		if !strutil.HasPrefixes(secret.Name, "dice-") {
			continue
		}
		dstsecret := &apiv1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secret.Name,
				Namespace: dstns,
			},
			Data:       secret.Data,
			StringData: secret.StringData,
			Type:       secret.Type,
		}
		if err := k.secret.CreateIfNotExist(dstsecret); err != nil {
			return nil, err
		}
		result = append(result, secret)
	}
	return result, nil
}

// SecretVolume
func (k *Kubernetes) SecretVolume(secret *apiv1.Secret) (apiv1.Volume, apiv1.VolumeMount) {
	return apiv1.Volume{
			Name: secret.Name,
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		},
		apiv1.VolumeMount{
			Name:      secret.Name,
			MountPath: fmt.Sprintf("/%s", secret.Name),
			ReadOnly:  true,
		}
}
