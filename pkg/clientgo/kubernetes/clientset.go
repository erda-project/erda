package kubernetes

import (
	"k8s.io/client-go/discovery"
	admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	appsv1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	appsv1beta2 "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
	auditregistrationv1alpha1 "k8s.io/client-go/kubernetes/typed/auditregistration/v1alpha1"
	authenticationv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
	authenticationv1beta1 "k8s.io/client-go/kubernetes/typed/authentication/v1beta1"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	authorizationv1beta1 "k8s.io/client-go/kubernetes/typed/authorization/v1beta1"
	autoscalingv1 "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	autoscalingv2beta1 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta2"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	batchv1beta1 "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	batchv2alpha1 "k8s.io/client-go/kubernetes/typed/batch/v2alpha1"
	certificatesv1beta1 "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	coordinationv1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
	coordinationv1beta1 "k8s.io/client-go/kubernetes/typed/coordination/v1beta1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	discoveryv1alpha1 "k8s.io/client-go/kubernetes/typed/discovery/v1alpha1"
	discoveryv1beta1 "k8s.io/client-go/kubernetes/typed/discovery/v1beta1"
	eventsv1beta1 "k8s.io/client-go/kubernetes/typed/events/v1beta1"
	extensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	flowcontrolv1alpha1 "k8s.io/client-go/kubernetes/typed/flowcontrol/v1alpha1"
	networkingv1 "k8s.io/client-go/kubernetes/typed/networking/v1"
	networkingv1beta1 "k8s.io/client-go/kubernetes/typed/networking/v1beta1"
	nodev1alpha1 "k8s.io/client-go/kubernetes/typed/node/v1alpha1"
	nodev1beta1 "k8s.io/client-go/kubernetes/typed/node/v1beta1"
	policyv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	rbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	rbacv1alpha1 "k8s.io/client-go/kubernetes/typed/rbac/v1alpha1"
	rbacv1beta1 "k8s.io/client-go/kubernetes/typed/rbac/v1beta1"
	schedulingv1 "k8s.io/client-go/kubernetes/typed/scheduling/v1"
	schedulingv1alpha1 "k8s.io/client-go/kubernetes/typed/scheduling/v1alpha1"
	schedulingv1beta1 "k8s.io/client-go/kubernetes/typed/scheduling/v1beta1"
	settingsv1alpha1 "k8s.io/client-go/kubernetes/typed/settings/v1alpha1"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	storagev1alpha1 "k8s.io/client-go/kubernetes/typed/storage/v1alpha1"
	storagev1beta1 "k8s.io/client-go/kubernetes/typed/storage/v1beta1"

	terminusdiscovery "github.com/erda-project/erda/pkg/clientgo/discovery"
	terminusappsv1 "github.com/erda-project/erda/pkg/clientgo/kubernetes/apps/v1"
	terminuscorev1 "github.com/erda-project/erda/pkg/clientgo/kubernetes/core/v1"
	terminuspolicyv1beta1 "github.com/erda-project/erda/pkg/clientgo/kubernetes/policy/v1beta1"
)

type Clientset struct {
	*discovery.DiscoveryClient
	admissionregistrationV1      *admissionregistrationv1.AdmissionregistrationV1Client
	admissionregistrationV1beta1 *admissionregistrationv1beta1.AdmissionregistrationV1beta1Client
	appsV1                       *appsv1.AppsV1Client
	appsV1beta1                  *appsv1beta1.AppsV1beta1Client
	appsV1beta2                  *appsv1beta2.AppsV1beta2Client
	auditregistrationV1alpha1    *auditregistrationv1alpha1.AuditregistrationV1alpha1Client
	authenticationV1             *authenticationv1.AuthenticationV1Client
	authenticationV1beta1        *authenticationv1beta1.AuthenticationV1beta1Client
	authorizationV1              *authorizationv1.AuthorizationV1Client
	authorizationV1beta1         *authorizationv1beta1.AuthorizationV1beta1Client
	autoscalingV1                *autoscalingv1.AutoscalingV1Client
	autoscalingV2beta1           *autoscalingv2beta1.AutoscalingV2beta1Client
	autoscalingV2beta2           *autoscalingv2beta2.AutoscalingV2beta2Client
	batchV1                      *batchv1.BatchV1Client
	batchV1beta1                 *batchv1beta1.BatchV1beta1Client
	batchV2alpha1                *batchv2alpha1.BatchV2alpha1Client
	certificatesV1beta1          *certificatesv1beta1.CertificatesV1beta1Client
	coordinationV1beta1          *coordinationv1beta1.CoordinationV1beta1Client
	coordinationV1               *coordinationv1.CoordinationV1Client
	coreV1                       *corev1.CoreV1Client
	discoveryV1alpha1            *discoveryv1alpha1.DiscoveryV1alpha1Client
	discoveryV1beta1             *discoveryv1beta1.DiscoveryV1beta1Client
	eventsV1beta1                *eventsv1beta1.EventsV1beta1Client
	extensionsV1beta1            *extensionsv1beta1.ExtensionsV1beta1Client
	flowcontrolV1alpha1          *flowcontrolv1alpha1.FlowcontrolV1alpha1Client
	networkingV1                 *networkingv1.NetworkingV1Client
	networkingV1beta1            *networkingv1beta1.NetworkingV1beta1Client
	nodeV1alpha1                 *nodev1alpha1.NodeV1alpha1Client
	nodeV1beta1                  *nodev1beta1.NodeV1beta1Client
	policyV1beta1                *policyv1beta1.PolicyV1beta1Client
	rbacV1                       *rbacv1.RbacV1Client
	rbacV1beta1                  *rbacv1beta1.RbacV1beta1Client
	rbacV1alpha1                 *rbacv1alpha1.RbacV1alpha1Client
	schedulingV1alpha1           *schedulingv1alpha1.SchedulingV1alpha1Client
	schedulingV1beta1            *schedulingv1beta1.SchedulingV1beta1Client
	schedulingV1                 *schedulingv1.SchedulingV1Client
	settingsV1alpha1             *settingsv1alpha1.SettingsV1alpha1Client
	storageV1beta1               *storagev1beta1.StorageV1beta1Client
	storageV1                    *storagev1.StorageV1Client
	storageV1alpha1              *storagev1alpha1.StorageV1alpha1Client
}

