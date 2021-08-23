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

package k8s

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/deployment"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/ingress"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8sservice"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/namespace"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/persistentvolumeclaim"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/statefulset"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

func TestSetBind(t *testing.T) {
	strs := []string{"x", "12", "y"}
	var dsts []string
	for i, str := range strs {
		name := strutil.Concat(str, "path-", strconv.Itoa(i))
		dsts = append(dsts, name)
	}
	assert.NotNil(t, dsts)

	val := "x${redis-master_HOST}y"
	v2 := strings.Replace(val, "${redis-master_HOST}", "x-0.x", 1)
	assert.Equal(t, v2, "xx-0.xy")
}

func TestParseSpecificEnv(t *testing.T) {
	annotations := map[string]string{
		"G0_ID":                     "terminus-zookeeper",
		"terminus-zookeeper-1":      "G0_N0",
		"terminus-zookeeper-2":      "G0_N1",
		"terminus-zookeeper-3":      "G0_N2",
		"TERMINUS_ZOOKEEPER_1_PORT": "2181",
		"TERMINUS_ZOOKEEPER_2_PORT": "2182",
		"TERMINUS_ZOOKEEPER_3_PORT": "2183",
		"K8S_NAMESPACE":             "addon-zookeeper--dc771ba61020d4d3fac0065f2498cb862",
	}
	val := "addr=${TERMINUS_ZOOKEEPER_1_HOST},${TERMINUS_ZOOKEEPER_2_HOST}:${TERMINUS_ZOOKEEPER_2_PORT}"
	newVal, ok := parseSpecificEnv(val, annotations)
	assert.True(t, ok)
	assert.Equal(t, newVal, "addr=terminus-zookeeper-0.terminus-zookeeper.addon-zookeeper--dc771ba61020d4d3fac0065f2498cb862.svc.cluster.local,"+
		"terminus-zookeeper-1.terminus-zookeeper.addon-zookeeper--dc771ba61020d4d3fac0065f2498cb862.svc.cluster.local:2182")
}

func TestReg(t *testing.T) {
	p := regexp.MustCompile(`.*\$\{([^}]*)\}.*`)
	v := p.MatchString("a${xy}b")
	assert.True(t, v)

	var re = regexp.MustCompile(`\$\{([^}]+?)\}`)
	results := re.FindAllString("key=${REDIS_HOST},${SLAVE_HOST},${TEST_HOST}", -1)
	assert.Equal(t, 3, len(results))
}

