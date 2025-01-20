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

package canal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon"
	canalv1 "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/canal/v1"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders"
)

type CanalOperator struct {
	k8s        addon.K8SUtil
	ns         addon.NamespaceUtil
	overcommit addon.OverCommitUtil
	secret     addon.SecretUtil
	pvc        addon.PVCUtil
	client     *httpclient.HTTPClient
}

func (c *CanalOperator) Name(sg *apistructs.ServiceGroup) string {
	s := sg.Services[0].Env["CANAL_SERVER"]
	if s == "" {
		s = "canal-" + sg.ID[:10]
	}
	return s
}
func (c *CanalOperator) Namespace(sg *apistructs.ServiceGroup) string {
	return sg.ProjectNamespace
}
func (c *CanalOperator) NamespacedName(sg *apistructs.ServiceGroup) string {
	return c.Namespace(sg) + "/" + c.Name(sg)
}

func New(k8s addon.K8SUtil, ns addon.NamespaceUtil, overcommit addon.OverCommitUtil,
	secret addon.SecretUtil, pvc addon.PVCUtil, client *httpclient.HTTPClient) *CanalOperator {
	return &CanalOperator{
		k8s:        k8s,
		ns:         ns,
		overcommit: overcommit,
		secret:     secret,
		pvc:        pvc,
		client:     client,
	}
}

func (c *CanalOperator) IsSupported() bool {
	res, err := c.client.Get(c.k8s.GetK8SAddr()).
		Path("/apis/database.erda.cloud/v1").Do().RAW()
	if err == nil {
		defer res.Body.Close()
		var b []byte
		b, err = io.ReadAll(res.Body)
		if err == nil {
			return bytes.Contains(b, []byte("canals"))
		}
	}
	logrus.Errorf("failed to query /apis/database.erda.cloud/v1, host: %s, err: %v",
		c.k8s.GetK8SAddr(), err)
	return false
}

func (c *CanalOperator) Validate(sg *apistructs.ServiceGroup) error {
	operator, ok := sg.Labels["USE_OPERATOR"]
	if !ok {
		return fmt.Errorf("[BUG] sg need USE_OPERATOR label")
	}
	if strings.ToLower(operator) != "canal" {
		return fmt.Errorf("[BUG] value of label USE_OPERATOR should be 'canal'")
	}
	if len(sg.Services) != 1 {
		return fmt.Errorf("illegal services num: %d", len(sg.Services))
	}
	if sg.Services[0].Name != "canal" {
		return fmt.Errorf("illegal service: %s, should be 'canal'", sg.Services[0].Name)
	}

	if sg.Services[0].Env["CANAL_DESTINATION"] == "" {
		return fmt.Errorf("illegal service: %s, need env 'CANAL_DESTINATION'", sg.Services[0].Name)
	}

	if sg.Services[0].Env["canal.admin.manager"] != "" {
		if sg.Services[0].Env["admin.spring.datasource.address"] == "" {
			return fmt.Errorf("illegal service: %s, need env 'admin.spring.datasource.address'", sg.Services[0].Name)
		}
		// if sg.Services[0].Env["admin.spring.datasource.database"] == "" {
		// 	return fmt.Errorf("illegal service: %s, need env 'admin.spring.datasource.database'", sg.Services[0].Name)
		// }
		if sg.Services[0].Env["admin.spring.datasource.username"] == "" {
			return fmt.Errorf("illegal service: %s, need env 'admin.spring.datasource.username'", sg.Services[0].Name)
		}
		if sg.Services[0].Env["admin.spring.datasource.password"] == "" {
			return fmt.Errorf("illegal service: %s, need env 'admin.spring.datasource.password'", sg.Services[0].Name)
		}
	} else {
		if sg.Services[0].Env["canal.instance.master.address"] == "" {
			return fmt.Errorf("illegal service: %s, need env 'canal.instance.master.address'", sg.Services[0].Name)
		}
		if sg.Services[0].Env["canal.instance.dbUsername"] == "" {
			return fmt.Errorf("illegal service: %s, need env 'canal.instance.dbUsername'", sg.Services[0].Name)
		}
		if sg.Services[0].Env["canal.instance.dbPassword"] == "" {
			return fmt.Errorf("illegal service: %s, need env 'canal.instance.dbPassword'", sg.Services[0].Name)
		}
	}

	return nil
}

