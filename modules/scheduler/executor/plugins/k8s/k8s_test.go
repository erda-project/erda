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

//import (
//	"bytes"
//	"context"
//	"encoding/json"
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/erda-project/erda/pkg/parser/diceyml"
//
//	"github.com/sirupsen/logrus"
//	"github.com/stretchr/testify/assert"
//	corev1 "k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/api/resource"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
//	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/cpupolicy"
//)
//
//var (
//	kubernetes executortypes.Executor
//	specObj    apistructs.ServiceGroup // self-defined runtime
//)
//
//func TestMain(m *testing.M) {
//	logrus.Infof("in testmain")
//	initK8S()
//	ret := m.Run()
//	os.Exit(ret)
//}
//
//func initK8S() {
//	specObj = apistructs.ServiceGroup{
//		Dice: apistructs.Dice{
//			ID:               "test-3456",
//			Type:             "service",
//			ProjectNamespace: "project-test-123",
//			Services: []apistructs.Service{
//				{
//					Name:     "fluentd",
//					Image:    "registry.cn-hangzhou.aliyuncs.com/terminus/fluentd:v2.5.2",
//					WorkLoad: ServicePerNode,
//					Resources: apistructs.Resources{
//						Cpu: 0.1,
//						Mem: 128,
//					},
//					Env: map[string]string{
//						"DICE_RUNTIME_ID": "11120",
//					},
//				},
//				{
//					Name:  "nginx",
//					Image: "registry.cn-hangzhou.aliyuncs.com/terminus/nginx:1.14.2",
//					Resources: apistructs.Resources{
//						Cpu: 0.1,
//						Mem: 64,
//					},
//					Scale: 1,
//					Env: map[string]string{
//						"DICE_RUNTIME_ID": "11120",
//					},
//					Ports: []diceyml.ServicePort{
//						{
//							Port:     80,
//							Protocol: "TCP",
//						},
//					},
//				},
//			},
//		},
//		ScheduleInfo2: apistructs.ScheduleInfo2{
//			IsUnLocked:   true,
//			HasOrg:       true,
//			Org:          "terminus",
//			HasWorkSpace: true,
//			WorkSpaces:   []string{"dev"},
//			Stateless:    true,
//		},
//	}
//
//	var err error
//	kubernetes, err = New("METRONOMEFORTERMINUSDEV", "terminus-dev", map[string]string{
//		"ADDR": "inet://ingress-nginx.kube-system.svc.cluster.local?direct=on&ssl=on/kubernetes.default.svc.cluster.local",
//	})
//	if err != nil {
//		logrus.Fatal(err)
//	}
//
//}
//
//func TestCreate(t *testing.T) {
//	ctx := context.Background()
//	_, err := kubernetes.Create(ctx, specObj)
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestDestroy(t *testing.T) {
//	ctx := context.Background()
//	err := kubernetes.Destroy(ctx, specObj)
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestStatus(t *testing.T) {
//	ctx := context.Background()
//	status, err := kubernetes.Status(ctx, specObj)
//	if err != nil {
//		t.Error(err)
//	}
//	logrus.Infof("deployment status: %+v", status)
//}
//
//func TestInspect(t *testing.T) {
//	ctx := context.Background()
//	runtime, err := kubernetes.Inspect(ctx, specObj)
//	if err != nil {
//		t.Error(err)
//	}
//	runtimeBytes, err := json.Marshal(runtime)
//	if err != nil {
//		t.Error(err)
//	}
//	var out bytes.Buffer
//	err = json.Indent(&out, runtimeBytes, "", "\t")
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Printf("runtime: %+v\n", out.String())
//}
//
//func TestUpdate(t *testing.T) {
//	ctx := context.Background()
//	specObj.Services[0].Resources.Cpu = 0.01
//	specObj.Services[1].Scale = 1
//	resp, err := kubernetes.Update(ctx, specObj)
//	if err != nil {
//		t.Error(err)
//	}
//	logrus.Infof("updated runtime: %+v", resp)
//}
//
//func TestX(t *testing.T) {
//	cpu := fmt.Sprintf("%.fm", 0.1*1000)
//	memory := fmt.Sprintf("%.fMi", 512.0)
//	assert.NotNil(t, cpu)
//	assert.NotNil(t, memory)
//}
//
//func TestSetFineGrainedCPU(t *testing.T) {
//	k := Kubernetes{
//		cpuNumQuota:       -1,
//		cpuSubscribeRatio: 2.0,
//	}
//	app1 := corev1.Container{
//		Resources: corev1.ResourceRequirements{
//			Requests: corev1.ResourceList{
//				corev1.ResourceCPU: resource.MustParse("0.1"),
//			},
//		},
//	}
//	extra1 := map[string]string{
//		"CPU_SUBSCRIBE_RATIO": "4.0",
//	}
//
//	err := k.SetFineGrainedCPU(&app1, extra1, 2)
//	assert.Nil(t, err)
//	// 0.1 的申请cpu 计算出的最大 cpu 为 0.2
//	assert.Equal(t, app1.Resources.Limits[corev1.ResourceCPU], resource.MustParse("0.2"))
//	// 实际分配的 cpu 是申请 cpu 除以超卖比，即 app1.Cpus / ratio = 0.1 / 4 = 0.025
//	assert.Equal(t, "25m", app1.Resources.Requests.Cpu().String())
//
//	app2 := corev1.Container{
//		Resources: corev1.ResourceRequirements{
//			Requests: corev1.ResourceList{
//				corev1.ResourceCPU: resource.MustParse("0.25"),
//			},
//		},
//	}
//
//	v := cpupolicy.AdjustCPUSize(0.25)
//	assert.Equal(t, v, 0.4)
//	err = k.SetFineGrainedCPU(&app2, nil, 2)
//	assert.Nil(t, err)
//	assert.Equal(t, app2.Resources.Limits[corev1.ResourceCPU], resource.MustParse("0.4"))
//	// map 为空，超卖比从集群配置中读出，为2.0
//	assert.Equal(t, "125m", app2.Resources.Requests.Cpu().String())
//
//	// quota 不限制的情况
//	k2 := Kubernetes{
//		cpuNumQuota: 0,
//	}
//
//	app3 := corev1.Container{
//		Resources: corev1.ResourceRequirements{
//			Requests: corev1.ResourceList{
//				corev1.ResourceCPU: resource.MustParse("0.5"),
//			},
//		},
//	}
//
//	err = k2.SetFineGrainedCPU(&app3, nil, 2)
//	assert.Nil(t, err)
//	assert.Equal(t, app3.Resources.Limits[corev1.ResourceCPU], resource.MustParse("0"))
//	// map 为空，并且集群中没有配置超卖比，即超卖比为 1
//	assert.Equal(t, "500m", app3.Resources.Requests.Cpu().String())
//
//	// 申请cpu值小于0.1的返回错误
//	app4 := corev1.Container{
//		Resources: corev1.ResourceRequirements{
//			Requests: corev1.ResourceList{
//				corev1.ResourceCPU: resource.MustParse("0.05"),
//			},
//		},
//	}
//	err = k.SetFineGrainedCPU(&app4, nil, 2)
//	assert.NotNil(t, err)
//
//	k3 := Kubernetes{
//		cpuSubscribeRatio: 2.5,
//	}
//	app5 := corev1.Container{
//		Resources: corev1.ResourceRequirements{
//			Requests: corev1.ResourceList{
//				corev1.ResourceCPU: resource.MustParse("1"),
//			},
//		},
//	}
//	extra5 := map[string]string{
//		"CPU_SUBSCRIBE_RATIO": "4.0",
//	}
//
//	err = k3.SetFineGrainedCPU(&app5, extra5, 2)
//	assert.Nil(t, err)
//	assert.Equal(t, "250m", app5.Resources.Requests.Cpu().String())
//}
