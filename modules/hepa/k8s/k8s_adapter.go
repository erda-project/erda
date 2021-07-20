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
	"bytes"
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/modules/hepa/common/util"
	"github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/pkg/k8s/interface_factory"
	"github.com/erda-project/erda/pkg/k8s/union_interface"
	"github.com/erda-project/erda/pkg/k8sclient"
)

const (
	REWRITE_HOST_KEY = "nginx.ingress.kubernetes.io/upstream-vhost"
	REWRITE_PATH_KEY = "nginx.ingress.kubernetes.io/rewrite-target"
	USE_REGEX_KEY    = "nginx.ingress.kubernetes.io/use-regex"
	SERVICE_PROTOCOL = "nginx.ingress.kubernetes.io/backend-protocol"
)

type BackendProtocl string

const (
	HTTP  BackendProtocl = "http"
	HTTPS BackendProtocl = "https"
	GRPC  BackendProtocl = "grpc"
	GRPCS BackendProtocl = "grpcs"
	FCGI  BackendProtocl = "fastcgi"
)

type RouteOptions struct {
	RewriteHost         *string
	RewritePath         *string
	UseRegex            bool
	EnableTLS           bool
	BackendProtocol     *BackendProtocl
	InjectRuntimeDomain bool
	Annotations         map[string]*string
	LocationSnippet     *string
}

type IngressRoute union_interface.IngressRoute

type IngressBackend union_interface.IngressBackend

type K8SAdapter interface {
	IsGatewaySupportHttps(namespace string) (bool, error)
	MakeGatewaySupportHttps(namespace string) error
	MakeGatewaySupportMesh(namespace string) error
	CountIngressController() (int, error)
	CheckDomainExist(name string) (bool, error)
	DeleteIngress(namespace, name string) error
	CreateOrUpdateIngress(namespace, name string, routes []IngressRoute, backend IngressBackend, options ...RouteOptions) (bool, error)
	SetUpstreamHost(namespace, name, host string) error
	SetRewritePath(namespace, name, target string) error
	EnableRegex(namespace, name string) error
	CheckIngressExist(namespace, name string) (bool, error)
	UpdateIngressAnnotaion(namespace, name string, annotaion map[string]*string, snippet *string) error
	UpdateIngressConroller(options map[string]*string, mainSnippet, httpSnippet, serverSnippet *string) error
}

type K8SAdapterImpl struct {
	client          *kubernetes.Clientset
	ingressesHelper union_interface.IngressesHelper
	pool            *util.GPool
}

const (
	HEPA_BEGIN          = "###HEPA-AUTO-BEGIN###\n"
	HEPA_END            = "###HEPA-AUTO-END###\n"
	SYSTEM_NS           = "kube-system"
	GATEWAY_SVC_NAME    = "api-gateway"
	INGRESS_APP_LABEL   = "app.kubernetes.io/name=ingress-nginx"
	INGRESS_CONFIG_NAME = "nginx-configuration"
	LOC_SNIPPET_KEY     = "nginx.ingress.kubernetes.io/configuration-snippet"
	MAIN_SNIPPET_KEY    = "main-snippet"
	HTTP_SNIPPET_KEY    = "http-snippet"
	SERVER_SNIPPET_KEY  = "server-snippet"
)

func (impl *K8SAdapterImpl) CountIngressController() (int, error) {
	pods, err := impl.client.CoreV1().Pods(SYSTEM_NS).List(context.Background(), metav1.ListOptions{
		LabelSelector: INGRESS_APP_LABEL,
	})
	if err != nil {
		return 0, errors.WithStack(err)
	}
	if pods == nil || len(pods.Items) == 0 {
		logrus.Warnf("can't find any ingress controllers with label:%s, use default count:1", INGRESS_APP_LABEL)
		return 1, nil
	}
	return len(pods.Items), nil
}

func (impl *K8SAdapterImpl) IsGatewaySupportHttps(namespace string) (bool, error) {
	svc, err := impl.client.CoreV1().Services(namespace).Get(context.Background(), GATEWAY_SVC_NAME, metav1.GetOptions{})
	if err != nil {
		return false, errors.WithStack(err)
	}
	if svc == nil {
		return false, errors.New("can't find gateway svc")
	}
	supportHttps := false
	for _, port := range svc.Spec.Ports {
		if port.Port == int32(vars.KONG_HTTPS_SERVICE_PORT) {
			supportHttps = true
			break
		}
	}
	return supportHttps, nil
}