// AdmissionregistrationV1 retrieves the AdmissionregistrationV1Client
func (c *Clientset) AdmissionregistrationV1() admissionregistrationv1.AdmissionregistrationV1Interface {
	return c.admissionregistrationV1
}

// AdmissionregistrationV1beta1 retrieves the AdmissionregistrationV1beta1Client
func (c *Clientset) AdmissionregistrationV1beta1() admissionregistrationv1beta1.AdmissionregistrationV1beta1Interface {
	return c.admissionregistrationV1beta1
}

// AppsV1 retrieves the AppsV1Client
func (c *Clientset) AppsV1() appsv1.AppsV1Interface {
	return c.appsV1
}

// AppsV1beta1 retrieves the AppsV1beta1Client
func (c *Clientset) AppsV1beta1() appsv1beta1.AppsV1beta1Interface {
	return c.appsV1beta1
}

// AppsV1beta2 retrieves the AppsV1beta2Client
func (c *Clientset) AppsV1beta2() appsv1beta2.AppsV1beta2Interface {
	return c.appsV1beta2
}

// AuditregistrationV1alpha1 retrieves the AuditregistrationV1alpha1Client
func (c *Clientset) AuditregistrationV1alpha1() auditregistrationv1alpha1.AuditregistrationV1alpha1Interface {
	return c.auditregistrationV1alpha1
}

// AuthenticationV1 retrieves the AuthenticationV1Client
func (c *Clientset) AuthenticationV1() authenticationv1.AuthenticationV1Interface {
	return c.authenticationV1
}

// AuthenticationV1beta1 retrieves the AuthenticationV1beta1Client
func (c *Clientset) AuthenticationV1beta1() authenticationv1beta1.AuthenticationV1beta1Interface {
	return c.authenticationV1beta1
}

// AuthorizationV1 retrieves the AuthorizationV1Client
func (c *Clientset) AuthorizationV1() authorizationv1.AuthorizationV1Interface {
	return c.authorizationV1
}

// AuthorizationV1beta1 retrieves the AuthorizationV1beta1Client
func (c *Clientset) AuthorizationV1beta1() authorizationv1beta1.AuthorizationV1beta1Interface {
	return c.authorizationV1beta1
}

// AutoscalingV1 retrieves the AutoscalingV1Client
func (c *Clientset) AutoscalingV1() autoscalingv1.AutoscalingV1Interface {
	return c.autoscalingV1
}

// AutoscalingV2beta1 retrieves the AutoscalingV2beta1Client
func (c *Clientset) AutoscalingV2beta1() autoscalingv2beta1.AutoscalingV2beta1Interface {
	return c.autoscalingV2beta1
}

// AutoscalingV2beta2 retrieves the AutoscalingV2beta2Client
func (c *Clientset) AutoscalingV2beta2() autoscalingv2beta2.AutoscalingV2beta2Interface {
	return c.autoscalingV2beta2
}

// BatchV1 retrieves the BatchV1Client
func (c *Clientset) BatchV1() batchv1.BatchV1Interface {
	return c.batchV1
}

// BatchV1beta1 retrieves the BatchV1beta1Client
func (c *Clientset) BatchV1beta1() batchv1beta1.BatchV1beta1Interface {
	return c.batchV1beta1
}

// BatchV2alpha1 retrieves the BatchV2alpha1Client
func (c *Clientset) BatchV2alpha1() batchv2alpha1.BatchV2alpha1Interface {
	return c.batchV2alpha1
}

// CertificatesV1beta1 retrieves the CertificatesV1beta1Client
func (c *Clientset) CertificatesV1beta1() certificatesv1beta1.CertificatesV1beta1Interface {
	return c.certificatesV1beta1
}

// CoordinationV1beta1 retrieves the CoordinationV1beta1Client
func (c *Clientset) CoordinationV1beta1() coordinationv1beta1.CoordinationV1beta1Interface {
	return c.coordinationV1beta1
}

