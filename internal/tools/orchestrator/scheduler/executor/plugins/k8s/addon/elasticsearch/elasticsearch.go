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

package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	commonv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/common/v1"
	elasticsearchv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/elasticsearch/v1"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sapi"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	esUsername      = "elastic"
	esExporterImage = "erda-registry.cn-hangzhou.cr.aliyuncs.com/retag/elasticsearch-exporter:v1.5.0"
)

type ElasticsearchOperator struct {
	k8s         addon.K8SUtil
	statefulset addon.StatefulsetUtil
	ns          addon.NamespaceUtil
	service     addon.ServiceUtil
	overcommit  addon.OvercommitUtil
	secret      addon.SecretUtil
	imageSecret addon.ImageSecretUtil
	client      *httpclient.HTTPClient
}

func New(k8s addon.K8SUtil,
	sts addon.StatefulsetUtil,
	ns addon.NamespaceUtil,
	service addon.ServiceUtil,
	overcommit addon.OvercommitUtil,
	secret addon.SecretUtil,
	imageSecret addon.ImageSecretUtil,
	client *httpclient.HTTPClient) *ElasticsearchOperator {
	return &ElasticsearchOperator{
		k8s:         k8s,
		statefulset: sts,
		ns:          ns,
		service:     service,
		overcommit:  overcommit,
		secret:      secret,
		imageSecret: imageSecret,
		client:      client,
	}
}

// IsSupported Determine whether to support  elasticseatch operator
func (eo *ElasticsearchOperator) IsSupported() bool {
	resp, err := eo.client.Get(eo.k8s.GetK8SAddr()).
		Path("/apis/elasticsearch.k8s.elastic.co/v1").
		Do().
		DiscardBody()
	if err != nil {
		logrus.Errorf("failed to query /apis/elasticsearch.k8s.elastic.co/v1, host: %v, err: %v",
			eo.k8s.GetK8SAddr(), err)
		return false
	}
	if !resp.IsOK() {
		return false
	}
	return true
}

// Validate Verify the legality of the ServiceGroup transformed from diceyml
func (eo *ElasticsearchOperator) Validate(sg *apistructs.ServiceGroup) error {
	operator, ok := sg.Labels["USE_OPERATOR"]
	if !ok {
		return fmt.Errorf("[BUG] sg need USE_OPERATOR label")
	}
	if strutil.ToLower(operator) != "elasticsearch" {
		return fmt.Errorf("[BUG] value of label USE_OPERATOR should be 'elasticsearch'")
	}
	if _, ok := sg.Labels["VERSION"]; !ok {
		return fmt.Errorf("[BUG] sg need VERSION label")
	}
	if _, err := convertJsontToMap(sg.Services[0].Env["config"]); err != nil {
		return fmt.Errorf("[BUG] config Abnormal format of ENV")
	}
	return nil
}

// Convert sg to cr, which is kubernetes yaml
func (eo *ElasticsearchOperator) Convert(sg *apistructs.ServiceGroup) interface{} {
	svc0 := sg.Services[0]
	scname := "dice-local-volume"
	// 官方建议，将堆内和堆外各设置一半, Xmx 和 Xms 设置的是堆内
	esMem := int(convertMiToMB(svc0.Resources.Mem) / 2)
	svc0.Env["ES_JAVA_OPTS"] = fmt.Sprintf("-Xms%dm -Xmx%dm", esMem, esMem)
	svc0.Env["ES_USER"] = "elastic"
	svc0.Env["ES_PASSWORD"] = svc0.Env["requirepass"]
	svc0.Env["SELF_PORT"] = "9200"

	scheinfo := sg.ScheduleInfo2
	scheinfo.Stateful = true
	affinity := constraintbuilders.K8S(&scheinfo, nil, nil, nil).Affinity.NodeAffinity

	nodeSets := eo.NodeSetsConvert(sg, scname, affinity)
	es := elasticsearchv1.Elasticsearch{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: "elasticsearch.k8s.elastic.co/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sg.ID,
			Namespace: genK8SNamespace(sg.Type, sg.ID),
		},
		Spec: elasticsearchv1.ElasticsearchSpec{
			HTTP: commonv1.HTTPConfig{
				TLS: commonv1.TLSOptions{
					SelfSignedCertificate: &commonv1.SelfSignedCertificate{
						Disabled: true,
					},
				},
			},
			Version: sg.Labels["VERSION"],
			Image:   svc0.Image,
			NodeSets: []elasticsearchv1.NodeSet{
				nodeSets,
			},
		},
	}

	// set Labels and annotations
	es.Labels = make(map[string]string)
	es.Annotations = make(map[string]string)
	addon.SetAddonLabelsAndAnnotations(svc0, es.Labels, es.Annotations)

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-es-elastic-user", sg.ID),
			Namespace: genK8SNamespace(sg.Type, sg.ID),
		},
		Data: map[string][]byte{
			"elastic": []byte(svc0.Env["requirepass"]),
		},
	}

	return ElasticsearchAndSecret{Elasticsearch: es, Secret: secret}
}

