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

package logic

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateInnerSecretIfNotExist(cs kubernetes.Interface, sourceNs, targetNs, secretName string) (bool, error) {
	if cs == nil {
		return false, errors.New("kubernetes client is nil")
	}

	// secret already exists in target namespace
	// TODO: secret content update
	if _, err := cs.CoreV1().Secrets(targetNs).Get(context.Background(), secretName, metav1.GetOptions{}); err == nil {
		return true, nil
	}

	s, err := cs.CoreV1().Secrets(sourceNs).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	targetSec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
		},
		Data:       s.Data,
		StringData: s.StringData,
		Type:       s.Type,
	}

	if _, err = cs.CoreV1().Secrets(targetNs).Create(context.Background(), targetSec, metav1.CreateOptions{}); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return true, nil
		}
		return false, err
	}

	return true, nil
}