func (c *CanalOperator) Convert(sg *apistructs.ServiceGroup) (any, error) {
	canal := sg.Services[0]

	scheinfo := sg.ScheduleInfo2
	scheinfo.Stateful = true
	affinity := constraintbuilders.K8S(&scheinfo, nil, nil, nil).Affinity

	v := "v1.1.5"
	if canal.Env["CANAL_VERSION"] != "" {
		v = canal.Env["CANAL_VERSION"]
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
	}

	props := ""
	canalOptions := make(map[string]string)
	adminOptions := make(map[string]string)
	for k, v := range canal.Env {
		//TODO 确认所有选项如canal.instance./canal.mq.等入props，其余入canalOptions
		switch k {
		case "canal.zkServers", "canal.zookeeper.flush.period":
			canalOptions[k] = v
		default:
			if strings.HasPrefix(k, "canal.admin.") {
				canalOptions[k] = v
			} else if strings.HasPrefix(k, "admin.") {
				adminOptions[strings.TrimPrefix(k, "admin.")] = v
			} else if strings.HasPrefix(k, "canal.") {
				props += k + "=" + v + "\n"
			}
		}
	}

	workspace, _ := util.GetDiceWorkspaceFromEnvs(canal.Env)
	containerResources, err := c.overcommit.ResourceOverCommit(workspace, canal.Resources)
	if err != nil {
		return nil, fmt.Errorf("failed to calc container resources, err: %v", err)
	}
	adminContainerResources, err := c.overcommit.ResourceOverCommit(workspace, apistructs.Resources{
		Cpu: canal.Resources.Cpu / 3,
		Mem: canal.Resources.Mem * 2 / 3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to calc admin container resources, err: %v", err)
	}

	obj := &canalv1.Canal{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "database.erda.cloud/v1",
			Kind:       "Canal",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name(sg),
			Namespace: c.Namespace(sg),
		},
		Spec: canalv1.CanalSpec{
			Version: v,

			Replicas: canal.Scale,

			Affinity:       &affinity,
			Resources:      containerResources,
			AdminResources: adminContainerResources,
			Labels:         make(map[string]string),
			CanalOptions:   canalOptions,
			AdminOptions:   adminOptions,
			Annotations: map[string]string{
				"_destination": canal.Env["CANAL_DESTINATION"],
				"_props":       props,
			},
		},
	}

	addon.SetAddonLabelsAndAnnotations(canal, obj.Spec.Labels, obj.Spec.Annotations)

	return obj, nil
}

