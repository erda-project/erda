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

package mysql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon"
	mysqlv1 "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/mysql/v1"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders"
)

type MysqlOperator struct {
	k8s        addon.K8SUtil
	ns         addon.NamespaceUtil
	overcommit addon.OverCommitUtil
	secret     addon.SecretUtil
	pvc        addon.PVCUtil
	client     *httpclient.HTTPClient
}

func (my *MysqlOperator) Name(sg *apistructs.ServiceGroup) string {
	return "mysql-" + sg.ID[:10]
}
func (my *MysqlOperator) Namespace(sg *apistructs.ServiceGroup) string {
	if sg.ProjectNamespace != "" {
		return sg.ProjectNamespace
	}
	if len(sg.Services) > 0 {
		return sg.Services[0].Namespace
	}
	return ""
}
func (my *MysqlOperator) NamespacedName(sg *apistructs.ServiceGroup) string {
	return my.Namespace(sg) + "/" + my.Name(sg)
}

func New(k8s addon.K8SUtil, ns addon.NamespaceUtil, overcommit addon.OverCommitUtil,
	secret addon.SecretUtil, pvc addon.PVCUtil, client *httpclient.HTTPClient) *MysqlOperator {
	return &MysqlOperator{
		k8s:        k8s,
		ns:         ns,
		overcommit: overcommit,
		secret:     secret,
		pvc:        pvc,
		client:     client,
	}
}

func (my *MysqlOperator) IsSupported() bool {
	res, err := my.client.Get(my.k8s.GetK8SAddr()).
		Path("/apis/database.erda.cloud/v1").Do().RAW()
	if err == nil {
		defer res.Body.Close()
		var b []byte
		b, err = io.ReadAll(res.Body)
		if err == nil {
			return bytes.Contains(b, []byte("mysqls"))
		}
	}
	logrus.Errorf("failed to query /apis/database.erda.cloud/v1, host: %s, err: %v",
		my.k8s.GetK8SAddr(), err)
	return false
}

func (my *MysqlOperator) Validate(sg *apistructs.ServiceGroup) error {
	operator, ok := sg.Labels["USE_OPERATOR"]
	if !ok {
		return fmt.Errorf("[BUG] sg need USE_OPERATOR label")
	}
	if strings.ToLower(operator) != "mysql" {
		return fmt.Errorf("[BUG] value of label USE_OPERATOR should be 'mysql'")
	}
	if len(sg.Services) != 1 {
		return fmt.Errorf("illegal services num: %d", len(sg.Services))
	}
	if sg.Services[0].Name != "mysql" {
		return fmt.Errorf("illegal service: %s, should be 'mysql'", sg.Services[0].Name)
	}
	if sg.Services[0].Env["MYSQL_ROOT_PASSWORD"] == "" {
		return fmt.Errorf("illegal service: %s, need env 'MYSQL_ROOT_PASSWORD'", sg.Services[0].Name)
	}
	return nil
}

func (my *MysqlOperator) Convert(sg *apistructs.ServiceGroup) interface{} {
	mysql := sg.Services[0]

	scname := "dice-local-volume"
	capacity := "20Gi"

	// volumes in an addon service has same storageclass
	if len(mysql.Volumes) > 0 {
		if mysql.Volumes[0].SCVolume.Capacity >= diceyml.AddonVolumeSizeMin && mysql.Volumes[0].SCVolume.Capacity <= diceyml.AddonVolumeSizeMax {
			capacity = fmt.Sprintf("%dGi", mysql.Volumes[0].SCVolume.Capacity)
		}

		if mysql.Volumes[0].SCVolume.Capacity > diceyml.AddonVolumeSizeMax {
			capacity = fmt.Sprintf("%dGi", diceyml.AddonVolumeSizeMax)
		}

		if mysql.Volumes[0].SCVolume.StorageClassName != "" {
			scname = mysql.Volumes[0].SCVolume.StorageClassName
		}
	}

	scheinfo := sg.ScheduleInfo2
	scheinfo.Stateful = true
	affinity := constraintbuilders.K8S(&scheinfo, nil, nil, nil).Affinity

	replicas := mysql.Scale - 1
	if replicas < 0 {
		replicas = 0
	} else if replicas > 9 {
		replicas = 9
	}

	v := "v5.7"
	if mysql.Env["MYSQL_VERSION"] != "" {
		v = mysql.Env["MYSQL_VERSION"]
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
	}

	obj := &mysqlv1.Mysql{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "database.erda.cloud/v1",
			Kind:       "Mysql",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      my.Name(sg),
			Namespace: my.Namespace(sg),
		},
		Spec: mysqlv1.MysqlSpec{
			Version: v,

			PrimaryMode: mysqlv1.ModeClassic,
			Primaries:   1,
			Replicas:    pointer.Int(replicas),

			LocalUsername: "root",
			LocalPassword: mysql.Env["MYSQL_ROOT_PASSWORD"],

			StorageClassName: scname,
			StorageSize:      resource.MustParse(capacity),

			Affinity:    &affinity,
			Resources:   my.overcommit.ResourceOverCommit(mysql.Resources),
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}

	addon.SetAddonLabelsAndAnnotations(mysql, obj.Spec.Labels, obj.Spec.Annotations)

	if addonID, ok := mysql.Env["ADDON_ID"]; ok {
		obj.Spec.Labels["ADDON_ID"] = addonID
	}

	if clusterName, ok := mysql.Env[apistructs.DICE_CLUSTER_NAME.String()]; ok {
		obj.Spec.Labels[apistructs.DICE_CLUSTER_NAME.String()] = clusterName
	}

	return obj
}