func (impl *K8SAdapterImpl) MakeGatewaySupportMesh(namespace string) error {
	ns, err := impl.client.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	if ns == nil || ns.Name == "" {
		return errors.New("can't find namespace")
	}
	if len(ns.Labels) == 0 {
		ns.Labels = map[string]string{}
	}
	if ns.Labels["istio-injection"] == "enabled" {
		return nil
	}
	ns.Labels["istio-injection"] = "enabled"
	_, err = impl.client.CoreV1().Namespaces().Update(context.Background(), ns, metav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsResourceExpired(err) {
			return nil
		}
		return errors.WithStack(err)
	}
	deployment, err := impl.client.AppsV1().Deployments(namespace).Get(context.Background(), "api-gateway", metav1.GetOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	if deployment == nil {
		return errors.New("can't find deployment")
	}
	containers := deployment.Spec.Template.Spec.Containers
	for i := 0; i < len(containers); i++ {
		if containers[i].Name == "api-gateway" &&
			(containers[i].LivenessProbe.TCPSocket != nil || containers[i].ReadinessProbe.TCPSocket != nil) {
			containers[i].LivenessProbe.TCPSocket = nil
			containers[i].LivenessProbe.HTTPGet = nil
			containers[i].LivenessProbe.Exec = &v1.ExecAction{
				Command: []string{"kong", "health"},
			}
			containers[i].ReadinessProbe.TCPSocket = nil
			containers[i].ReadinessProbe.HTTPGet = nil
			containers[i].ReadinessProbe.Exec = &v1.ExecAction{
				Command: []string{"kong", "health"},
			}
			containers[i].Env = append(containers[i].Env, v1.EnvVar{
				Name:  "SERVICE_MESH",
				Value: "on",
			})
		}
	}
	_, err = impl.client.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsResourceExpired(err) {
			return nil
		}
		return errors.WithStack(err)
	}
	return nil
}

func (impl *K8SAdapterImpl) MakeGatewaySupportHttps(namespace string) error {
	ns := impl.client.CoreV1().Services(namespace)
	svc, err := ns.Get(context.Background(), GATEWAY_SVC_NAME, metav1.GetOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	if svc == nil || svc.Name == "" {
		return errors.New("can't find gateway svc")
	}
	supportHttps := false
	for _, port := range svc.Spec.Ports {
		if port.Port == int32(vars.KONG_HTTPS_SERVICE_PORT) {
			supportHttps = true
			break
		}
	}
	if !supportHttps {
		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       "https-" + "gateway",
			Protocol:   v1.ProtocolTCP,
			Port:       int32(vars.KONG_HTTPS_SERVICE_PORT),
			TargetPort: intstr.FromInt(vars.KONG_HTTPS_SERVICE_PORT),
		})
		_, err = ns.Update(context.Background(), svc, metav1.UpdateOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func doRecover() {
	if r := recover(); r != nil {
		log.Errorf("recovered from: %+v ", r)
		debug.PrintStack()
	}
}

func (impl *K8SAdapterImpl) CheckDomainExist(domain string) (bool, error) {
	nsList, err := impl.client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return false, errors.WithStack(err)
	}
	exist := false
	wg := sync.WaitGroup{}
	for _, ns := range nsList.Items {
		impl.pool.Acquire(1)
		wg.Add(1)
		go func(nsName string) {
			defer doRecover()
			ingressList, err := impl.client.ExtensionsV1beta1().Ingresses(nsName).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				log.Errorf("ingress error happened:%+v", errors.WithStack(err))
				goto done
			}
			for _, ingress := range ingressList.Items {
				for _, rule := range ingress.Spec.Rules {
					if domain == rule.Host {
						log.Infof("domain %s already exists, ns:%s, ingress:%s",
							domain, nsName, ingress.Name)
						exist = true
						goto done
					}
				}
			}
		done:
			wg.Done()
			impl.pool.Release(1)
		}(ns.Name)
	}
	wg.Wait()
	return exist, nil
}

