package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

// ReplaceOrCreateDaemonSet 通过 API Server 地址创建 DaemonSet
func ReplaceOrCreateDaemonSet(as string, ds appsv1.DaemonSet) error {
	_, err := ReadDaemonSet(as, ds.Namespace, ds.Name)
	if err == ErrNotFound {
		return CreateDaemonSet(as, ds)
	}
	if err != nil {
		return err
	}
	return ReplaceDaemonSet(as, ds)
}

// ReplaceOrCreateDeployment 通过 API Server 地址创建 Deployment
func ReplaceOrCreateDeployment(as string, deploy appsv1.Deployment) error {
	_, err := ReadDeployment(as, deploy.Namespace, deploy.Name)
	if err == ErrNotFound {
		return CreateDeployment(as, deploy)
	}
	if err != nil {
		return err
	}
	return ReplaceDeployment(as, deploy)
}

// ReplaceOrCreateService 通过 API Server 地址创建 Service
func ReplaceOrCreateService(as string, svc corev1.Service) error {
	old, err := ReadService(as, svc.Namespace, svc.Name)
	if err == ErrNotFound {
		return CreateService(as, svc)
	}
	if err != nil {
		return err
	}
	svc.ObjectMeta.ResourceVersion = old.ObjectMeta.ResourceVersion
	svc.Spec.ClusterIP = old.Spec.ClusterIP
	return ReplaceService(as, svc)
}

// ReplaceOrCreateIngress 通过 API Server 地址创建 Ingress
func ReplaceOrCreateIngress(as string, ingress extensionsv1beta1.Ingress) error {
	_, err := ReadIngress(as, ingress.Namespace, ingress.Name)
	if err == ErrNotFound {
		return CreateIngress(as, ingress)
	}
	if err != nil {
		return err
	}
	return ReplaceIngress(as, ingress)
}
