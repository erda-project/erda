package k8s

import (
	apiv1 "k8s.io/api/core/v1"
)

func (k *Kubernetes) newPVC(pvc *apiv1.PersistentVolumeClaim) error {
	return k.pvc.Create(pvc)
}

// todo: deletePVC