func (impl *K8SAdapterImpl) DeleteIngress(namespace, name string) error {
	exist, err := impl.CheckIngressExist(namespace, name)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	err = impl.ingressesHelper.Ingresses(namespace).Delete(context.Background(), strings.ToLower(name), metav1.DeleteOptions{})
	if err != nil {
		return errors.Errorf("delete ingress %s failed, ns:%s, err:%s", name, namespace, err)
	}
	return nil
}

func (impl *K8SAdapterImpl) newIngress(ns, name string, routes []IngressRoute, backend IngressBackend, needTLS bool) interface{} {
	material := union_interface.IngressMaterial{
		Name:      strings.ToLower(name),
		Namespace: ns,
		Routes:    *(*[]union_interface.IngressRoute)(unsafe.Pointer(&routes)),
		Backend:   union_interface.IngressBackend(backend),
		NeedTLS:   needTLS,
	}
	return impl.ingressesHelper.NewIngress(material)

}

func (impl *K8SAdapterImpl) setOptionAnnotations(ingress interface{}, options RouteOptions) error {
	annotations := map[string]string{}
	if options.RewriteHost != nil {
		annotations[REWRITE_HOST_KEY] = *options.RewriteHost
	}
	if options.RewritePath != nil {
		annotations[REWRITE_PATH_KEY] = *options.RewritePath
	}
	if options.UseRegex {
		annotations[USE_REGEX_KEY] = "true"
	}
	if options.BackendProtocol != nil {
		switch *options.BackendProtocol {
		case HTTP:
			annotations[SERVICE_PROTOCOL] = "HTTP"
		case HTTPS:
			annotations[SERVICE_PROTOCOL] = "HTTPS"
		case GRPC:
			annotations[SERVICE_PROTOCOL] = "GRPC"
		case GRPCS:
			annotations[SERVICE_PROTOCOL] = "GRPCS"
		case FCGI:
			annotations[SERVICE_PROTOCOL] = "FCGI"
		}
	} else {
		impl.ingressesHelper.IngressAnnotationClear(ingress, SERVICE_PROTOCOL)
	}
	for key, value := range options.Annotations {
		if value == nil {
			impl.ingressesHelper.IngressAnnotationClear(ingress, key)
			continue
		}
		annotations[key] = *value
	}
	if options.LocationSnippet != nil {
		locationSnippet, err := impl.ingressesHelper.IngressAnnotationGet(ingress, LOC_SNIPPET_KEY)
		if err != nil {
			return err
		}
		newSnippet, err := impl.replaceSnippet(locationSnippet, *options.LocationSnippet)
		if err != nil {
			return err
		}
		annotations[LOC_SNIPPET_KEY] = newSnippet

	}
	err := impl.ingressesHelper.IngressAnnotationBatchSet(ingress, annotations)
	if err != nil {
		return err
	}
	return nil
}

func (impl *K8SAdapterImpl) CreateOrUpdateIngress(namespace, name string, routes []IngressRoute, backend IngressBackend, options ...RouteOptions) (bool, error) {
	ns := impl.ingressesHelper.Ingresses(namespace)
	ingressName := strings.ToLower(name)
	exist, err := ns.Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return false, errors.WithStack(err)
	}
	var ingress interface{}
	routeOptions := RouteOptions{}
	if len(options) > 0 {
		routeOptions = options[0]
	}
	ingress = impl.newIngress(namespace, ingressName, routes, backend, routeOptions.EnableTLS)
	if k8serrors.IsNotFound(err) {
		err := impl.setOptionAnnotations(ingress, routeOptions)
		if err != nil {
			return false, err
		}
		log.Debugf("begin create ingress, name:%s, ns:%s", ingressName, namespace)
		if !routeOptions.InjectRuntimeDomain {
			_, err = ns.Create(context.Background(), ingress, metav1.CreateOptions{})
			if err != nil {
				return false, errors.Errorf("create ingress %s failed, ns:%s, err:%s",
					ingressName, namespace, err)
			}
			log.Infof("new ingress created, name:%s, ns:%s", ingressName, namespace)
			return false, nil
		} else {
			//TODO optimize kong sync
			go func() {
				log.Infof("start async create ingress, name:%s, ns:%s", ingressName, namespace)
				time.Sleep(time.Duration(60) * time.Second)
				_, err = ns.Create(context.Background(), ingress, metav1.CreateOptions{})
				if err != nil {
					log.Errorf("create ingress %s failed, ns:%s, err:%s",
						ingressName, namespace, err)
					return
				}
				log.Infof("new ingress created, name:%s, ns:%s", ingressName, namespace)
			}()
			return false, nil
		}
	}
	oldAnnotations, err := impl.ingressesHelper.IngressAnnotationBatchGet(exist)
	if err != nil {
		return true, err
	}
	err = impl.ingressesHelper.IngressAnnotationBatchSet(ingress, oldAnnotations)
	if err != nil {
		return true, err
	}
	err = impl.setOptionAnnotations(ingress, routeOptions)
	if err != nil {
		return true, err
	}
	log.Debugf("begin update ingress, name:%s, ns:%s", ingressName, namespace)
	_, err = ns.Update(context.Background(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return true, errors.Errorf("update ingress %s failed, ns:%s, err:%s",
			ingressName, namespace, err)
	}
	log.Infof("ingress updated, name:%s, ns:%s", ingressName, namespace)
	return true, nil
}

