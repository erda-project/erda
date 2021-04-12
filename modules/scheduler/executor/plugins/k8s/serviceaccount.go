// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newServiceAccount(name, namespace string, imageSecrets []string) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	for _, is := range imageSecrets {
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
			Name: is,
		})
	}

	return sa
}

func (k *Kubernetes) updateDefaultServiceAccountForImageSecret(namespace, secretName string) error {
	var err error

	// Try to create first, then update after failure
	// k8s will automatically create the default serviceaccount, but there will be a delay, resulting in failure to update the probability.
	if err = k.sa.Create(newServiceAccount(defaultServiceAccountName, namespace, []string{secretName})); err != nil {
		for {
			serviceaccount, err := k.sa.Get(namespace, defaultServiceAccountName)
			if err != nil {
				return err
			}

			serviceaccount.ImagePullSecrets = append(serviceaccount.ImagePullSecrets, corev1.LocalObjectReference{
				Name: secretName,
			})
			err = k.sa.Patch(serviceaccount)
			if err == nil {
				break
			}
			if err.Error() != "Conflict" {
				return err
			}
		}
	}

	return nil
}