func (c *CanalOperator) Create(k8syml interface{}) error {
	obj, ok := k8syml.(*canalv1.Canal)
	if !ok {
		return fmt.Errorf("[BUG] this k8syml should be *canalv1.Canal")
	}
	if err := c.ns.Exists(obj.Namespace); err != nil {
		if err := c.ns.Create(obj.Namespace, nil); err != nil {
			return err
		}
	}

	destination := obj.Spec.Annotations["_destination"]
	props := obj.Spec.Annotations["_props"]
	delete(obj.Spec.Annotations, "_destination")
	delete(obj.Spec.Annotations, "_props")

	{ // canal cm
		var b bytes.Buffer
		res, err := c.client.Get(c.k8s.GetK8SAddr()).
			Path(fmt.Sprintf("/api/v1/namespaces/%s/configmaps/%s", obj.Namespace, obj.Name)).
			Do().
			Body(&b)
		if err != nil {
			return fmt.Errorf("failed to get canal cm, %s/%s, err: %s, body: %s",
				obj.Namespace, obj.Name, err.Error(), b.String())
		}
		if res.IsNotfound() {
			cm := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      obj.Name,
					Namespace: obj.Namespace,
				},
				Data: map[string]string{
					destination: props,
				},
			}
			{ //create cm
				var b bytes.Buffer
				res, err := c.client.Post(c.k8s.GetK8SAddr()).
					Path(fmt.Sprintf("/api/v1/namespaces/%s/configmaps", obj.Namespace)).
					JSONBody(cm).
					Do().
					Body(&b)
				if err != nil {
					return fmt.Errorf("failed to create canal cm, %s/%s, err: %s, body: %s",
						obj.Namespace, obj.Name, err.Error(), b.String())
				}
				if !res.IsOK() {
					return fmt.Errorf("failed to create canal cm, %s/%s, statuscode: %d, body: %s",
						obj.Namespace, obj.Name, res.StatusCode(), b.String())
				}
			}
		} else if !res.IsOK() {
			return fmt.Errorf("failed to get canal cm, %s/%s, statuscode: %d, body: %s",
				obj.Namespace, obj.Name, res.StatusCode(), b.String())
		} else {
			cm := &corev1.ConfigMap{}
			err := json.Unmarshal(b.Bytes(), cm)
			if err != nil {
				return fmt.Errorf("failed to unmarshal canal cm, %s/%s, err: %s, body: %s",
					obj.Namespace, obj.Name, err.Error(), b.String())
			}
			if cm.Data == nil {
				cm.Data = map[string]string{}
			}
			cm.Data[destination] = props

			{ //update cm
				var b bytes.Buffer
				res, err := c.client.Put(c.k8s.GetK8SAddr()).
					Path(fmt.Sprintf("/api/v1/namespaces/%s/configmaps/%s", obj.Namespace, obj.Name)).
					JSONBody(cm).
					Do().
					Body(&b)
				if err != nil {
					return fmt.Errorf("failed to update canal cm, %s/%s, err: %s, body: %s",
						obj.Namespace, obj.Name, err.Error(), b.String())
				}
				if !res.IsOK() {
					return fmt.Errorf("failed to update canal cm, %s/%s, statuscode: %d, body: %s",
						obj.Namespace, obj.Name, res.StatusCode(), b.String())
				}
			}
		}
	}

	{ // canal crd
		var b bytes.Buffer
		res, err := c.client.Get(c.k8s.GetK8SAddr()).
			Path(fmt.Sprintf("/apis/database.erda.cloud/v1/namespaces/%s/canals/%s", obj.Namespace, obj.Name)).
			Do().
			Body(&b)
		if err != nil {
			return fmt.Errorf("failed to get canal, %s/%s, err: %s, body: %s",
				obj.Namespace, obj.Name, err.Error(), b.String())
		}
		if res.IsNotfound() {
			var b bytes.Buffer
			res, err := c.client.Post(c.k8s.GetK8SAddr()).
				Path(fmt.Sprintf("/apis/database.erda.cloud/v1/namespaces/%s/canals", obj.Namespace)).
				JSONBody(obj).
				Do().
				Body(&b)
			if err != nil {
				return fmt.Errorf("failed to create canal, %s/%s, err: %s, body: %s",
					obj.Namespace, obj.Name, err.Error(), b.String())
			}
			if !res.IsOK() {
				return fmt.Errorf("failed to create canal, %s/%s, statuscode: %d, body: %s",
					obj.Namespace, obj.Name, res.StatusCode(), b.String())
			}
		} else if !res.IsOK() {
			return fmt.Errorf("failed to get canal, %s/%s, statuscode: %d, body: %s",
				obj.Namespace, obj.Name, res.StatusCode(), b.String())
		} else {
			//TODO: wait canal cm destination
			time.Sleep(5 * time.Second)
		}
	}
	return nil
}

