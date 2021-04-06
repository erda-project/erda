package k8s

import (
	"strings"

	"github.com/pkg/errors"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/erda-project/erda/apistructs"
)

func (k *Kubernetes) createIngress(svc *apistructs.Service) error {
	ing, err := buildIngress(svc)
	if err != nil {
		return err
	}
	if ing != nil {
		return k.ingress.Create(ing)
	}
	return nil
}
func buildIngress(svc *apistructs.Service) (*extensionsv1beta1.Ingress, error) {
	if svc.Labels["IS_ENDPOINT"] != "true" {
		return nil, nil
	}
	// 需要对公网暴露的服务
	// 将label中HAPROXY_0_VHOST对应的域名/vip集合都转发到该服务的第0个端口上
	publicHosts := strings.Split(svc.Labels["HAPROXY_0_VHOST"], ",")
	if len(publicHosts) == 0 {
		return nil, errors.Errorf("failed to set label IS_ENDPOINT true but label HAPROXY_0_VHOST empty, service: %s", svc.Name)
	}
	if len(svc.Ports) == 0 {
		return nil, errors.Errorf("failed to create ingress as ports is empty, service: %s", svc.Name)
	}
	// 创建ingress
	rules := buildRules(publicHosts, svc.Name, svc.Ports[0].Port)

	// tls
	tls := buildTLS(publicHosts)
	ingress := &extensionsv1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: rules,
			TLS:   tls,
		},
	}

	return ingress, nil
}

func buildRules(publicHosts []string, name string, port int) []extensionsv1beta1.IngressRule {
	rules := make([]extensionsv1beta1.IngressRule, len(publicHosts))
	for i, host := range publicHosts {
		rules[i].Host = host
		rules[i].HTTP = &extensionsv1beta1.HTTPIngressRuleValue{
			Paths: []extensionsv1beta1.HTTPIngressPath{
				{
					//TODO: add Path
					// Path:
					Backend: extensionsv1beta1.IngressBackend{
						ServiceName: name,
						ServicePort: intstr.FromInt(port),
					},
				},
			},
		}
	}
	return rules
}

func buildTLS(publicHosts []string) []extensionsv1beta1.IngressTLS {
	tls := make([]extensionsv1beta1.IngressTLS, 1)
	tls[0].Hosts = make([]string, len(publicHosts))
	for i, host := range publicHosts {
		tls[0].Hosts[i] = host
	}
	return tls
}

func (k *Kubernetes) updateIngress(svc *apistructs.Service) error {
	var err error

	ing, err := buildIngress(svc)
	if err != nil {
		return err
	}

	// 如果需要更新 ingress，则判断是 create 还是 update
	if ing != nil {
		return k.ingress.CreateOrUpdate(ing)
	}

	// 如果不需要更新，则判断是否需要删除残留的 ingress
	return k.ingress.DeleteIfExists(svc.Namespace, svc.Name)
}

// 删除ingress资源
func (k *Kubernetes) deleteIngress(namespace, name string) error {
	return k.ingress.Delete(namespace, name)
}