func (impl *K8SAdapterImpl) SetUpstreamHost(namespace, name, host string) error {
	ns := impl.ingressesHelper.Ingresses(namespace)
	ingressName := strings.ToLower(name)
	ingress, err := ns.Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("get ingress %s failed, ns:%s, err:%s", ingressName, namespace, err)
	}
	err = impl.ingressesHelper.IngressAnnotationSet(ingress, "nginx.ingress.kubernetes.io/upstream-vhost", host)
	if err != nil {
		return err
	}
	_, err = ns.Update(context.Background(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return errors.Errorf("set upstream host %s failed, name:%s, ns:%s, err:%s",
			host, ingressName, namespace, err)
	}
	return nil
}

func (impl *K8SAdapterImpl) SetRewritePath(namespace, name, target string) error {
	ns := impl.ingressesHelper.Ingresses(namespace)
	ingressName := strings.ToLower(name)
	ingress, err := ns.Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("get ingress %s failed, ns:%s, err:%s", ingressName, namespace, err)
	}
	err = impl.ingressesHelper.IngressAnnotationSet(ingress, "nginx.ingress.kubernetes.io/rewrite-target", target)
	if err != nil {
		return err
	}
	_, err = ns.Update(context.Background(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return errors.Errorf("set rewrite path %s failed, name:%s, ns:%s, err:%s",
			target, ingressName, namespace, err)
	}
	return nil
}

func (impl *K8SAdapterImpl) EnableRegex(namespace, name string) error {
	ns := impl.ingressesHelper.Ingresses(namespace)
	ingressName := strings.ToLower(name)
	ingress, err := ns.Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("get ingress %s failed, ns:%s, err:%s", ingressName, namespace, err)
	}
	err = impl.ingressesHelper.IngressAnnotationSet(ingress, "nginx.ingress.kubernetes.io/use-regex", "true")
	if err != nil {
		return err
	}
	_, err = ns.Update(context.Background(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return errors.Errorf("enable regex failed, name:%s, ns:%s, err:%s",
			ingressName, namespace, err)
	}
	return nil
}

func (impl *K8SAdapterImpl) replaceSnippet(source, replace string) (string, error) {
	replace += "\n"
	beginIndex := strings.Index(source, HEPA_BEGIN)
	var b bytes.Buffer
	if beginIndex == -1 {
		source += "\n"
		_, _ = b.WriteString(source)
		_, _ = b.WriteString(HEPA_BEGIN)
		_, _ = b.WriteString(replace)
		_, _ = b.WriteString(HEPA_END)
		return b.String(), nil

	}
	endIndex := strings.Index(source, HEPA_END)
	if endIndex == -1 {
		return "", errors.Errorf("invalid source snippet:%s", source)
	}
	prefix := source[:beginIndex]
	suffix := source[endIndex+len(HEPA_END):]

	_, _ = b.WriteString(prefix)
	_, _ = b.WriteString(HEPA_BEGIN)
	_, _ = b.WriteString(replace)
	_, _ = b.WriteString(HEPA_END)
	_, _ = b.WriteString(suffix)
	return b.String(), nil
}

func (impl *K8SAdapterImpl) CheckIngressExist(namespace, name string) (bool, error) {
	ns := impl.ingressesHelper.Ingresses(namespace)
	_, err := ns.Get(context.Background(), strings.ToLower(name), metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return false, errors.WithStack(err)
	}
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	return true, nil
}