func (my *MysqlOperator) Create(k8syml interface{}) error {
	obj, ok := k8syml.(*mysqlv1.Mysql)
	if !ok {
		return fmt.Errorf("[BUG] this k8syml should be *mysqlv1.Mysql")
	}
	if err := my.ns.Exists(obj.Namespace); err != nil {
		if err := my.ns.Create(obj.Namespace, nil); err != nil {
			return err
		}
	}
	var b bytes.Buffer
	resp, err := my.client.Post(my.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/database.erda.cloud/v1/namespaces/%s/mysqls", obj.Namespace)).
		JSONBody(obj).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to create mysql, %s/%s, err: %s, body: %s",
			obj.Namespace, obj.Name, err.Error(), b.String())
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to create mysql, %s/%s, statuscode: %d, body: %s",
			obj.Namespace, obj.Name, resp.StatusCode(), b.String())
	}
	return nil
}

func (my *MysqlOperator) Inspect(sg *apistructs.ServiceGroup) (*apistructs.ServiceGroup, error) {
	var b bytes.Buffer
	resp, err := my.client.Get(my.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/database.erda.cloud/v1/namespaces/%s/mysqls/%s", my.Namespace(sg), my.Name(sg))).
		Do().
		Body(&b)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect mysql, %s, err: %s",
			my.NamespacedName(sg), err.Error())
	}
	if !resp.IsOK() {
		return nil, fmt.Errorf("failed to inspect mysql, %s, statuscode: %d, body: %s",
			my.NamespacedName(sg), resp.StatusCode(), b.String())
	}
	obj := new(mysqlv1.Mysql)
	if err := json.NewDecoder(&b).Decode(obj); err != nil {
		return nil, err
	}

	mysqlsvc := &(sg.Services[0])

	if obj.Status.Color == mysqlv1.Green {
		mysqlsvc.Status = apistructs.StatusHealthy
		sg.Status = apistructs.StatusHealthy
	} else {
		mysqlsvc.Status = apistructs.StatusUnHealthy
		sg.Status = apistructs.StatusUnHealthy
	}

	mysqlsvc.Vip = strings.Join([]string{
		obj.BuildName("write"),
		obj.Namespace,
		"svc.cluster.local",
	}, ".")

	sg.Labels["PASSWORD"] = obj.Spec.LocalPassword

	return sg, nil
}

func (my *MysqlOperator) Remove(sg *apistructs.ServiceGroup) error {
	var b bytes.Buffer
	resp, err := my.client.Delete(my.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/database.erda.cloud/v1/namespaces/%s/mysqls/%s", my.Namespace(sg), my.Name(sg))).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to remove mysql, %s, err: %s", my.NamespacedName(sg), err.Error())
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to remove mysql, %s, statuscode: %d, body: %s",
			my.NamespacedName(sg), resp.StatusCode(), b.String())
	}

	//TODO remove pvc

	return nil
}

func (my *MysqlOperator) Update(k8syml interface{}) error {
	//TODO
	return fmt.Errorf("mysqloperator not impl Update yet")
}