func (eo *ElasticsearchOperator) Create(k8syml interface{}) error {
	elasticsearchAndSecret, ok := k8syml.(ElasticsearchAndSecret)
	if !ok {
		return fmt.Errorf("[BUG] this k8syml should be elasticsearchAndSecret")
	}
	elasticsearch := elasticsearchAndSecret.Elasticsearch
	secret := elasticsearchAndSecret.Secret
	if err := eo.ns.Exists(elasticsearch.Namespace); err != nil {
		if err := eo.ns.Create(elasticsearch.Namespace, nil); err != nil {
			return err
		}
	}
	if err := eo.secret.CreateIfNotExist(&secret); err != nil {
		return err
	}
	//logrus.Info("es operator, start to create image secret, sepc:%+v", elasticsearch)
	if err := eo.imageSecret.NewImageSecret(elasticsearch.Namespace); err != nil {
		return err
	}
	var b bytes.Buffer
	resp, err := eo.client.Post(eo.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/elasticsearch.k8s.elastic.co/v1/namespaces/%s/elasticsearches", elasticsearch.Namespace)).
		JSONBody(elasticsearch).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to create elasticsearch, %s/%s, err: %v", elasticsearch.Namespace, elasticsearch.Name, err)
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to create elasticsearch, %s/%s, statuscode: %v, body: %v",
			elasticsearch.Namespace, elasticsearch.Name, resp.StatusCode(), b.String())
	}

	if err := eo.createMetricSvcIfNotExist(elasticsearch); err != nil {
		return err
	}
	return nil
}

func (eo *ElasticsearchOperator) Inspect(sg *apistructs.ServiceGroup) (*apistructs.ServiceGroup, error) {
	stslist, err := eo.statefulset.List(genK8SNamespace(sg.Type, sg.ID))
	if err != nil {
		return nil, err
	}
	svclist, err := eo.service.List(genK8SNamespace(sg.Type, sg.ID), nil)
	if err != nil {
		return nil, err
	}
	var elasticsearch *apistructs.Service
	if sg.Services[0].Name == "elasticsearch" {
		elasticsearch = &(sg.Services[0])
	}
	for _, sts := range stslist.Items {
		if sts.Spec.Replicas == nil {
			elasticsearch.Status = apistructs.StatusUnknown
		} else if *sts.Spec.Replicas == sts.Status.ReadyReplicas {
			elasticsearch.Status = apistructs.StatusHealthy
		} else {
			elasticsearch.Status = apistructs.StatusUnHealthy
		}
	}

	for _, svc := range svclist.Items {
		if strings.Contains(svc.Name, "es-http") {
			elasticsearch.Vip = strutil.Join([]string{svc.Name, svc.Namespace, "svc.cluster.local"}, ".")
		}
	}

	if elasticsearch.Status == apistructs.StatusHealthy {
		sg.Status = apistructs.StatusHealthy
	} else {
		sg.Status = apistructs.StatusUnHealthy
	}
	return sg, nil
}

func (eo *ElasticsearchOperator) Remove(sg *apistructs.ServiceGroup) error {
	k8snamespace := genK8SNamespace(sg.Type, sg.ID)
	var b bytes.Buffer
	resp, err := eo.client.Delete(eo.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/elasticsearch.k8s.elastic.co/v1/namespaces/%s/elasticsearches/%s", k8snamespace, sg.ID)).
		JSONBody(k8sapi.DeleteOptions).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to delele elasticsearch: %s/%s, err: %v", sg.Type, sg.ID, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil
		}
		return fmt.Errorf("failed to delete elasticsearch: %s/%s, statuscode: %v, body: %v",
			sg.Type, sg.ID, resp.StatusCode(), b.String())
	}

	if err := eo.ns.Delete(k8snamespace); err != nil {
		logrus.Errorf("failed to delete namespace: %s: %v", k8snamespace, err)
		return fmt.Errorf("failed to delete namespace: %s: %v", k8snamespace, err)
	}
	return nil
}