func TestStatefulset(t *testing.T) {
	sjson := string(`{
    "clusterName": "k8s-test",
    "labels": {
        "DICE_ADDON": "d1b404cdf6444c22be9a036f16f90a71",
        "DICE_ADDON_NAME": "eai_mq",
        "DICE_ORG_ID": "2",
		"GROUP_ID": "2",
        "DICE_ORG_NAME": "terminus",
        "DICE_PROJECT_ID": "174",
        "DICE_PROJECT_NAME": "FlowDesign",
        "DICE_SHARED_LEVEL": "PROJECT",
        "DICE_WORKSPACE": "dev",
	    "ADDON_GROUPS": "3",
        "SERVICE_TYPE": "ADDONS"
    },
    "name": "myrocket",
    "namespace": "myaddon-rocketmq",
    "serviceDiscoveryKind": "VIP",
    "serviceDiscoveryMode": "GLOBAL",
    "services": [
        {
            "binds": [
                {
                    "containerPath": "/opt/store",
                    "hostPath": "/netdata/addon/rocketmq/d1b404cdf6444c22be9a036f16f90a71/5e704c77de29458abe93b3563381492a/namesrv/store"
                },
                {
                    "containerPath": "/opt/logs",
                    "hostPath": "/netdata/addon/rocketmq/d1b404cdf6444c22be9a036f16f90a71/5e704c77de29458abe93b3563381492a/namesrv/logs"
                }
            ],
            "env": {
                "DICE_ADDON": "d1b404cdf6444c22be9a036f16f90a71",
                "DICE_ADDON_NAME": "eai_mq",
                "DICE_ORG_ID": "2",
                "DICE_ORG_NAME": "terminus",
                "DICE_PROJECT_ID": "174",
                "DICE_PROJECT_NAME": "FlowDesign",
                "DICE_SHARED_LEVEL": "PROJECT",
                "DICE_WORKSPACE": "dev",
                "JAVA_OPTS": "-Xms256m -Xmx256m",
                "SERVICE_TYPE": "ADDONS",
                "Xmn": "128m",
                "Xms": "256m",
                "Xmx": "256m"
            },
            "healthCheck": {
                "kind": "TCP",
                "path": "/",
                "port": 9876
            },
            "image": "registry.cn-hangzhou.aliyuncs.com/terminus/addon-rocketmq:namesrv-4.2.0.1",
            "labels": {
				"GROUP_ID": 111
			},
            "name": "rocketmq-namesrv-1",
            "ports": [
                9876
            ],
            "resources": {
                "cpu": 1,
                "disk": 0,
                "mem": 512
            },
            "scale": 1,
            "status": "",
            "unScheduledReasons": {},
            "vip": ""
        },
        {
            "binds": [
                {
                    "containerPath": "/opt/store",
                    "hostPath": "/netdata/addon/rocketmq/d1b404cdf6444c22be9a036f16f90a71/4e9a15bd4318401d9e3f836578d54eec/broker/store"
                },
                {
                    "containerPath": "/opt/logs",
                    "hostPath": "/netdata/addon/rocketmq/d1b404cdf6444c22be9a036f16f90a71/4e9a15bd4318401d9e3f836578d54eec/broker/logs"
                }
            ],
            "depends": [
                "rocketmq-namesrv-1"
            ],
            "env": {
                "BROKER_NAME": "rocketmq-broker-1",
                "DICE_ADDON": "d1b404cdf6444c22be9a036f16f90a71",
                "DICE_ADDON_NAME": "eai_mq",
                "DICE_ORG_ID": "2",
                "DICE_ORG_NAME": "terminus",
                "DICE_PROJECT_ID": "174",
                "DICE_PROJECT_NAME": "FlowDesign",
                "DICE_SHARED_LEVEL": "PROJECT",
                "DICE_WORKSPACE": "dev",
                "JAVA_OPTS": "-Xms1433m -Xmx1433m",
                "NAMESRV_ADDR": "${ROCKETMQ_NAMESRV_1_HOST}:${ROCKETMQ_NAMESRV_1_PORT}",
                "SERVICE_TYPE": "ADDONS",
                "Xmn": "716m",
                "Xms": "1433m",
                "Xmx": "1433m"
            },
            "healthCheck": {
                "kind": "TCP",
                "path": "/",
                "port": 10909
            },
            "image": "registry.cn-hangzhou.aliyuncs.com/terminus/addon-rocketmq:broker-4.2.0.1",
            "labels": {},
            "name": "rocketmq-broker-1",
            "ports": [
                10909,
                10911
            ],
            "resources": {
                "cpu": 1,
                "disk": 0,
                "mem": 1024
            },
            "scale": 1,
            "status": "",
            "unScheduledReasons": {},
            "vip": ""
        },
        {
            "depends": [
                "rocketmq-namesrv-1"
            ],
            "env": {
                "DICE_ADDON": "d1b404cdf6444c22be9a036f16f90a71",
                "DICE_ADDON_NAME": "eai_mq",
                "DICE_ORG_ID": "2",
                "DICE_ORG_NAME": "terminus",
                "DICE_PROJECT_ID": "174",
                "DICE_PROJECT_NAME": "FlowDesign",
                "DICE_SHARED_LEVEL": "PROJECT",
                "DICE_WORKSPACE": "dev",
                "JAVA_OPTS": "-Xms256m -Xmx256m -Drocketmq.namesrv.addr=${ROCKETMQ_NAMESRV_1_HOST}:${ROCKETMQ_NAMESRV_1_PORT} -Dcom.rocketmq.sendMessageWithVIPChannel=false",
                "NAMESRV_ADDR": "${ROCKETMQ_NAMESRV_1_HOST}:${ROCKETMQ_NAMESRV_1_PORT}",
                "SERVICE_TYPE": "ADDONS",
                "Xmn": "128m",
                "Xms": "256m",
                "Xmx": "256m"
            },
            "healthCheck": {
                "kind": "TCP",
                "path": "/",
                "port": 8080
            },
            "image": "registry.cn-hangzhou.aliyuncs.com/terminus/addon-rocketmq:console-1.0.0",
            "labels": {
                "HAPROXY_0_VHOST": "rocketmq-console-1-d1b404cdf6444c22be9a036f16f90a71.app.terminus.io",
                "HAPROXY_GROUP": "external",
				"GROUP_ID": 111
            },
            "name": "rocketmq-console-1",
            "ports": [
                8080
            ],
            "resources": {
                "cpu": 1,
                "disk": 0,
                "mem": 512
            },
            "scale": 1,
            "status": "",
            "unScheduledReasons": {},
            "vip": ""
        }
    ],
    "status": "Created",
    "unScheduledReasons": {}
}`)

	sg := &apistructs.ServiceGroup{}
	err := json.Unmarshal([]byte(sjson), sg)
	assert.NotNil(t, err)

	kubernetes := &Kubernetes{
		name:      "whatever",
		addr:      "10.167.0.248:8080",
		client:    httpclient.New(),
		deploy:    deployment.New(deployment.WithCompleteParams("10.167.0.248:8080", httpclient.New())),
		ingress:   ingress.New(ingress.WithCompleteParams("10.167.0.248:8080", httpclient.New())),
		namespace: namespace.New(namespace.WithCompleteParams("10.167.0.248:8080", httpclient.New())),
		service:   k8sservice.New(k8sservice.WithCompleteParams("10.167.0.248:8080", httpclient.New())),
		pvc:       persistentvolumeclaim.New(persistentvolumeclaim.WithCompleteParams("10.167.0.248:8080", httpclient.New())),
		sts:       statefulset.New(statefulset.WithCompleteParams("10.167.0.248:8080", httpclient.New())),
	}

	//_, err = kubernetes.Create(context.Background(), *sg)
	//assert.Nil(t, err)

	_, err = kubernetes.Inspect(context.Background(), *sg)
	assert.NotNil(t, err)
}

func TestParseJobSpecTemplate(t *testing.T) {
	var hostPath = "{{.MOUNTPOINT_PATH}}"

	clusterInfo := map[string]string{
		"MOUNTPOINT_PATH": "/netdata",
		"ETCD_ENDPOINTS":  "http://10.0.6.198:2379",
	}

	newPath, err := ParseJobHostBindTemplate(hostPath, clusterInfo)
	assert.Nil(t, err)

	assert.Equal(t, clusterInfo["MOUNTPOINT_PATH"], newPath)
}