// CoordinationV1 retrieves the CoordinationV1Client
func (c *Clientset) CoordinationV1() coordinationv1.CoordinationV1Interface {
	return c.coordinationV1
}

// CoreV1 retrieves the CoreV1Client
func (c *Clientset) CoreV1() corev1.CoreV1Interface {
	return c.coreV1
}

// DiscoveryV1alpha1 retrieves the DiscoveryV1alpha1Client
func (c *Clientset) DiscoveryV1alpha1() discoveryv1alpha1.DiscoveryV1alpha1Interface {
	return c.discoveryV1alpha1
}

// DiscoveryV1beta1 retrieves the DiscoveryV1beta1Client
func (c *Clientset) DiscoveryV1beta1() discoveryv1beta1.DiscoveryV1beta1Interface {
	return c.discoveryV1beta1
}

// EventsV1beta1 retrieves the EventsV1beta1Client
func (c *Clientset) EventsV1beta1() eventsv1beta1.EventsV1beta1Interface {
	return c.eventsV1beta1
}

// ExtensionsV1beta1 retrieves the ExtensionsV1beta1Client
func (c *Clientset) ExtensionsV1beta1() extensionsv1beta1.ExtensionsV1beta1Interface {
	return c.extensionsV1beta1
}

// FlowcontrolV1alpha1 retrieves the FlowcontrolV1alpha1Client
func (c *Clientset) FlowcontrolV1alpha1() flowcontrolv1alpha1.FlowcontrolV1alpha1Interface {
	return c.flowcontrolV1alpha1
}

// NetworkingV1 retrieves the NetworkingV1Client
func (c *Clientset) NetworkingV1() networkingv1.NetworkingV1Interface {
	return c.networkingV1
}

// NetworkingV1beta1 retrieves the NetworkingV1beta1Client
func (c *Clientset) NetworkingV1beta1() networkingv1beta1.NetworkingV1beta1Interface {
	return c.networkingV1beta1
}

// NodeV1alpha1 retrieves the NodeV1alpha1Client
func (c *Clientset) NodeV1alpha1() nodev1alpha1.NodeV1alpha1Interface {
	return c.nodeV1alpha1
}

// NodeV1beta1 retrieves the NodeV1beta1Client
func (c *Clientset) NodeV1beta1() nodev1beta1.NodeV1beta1Interface {
	return c.nodeV1beta1
}

// PolicyV1beta1 retrieves the PolicyV1beta1Client
func (c *Clientset) PolicyV1beta1() policyv1beta1.PolicyV1beta1Interface {
	return c.policyV1beta1
}

// RbacV1 retrieves the RbacV1Client
func (c *Clientset) RbacV1() rbacv1.RbacV1Interface {
	return c.rbacV1
}

// RbacV1beta1 retrieves the RbacV1beta1Client
func (c *Clientset) RbacV1beta1() rbacv1beta1.RbacV1beta1Interface {
	return c.rbacV1beta1
}

// RbacV1alpha1 retrieves the RbacV1alpha1Client
func (c *Clientset) RbacV1alpha1() rbacv1alpha1.RbacV1alpha1Interface {
	return c.rbacV1alpha1
}

// SchedulingV1alpha1 retrieves the SchedulingV1alpha1Client
func (c *Clientset) SchedulingV1alpha1() schedulingv1alpha1.SchedulingV1alpha1Interface {
	return c.schedulingV1alpha1
}

// SchedulingV1beta1 retrieves the SchedulingV1beta1Client
func (c *Clientset) SchedulingV1beta1() schedulingv1beta1.SchedulingV1beta1Interface {
	return c.schedulingV1beta1
}

// SchedulingV1 retrieves the SchedulingV1Client
func (c *Clientset) SchedulingV1() schedulingv1.SchedulingV1Interface {
	return c.schedulingV1
}

// SettingsV1alpha1 retrieves the SettingsV1alpha1Client
func (c *Clientset) SettingsV1alpha1() settingsv1alpha1.SettingsV1alpha1Interface {
	return c.settingsV1alpha1
}

// StorageV1beta1 retrieves the StorageV1beta1Client
func (c *Clientset) StorageV1beta1() storagev1beta1.StorageV1beta1Interface {
	return c.storageV1beta1
}

// StorageV1 retrieves the StorageV1Client
func (c *Clientset) StorageV1() storagev1.StorageV1Interface {
	return c.storageV1
}

// StorageV1alpha1 retrieves the StorageV1alpha1Client
func (c *Clientset) StorageV1alpha1() storagev1alpha1.StorageV1alpha1Interface {
	return c.storageV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

func NewKubernetesClientSet(addr string) (*Clientset, error) {
	var cs Clientset
	var err error
	cs.DiscoveryClient, err = terminusdiscovery.NewDiscoveryClient(addr)
	if err != nil {
		return nil, err
	}
	cs.coreV1, err = terminuscorev1.NewCoreClient(addr)
	if err != nil {
		return nil, err
	}
	cs.appsV1, err = terminusappsv1.NewAppClient(addr)
	if err != nil {
		return nil, err
	}
	cs.policyV1beta1, err = terminuspolicyv1beta1.NewPolicyClient(addr)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