func (impl *K8SAdapterImpl) UpdateIngressAnnotaion(namespace, name string, annotaion map[string]*string, snippet *string) error {
	ns := impl.ingressesHelper.Ingresses(namespace)
	ingressName := strings.ToLower(name)
	ingress, err := ns.Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("get ingress %s failed, ns:%s, err:%s", ingressName, namespace, err)
		return errors.Errorf("ingress %s is creating, please retry after about 60 seconds", ingressName)
	}
	for key, value := range annotaion {
		if value == nil {
			impl.ingressesHelper.IngressAnnotationClear(ingress, key)
			continue
		}
		impl.ingressesHelper.IngressAnnotationSet(ingress, key, *value)
	}
	if snippet != nil {
		locationSnippet, err := impl.ingressesHelper.IngressAnnotationGet(ingress, LOC_SNIPPET_KEY)
		if err != nil {
			return err
		}
		newSnippet, err := impl.replaceSnippet(locationSnippet, *snippet)
		if err != nil {
			return err
		}
		impl.ingressesHelper.IngressAnnotationSet(ingress, LOC_SNIPPET_KEY, newSnippet)
		log.Debugf("ns:%s ingress:%s new snippet:%s", namespace, ingressName, newSnippet)
	}
	_, err = ns.Update(context.Background(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return errors.Errorf("update ingress annotation failed, name:%s, ns:%s, err:%s",
			ingressName, namespace, err)
	}
	return nil
}

func (impl *K8SAdapterImpl) UpdateIngressConroller(options map[string]*string, mainSnippet, httpSnippet, serverSnippet *string) error {
	ns := impl.client.CoreV1().ConfigMaps(SYSTEM_NS)
	configmap, err := ns.Get(context.Background(), INGRESS_CONFIG_NAME, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("get ingress config map failed, err:%s", err)
	}
	for key, value := range options {
		if value == nil {
			delete(configmap.Data, key)
			continue
		}
		if configmap.Data == nil {
			configmap.Data = map[string]string{}
		}
		configmap.Data[key] = *value
	}
	if mainSnippet != nil {
		if configmap.Data == nil {
			configmap.Data = map[string]string{}
		}
		snippet := configmap.Data[MAIN_SNIPPET_KEY]
		newSnippet, err := impl.replaceSnippet(snippet, *mainSnippet)
		if err != nil {
			return err
		}
		configmap.Data[MAIN_SNIPPET_KEY] = newSnippet
		log.Debugf("ingress conrtoller new main snippet:%s", newSnippet)
	}
	if httpSnippet != nil {
		if configmap.Data == nil {
			configmap.Data = map[string]string{}
		}
		snippet := configmap.Data[HTTP_SNIPPET_KEY]
		newSnippet, err := impl.replaceSnippet(snippet, *httpSnippet)
		if err != nil {
			return err
		}
		configmap.Data[HTTP_SNIPPET_KEY] = newSnippet
		log.Debugf("ingress conrtoller new http snippet:%s", newSnippet)
	}
	if serverSnippet != nil {
		if configmap.Data == nil {
			configmap.Data = map[string]string{}
		}
		snippet := configmap.Data[SERVER_SNIPPET_KEY]
		newSnippet, err := impl.replaceSnippet(snippet, *serverSnippet)
		if err != nil {
			return err
		}
		configmap.Data[SERVER_SNIPPET_KEY] = newSnippet
		log.Debugf("ingress controller new server snippet:%s", newSnippet)
	}
	_, err = ns.Update(context.Background(), configmap, metav1.UpdateOptions{})
	if err != nil {
		return errors.Errorf("ingress controller update configmap failed, err:%s", err)
	}
	return nil
}

func NewAdapter(clusterKey string) (K8SAdapter, error) {
	client, err := k8sclient.New(clusterKey)
	if err != nil {
		return nil, err
	}
	helper, err := interface_factory.CreateIngressesHelper(client.ClientSet)
	if err != nil {
		return nil, err
	}
	pool := util.NewGPool(1000)
	return &K8SAdapterImpl{
		client:          client.ClientSet,
		ingressesHelper: helper,
		pool:            pool,
	}, nil
}