// Update
// secret The update will not be performed, and a restart is required due to the update of the static password. (You can improve the multi-user authentication through the user management machine with perfect service)
func (eo *ElasticsearchOperator) Update(k8syml interface{}) error {
	elasticsearchAndSecret, ok := k8syml.(ElasticsearchAndSecret)
	if !ok {
		return fmt.Errorf("[BUG] this k8syml should be elasticsearchAndSecret")
	}
	elasticsearch := elasticsearchAndSecret.Elasticsearch
	secret := elasticsearchAndSecret.Secret
	if err := eo.ns.Exists(elasticsearch.Namespace); err != nil {
		return err
	}
	if err := eo.secret.CreateIfNotExist(&secret); err != nil {
		return err
	}

	es, err := eo.Get(elasticsearch.Namespace, elasticsearch.Name)
	if err != nil {
		return err
	}
	// set last resource version.
	// fix error: "metadata.resourceVersion: Invalid value: 0x0: must be specified for an update"
	elasticsearch.ObjectMeta.ResourceVersion = es.ObjectMeta.ResourceVersion
	// set annotations controller-version
	// fix error: Resource was created with older version of operator, will not take action
	elasticsearch.ObjectMeta.Annotations = es.ObjectMeta.Annotations

	var b bytes.Buffer
	resp, err := eo.client.Put(eo.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/elasticsearch.k8s.elastic.co/v1/namespaces/%s/elasticsearches/%s", elasticsearch.Namespace, elasticsearch.Name)).
		JSONBody(elasticsearch).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to update elasticsearchs, %s/%s, err: %v", elasticsearch.Namespace, elasticsearch.Name, err)
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to update elasticsearchs, %s/%s, statuscode: %v, body: %v",
			elasticsearch.Namespace, elasticsearch.Name, resp.StatusCode(), b.String())
	}
	return nil
}

func genK8SNamespace(namespace, name string) string {
	return strutil.Concat(namespace, "--", name)
}

func (eo *ElasticsearchOperator) NodeSetsConvert(sg *apistructs.ServiceGroup, scname string, affinity *corev1.NodeAffinity) elasticsearchv1.NodeSet {
	// 默认 volume size
	capacity := "20Gi"
	// ES 只有一个 pvc，所以只取第一个 volume 的设置
	svc := sg.Services[0]
	if len(svc.Volumes) > 0 {
		if svc.Volumes[0].SCVolume.Capacity >= diceyml.AddonVolumeSizeMin && svc.Volumes[0].SCVolume.Capacity <= diceyml.AddonVolumeSizeMax {
			capacity = fmt.Sprintf("%dGi", svc.Volumes[0].SCVolume.Capacity)
		}

		if svc.Volumes[0].SCVolume.Capacity > diceyml.AddonVolumeSizeMax {
			capacity = fmt.Sprintf("%dGi", diceyml.AddonVolumeSizeMax)
		}

		if svc.Volumes[0].SCVolume.StorageClassName != "" {
			scname = svc.Volumes[0].SCVolume.StorageClassName
		}
	}
	labels := makeLabels(sg)

	esUri := fmt.Sprintf("--es.uri=http://%s:%s@localhost:9200", esUsername, svc.Env["requirepass"])
	config, _ := convertJsontToMap(svc.Env["config"])
	nodeSets := elasticsearchv1.NodeSet{
		Name:   "addon",
		Count:  int32(svc.Scale),
		Config: config,
		PodTemplate: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "es",
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Affinity: &corev1.Affinity{NodeAffinity: affinity},
				SecurityContext: &corev1.PodSecurityContext{
					FSGroup:      &[]int64{1000}[0],
					RunAsUser:    &[]int64{1000}[0],
					RunAsNonRoot: &[]bool{true}[0],
					RunAsGroup:   &[]int64{1000}[0],
				},
				Containers: []corev1.Container{
					{
						Name: "elasticsearch",
						Env:  envs(svc.Env),
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"cpu": resource.MustParse(
									fmt.Sprintf("%dm", int(1000*eo.overcommit.CPUOvercommit(svc.Resources.Cpu)))),
								"memory": resource.MustParse(
									fmt.Sprintf("%dMi", int(svc.Resources.Mem))),
							},
							Limits: corev1.ResourceList{
								"cpu": resource.MustParse(
									fmt.Sprintf("%dm", int(1000*svc.Resources.Cpu))),
								"memory": resource.MustParse(
									fmt.Sprintf("%dMi", int(svc.Resources.Mem))),
							},
						},
					}, {
						Name:    "es-exporter",
						Image:   esExporterImage,
						Command: []string{"/bin/elasticsearch_exporter", esUri},
						Ports: []corev1.ContainerPort{
							{
								Name:          "metrics",
								ContainerPort: 9114,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "ES_USERNAME",
								Value: "elastic",
							},
							{
								Name:  "ES_PASSWORD",
								Value: svc.Env["requirepass"],
							},
						},
					},
				},
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "elasticsearch-data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": resource.MustParse(capacity),
						},
					},
					StorageClassName: &scname,
				},
			},
		},
	}

	// set pvc annotations for snapshot
	if len(svc.Volumes) > 0 {
		if svc.Volumes[0].SCVolume.Snapshot.MaxHistory > 0 {
			if scname == apistructs.AlibabaSSDSC {
				nodeSets.VolumeClaimTemplates[0].Annotations = make(map[string]string)
				vs := diceyml.VolumeSnapshot{
					MaxHistory: svc.Volumes[0].SCVolume.Snapshot.MaxHistory,
				}
				vsMap := map[string]diceyml.VolumeSnapshot{}
				vsMap[scname] = vs
				data, _ := json.Marshal(vsMap)
				nodeSets.VolumeClaimTemplates[0].Annotations[apistructs.CSISnapshotMaxHistory] = string(data)
			} else {
				logrus.Warnf("Service %s pvc volume use storageclass %v, it do not support snapshot. Only volume.type SSD for Alibaba disk SSD support snapshot\n", svc.Name, scname)
			}
		}
	}

	return nodeSets
}