func (c *CanalOperator) Inspect(sg *apistructs.ServiceGroup) (*apistructs.ServiceGroup, error) {
	var b bytes.Buffer
	res, err := c.client.Get(c.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/database.erda.cloud/v1/namespaces/%s/canals/%s", c.Namespace(sg), c.Name(sg))).
		Do().
		Body(&b)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect canal, %s, err: %s",
			c.NamespacedName(sg), err.Error())
	}
	if !res.IsOK() {
		return nil, fmt.Errorf("failed to inspect canal, %s, statuscode: %d, body: %s",
			c.NamespacedName(sg), res.StatusCode(), b.String())
	}
	obj := new(canalv1.Canal)
	if err := json.NewDecoder(&b).Decode(obj); err != nil {
		return nil, err
	}

	canalsvc := &(sg.Services[0])

	if obj.Status.Color == canalv1.Green {
		canalsvc.Status = apistructs.StatusHealthy
		sg.Status = apistructs.StatusHealthy
	} else {
		canalsvc.Status = apistructs.StatusUnHealthy
		sg.Status = apistructs.StatusUnHealthy
	}

	canalsvc.Vip = strings.Join([]string{
		obj.BuildName("x"),
		obj.Namespace,
		"svc.cluster.local",
	}, ".")
	if canalsvc.Env == nil {
		canalsvc.Env = make(map[string]string)
	}
	canalsvc.Env["CANAL_NAME"] = obj.Name
	canalsvc.Env["CANAL_NAMESPACE"] = obj.Namespace
	canalsvc.Env["CANAL_REPLICAS"] = strconv.Itoa(obj.Spec.Replicas)
	canalsvc.Env["CANAL_ADMIN_MANAGER"] = obj.Spec.CanalOptions["canal.admin.manager"]

	//TODO: check canal cm destination
	time.Sleep(5 * time.Second)

	return sg, nil
}

func (c *CanalOperator) Remove(sg *apistructs.ServiceGroup) error {
	canal := sg.Services[0]
	destination := canal.Env["CANAL_DESTINATION"]

	{ // canal cm
		var b bytes.Buffer
		res, err := c.client.Get(c.k8s.GetK8SAddr()).
			Path(fmt.Sprintf("/api/v1/namespaces/%s/configmaps/%s", c.Namespace(sg), c.Name(sg))).
			Do().
			Body(&b)
		if err != nil {
			return fmt.Errorf("failed to get canal cm, %s/%s, err: %s, body: %s",
				c.Namespace(sg), c.Name(sg), err.Error(), b.String())
		}
		if res.IsNotfound() {
			return nil
		} else if !res.IsOK() {
			return fmt.Errorf("failed to get canal cm, %s/%s, statuscode: %d, body: %s",
				c.Namespace(sg), c.Name(sg), res.StatusCode(), b.String())
		} else {
			cm := &corev1.ConfigMap{}
			err := json.Unmarshal(b.Bytes(), cm)
			if err != nil {
				return fmt.Errorf("failed to unmarshal canal cm, %s/%s, err: %s, body: %s",
					c.Namespace(sg), c.Name(sg), err.Error(), b.String())
			}

			if _, ok := cm.Data[destination]; !ok {
				return nil
			}
			delete(cm.Data, destination)

			var b bytes.Buffer
			res, err := c.client.Put(c.k8s.GetK8SAddr()).
				Path(fmt.Sprintf("/api/v1/namespaces/%s/configmaps/%s", c.Namespace(sg), c.Name(sg))).
				JSONBody(cm).
				Do().
				Body(&b)
			if err != nil {
				return fmt.Errorf("failed to update canal cm, %s/%s, err: %s, body: %s",
					c.Namespace(sg), c.Name(sg), err.Error(), b.String())
			}
			if !res.IsOK() {
				return fmt.Errorf("failed to update canal cm, %s/%s, statuscode: %d, body: %s",
					c.Namespace(sg), c.Name(sg), res.StatusCode(), b.String())
			}

			if len(cm.Data) > 0 {
				return nil
			}
		}
	}

	var b bytes.Buffer
	res, err := c.client.Delete(c.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/database.erda.cloud/v1/namespaces/%s/canals/%s", c.Namespace(sg), c.Name(sg))).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to remove canal, %s, err: %s",
			c.NamespacedName(sg), err.Error())
	}
	if !res.IsOK() {
		return fmt.Errorf("failed to remove canal, %s, statuscode: %d, body: %s",
			c.NamespacedName(sg), res.StatusCode(), b.String())
	}
	return nil
}

func (c *CanalOperator) Update(k8syml interface{}) error {
	//TODO
	return fmt.Errorf("canaloperator not impl Update yet")
}
