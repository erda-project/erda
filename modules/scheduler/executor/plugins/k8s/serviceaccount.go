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

	// 先尝试创建，失败后再进行更新
	// k8s 会自动创建 default serviceaccount, 但是会有时延，导致概率更新失败。
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