func envs(envs map[string]string) []corev1.EnvVar {
	r := []corev1.EnvVar{}
	for k, v := range envs {
		r = append(r, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	selfHost := apiv1.EnvVar{
		Name: "SELF_HOST",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}
	addonNodeID := apiv1.EnvVar{
		Name: "ADDON_NODE_ID",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}

	r = append(r, selfHost, addonNodeID)

	return r
}

// convertMiToMB Convert MiB to MB
// 1 MiB = 1.048576 MB
func convertMiToMB(mem float64) float64 {
	return mem * 1.048576
}

// convertJsontToMap Convert json to map
func convertJsontToMap(str string) (*commonv1.Config, error) {
	var tempMap commonv1.Config
	if str == "" {
		return &tempMap, nil
	}

	if err := json.Unmarshal([]byte(str), &tempMap); err != nil {
		return nil, err
	}
	return &tempMap, nil
}

// Get get elasticsearchs resource information
func (eo *ElasticsearchOperator) Get(namespace, name string) (*elasticsearchv1.Elasticsearch, error) {
	var b bytes.Buffer

	resp, err := eo.client.Get(eo.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/elasticsearch.k8s.elastic.co/v1/namespaces/%s/elasticsearches/%s", namespace, name)).
		Do().
		Body(&b)
	if err != nil {
		return nil, fmt.Errorf("failed to get elasticsearchs info: %s/%s, err: %v", namespace, name, err)
	}

	if !resp.IsOK() {
		if resp.IsNotfound() {
			return nil, fmt.Errorf("failed to get elasticsearchs info: %s/%s, err: %v", namespace, name, k8serror.ErrNotFound)
		}

		return nil, fmt.Errorf("failed to get elasticsearchs info: %s/%s, statusCode: %v, body: %v",
			namespace, name, resp.StatusCode(), b.String())
	}

	elasticsearch := &elasticsearchv1.Elasticsearch{}
	if err = json.NewDecoder(&b).Decode(elasticsearch); err != nil {
		return nil, fmt.Errorf("failed to get elasticsearchs info: %s/%s, err: %v", namespace, name, err)
	}

	return elasticsearch, nil
}

func (eo *ElasticsearchOperator) createMetricSvcIfNotExist(sg elasticsearchv1.Elasticsearch) error {
	svcName := fmt.Sprintf("es-exporter-%s", sg.Name)
	namespace := sg.Namespace
	_, err := eo.service.Get(namespace, svcName)
	if err != nil && err != k8serror.ErrNotFound {
		return err
	}
	metricSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: namespace,
			Labels: map[string]string{
				"common.k8s.elastic.co/type":                "elasticsearch",
				"elasticsearch.k8s.elastic.co/cluster-name": sg.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       9114,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(9114),
				},
			},
			Selector: map[string]string{
				"common.k8s.elastic.co/type":                "elasticsearch",
				"elasticsearch.k8s.elastic.co/cluster-name": sg.Name,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	if err = eo.service.Create(metricSvc); err != nil {
		return err
	}
	return nil
}

func makeLabels(sg *apistructs.ServiceGroup) map[string]string {
	svc := sg.Services[0]
	labels := map[string]string{}
	for k, v := range svc.Labels {
		labels[k] = v
	}
	for k, v := range svc.DeploymentLabels {
		labels[k] = v
	}
	for k, v := range sg.Labels {
		labels[k] = v
	}
	for k, v := range labels {
		if errs := validation.IsValidLabelValue(v); len(errs) > 0 {
			delete(labels, k)
		}
	}
	labels["ADDON_ID"] = sg.ID
	return labels
}
