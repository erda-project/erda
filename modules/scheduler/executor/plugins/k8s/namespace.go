package k8s

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/strutil"
)

// MakeNamespace 生成一个 Namespace 名字
// 每一个runtime对应到k8s上是一个k8s namespace,
// 格式为 ${runtimeNamespace}--${runtimeName}
func MakeNamespace(sg *apistructs.ServiceGroup) string {
	if IsGroupStateful(sg) {
		// 针对需要拆成多个 statefulset 的 servicegroup 创建一个新的 namespace, 即加上group-的前缀
		if v, ok := sg.Labels[groupNum]; ok && v != "" && v != "1" {
			return strutil.Concat("group-", sg.Type, "--", sg.ID)
		}
	}
	return strutil.Concat(sg.Type, "--", sg.ID)
}

// CreateNamespace 创建 namespace
func (k *Kubernetes) CreateNamespace(ns string, sg *apistructs.ServiceGroup) error {
	notfound, err := k.NotfoundNamespace(ns)
	if err != nil {
		return err
	}

	if !notfound {
		if sg.ProjectNamespace != "" {
			return nil
		}
		return errors.Errorf("failed to create namespace, ns: %s, (namespace already exists)", ns)
	}

	labels := map[string]string{}

	if sg.Labels["service-mesh"] == "on" {
		labels["istio-injection"] = "enabled"
	}

	if err = k.namespace.Create(ns, labels); err != nil {
		return err
	}
	// 创建该 namespace 下的 imagePullSecret
	if err = k.NewRuntimeImageSecret(ns, sg); err != nil {
		logrus.Errorf("failed to create imagePullSecret, namespace: %s, (%v)", ns, err)
	}
	return nil
}

// UpdateNamespace
func (k *Kubernetes) UpdateNamespace(ns string, sg *apistructs.ServiceGroup) error {
	notfound, err := k.NotfoundNamespace(ns)
	if err != nil {
		return err
	}
	if notfound {
		return errors.Errorf("not found ns: %v", ns)
	}

	labels := map[string]string{}

	if sg.Labels["service-mesh"] == "on" {
		labels["istio-injection"] = "enabled"
	}

	return k.namespace.Update(ns, labels)
}

// NotfoundNamespace not found namespace
func (k *Kubernetes) NotfoundNamespace(ns string) (bool, error) {
	err := k.namespace.Exists(ns)
	if err != nil {
		if k8serror.NotFound(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// DeleteNamespace delete namepsace
func (k *Kubernetes) DeleteNamespace(ns string) error {
	return k.namespace.Delete(ns)
}
