package k8s

func (k *Kubernetes) killPod(k8snamespace, podname string) error {
	if err := k.pod.Delete(k8snamespace, podname); err != nil {
		return err
	}
	return nil
}
