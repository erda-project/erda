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

//// +build !default
//
package marathon

//
//import (
//	"context"
//	"encoding/json"
//	"os"
//	"strings"
//	"testing"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
//	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/cpupolicy"
//	"github.com/erda-project/erda/pkg/http/httpclient"
//
//	"github.com/stretchr/testify/assert"
//)
//
//var marathon executortypes.Executor
//var (
//	specObj     apistructs.ServiceGroup
//	specObjDice apistructs.ServiceGroup
//	specObjBlog apistructs.ServiceGroup
//)
//
//func TestMain(m *testing.M) {
//	initMarathon()
//	ret := m.Run()
//	os.Exit(ret)
//}
//
//func initMarathon() {
//	marathon = &Marathon{
//		name:   "MARATHONFORSERVICE",
//		prefix: "/zjt/runtimes/v1",
//		addr:   "dcos.test.terminus.io/service/marathon",
//		options: map[string]string{
//			"CONSTRAINS": "dice-role:UNLIKE:platform",
//		},
//		version: Ver{1, 6, 0},
//		client:  httpclient.New().BasicAuth("admin", "Terminus1234"),
//	}
//	specObj = apistructs.ServiceGroup{
//		Force: true,
//		Dice: apistructs.Dice{
//			Name:                 "dice-test",
//			Namespace:            "default",
//			ServiceDiscoveryKind: "VIP",
//			Services: []apistructs.Service{
//				{
//					Name: "web",
//					Resources: apistructs.Resources{
//						Cpu:  0.1,
//						Mem:  256,
//						Disk: 0,
//					},
//					Scale: 1,
//					Ports: []int{
//						8080,
//					},
//					Image: "docker-registry.registry.marathon.mesos:5000/org-default/dice-testweb-0229f3bada35f5e437ae0b4ddcbcb4b61524619410171",
//					HealthCheck: &apistructs.HealthCheck{
//						Kind: "TCP",
//						Port: 8080,
//					},
//					Env: map[string]string{
//						"APP_DIR":      "/web",
//						"TERMINUS_APP": "web",
//					},
//					Labels: map[string]string{
//						"HAPROXY_GROUP":   "external",
//						"HAPROXY_0_VHOST": "zjt-test.app.terminus.io",
//					},
//				},
//			},
//		},
//	}
//	specObjDice = apistructs.ServiceGroup{
//		Force: true,
//		Dice: apistructs.Dice{
//			Name:                 "dice",
//			Namespace:            "default",
//			ServiceDiscoveryKind: "VIP",
//			Services: []apistructs.Service{
//				{
//					Name: "ui",
//					Resources: apistructs.Resources{
//						Cpu:  0.1,
//						Mem:  256,
//						Disk: 0,
//					},
//					Scale: 1,
//					Ports: []int{
//						80,
//					},
//					Image: "docker-registry.registry.marathon.mesos:5000/org-default/diceui-0e63c9044c9cf8aa85b482161831a4311524808186712",
//					HealthCheck: &apistructs.HealthCheck{
//						Kind: "TCP",
//						Port: 80,
//					},
//					Env: map[string]string{
//						"CONSOLE_URL":          "http://${CONSOLE_HOST}:${CONSOLE_PORT}",
//						"ADDON_PLATFORM_URL":   "http://${ADDON_PLATFORM_HOST}:${ADDON_PLATFORM_PORT}",
//						"JOB_URL":              "http://${JOB_HOST}:${JOB_PORT}",
//						"MONITOR_URL":          "http://spotmonitor.marathon.l4lb.thisdcos.directory:8080",
//						"OFFICIAL_ADDONS_URL":  "http://${ADDONS_HOST}:${ADDONS_PORT}",
//						"PLATFORM_INSIGHT_URL": "http://${PLATFORM_INSIGHT_HOST}:${PLATFORM_INSIGHT_PORT}",
//						"SSO_URL":              "http://account0.app.terminus.io",
//					},
//					Labels: map[string]string{
//						"HAPROXY_GROUP":   "external",
//						"HAPROXY_0_VHOST": "ui.zjt-dice.app.terminus.io",
//					},
//					Depends: []string{
//						"console",
//						//"addon-platform",
//						//"addons",
//						//"job",
//					},
//				},
//				{
//					Name: "console",
//					Resources: apistructs.Resources{
//						Cpu:  0.1,
//						Mem:  512,
//						Disk: 0,
//					},
//					Scale: 1,
//					Ports: []int{
//						8080,
//					},
//					Image: "docker-registry.registry.marathon.mesos:5000/org-default/diceaddon-monitor_addon-platform_addons_console_job_platform-insight-3cfc40f613d5d3c8f197b86840c969ec1524808186712",
//					HealthCheck: &apistructs.HealthCheck{
//						Kind: "TCP",
//						Port: 8080,
//					},
//					Env: map[string]string{
//						"APP_DIR":                "/web",
//						"GITTAR_PACKER_MOBILE":   "18000000000",
//						"GITTAR_URL":             "http://analyzer:dice@dicegittar.marathon.l4lb.thisdcos.directory:5566",
//						"GITTAR_URL_FOR_CONSOLE": "http://gittar.app.terminus.io",
//						"MONITOR_URL":            "http://monitor-web.distracted-tepig.runtimes.marathon.l4lb.thisdcos.directory:8080",
//					},
//				},
//			},
//		},
//	}
//	specObjBlog = apistructs.ServiceGroup{
//		Force: true,
//		Extra: map[string]string{
//			"lastRestartTime": "123",
//		},
//		Dice: apistructs.Dice{
//			Name:                 "blog-test",
//			Namespace:            "default",
//			ServiceDiscoveryKind: "PROXY",
//			ServiceDiscoveryMode: "GLOBAL",
//			Labels: map[string]string{
//				"MATCH_TAGS":   "service-stateless",
//				"EXCLUDE_TAGS": "locked",
//			},
//			Services: []apistructs.Service{
//				{
//					Name: "showcase-Front",
//					Resources: apistructs.Resources{
//						Cpu:  0.1,
//						Mem:  128,
//						Disk: 0,
//					},
//					Scale: 1,
//					Ports: []int{
//						12300,
//						8080,
//					},
//					Image: "docker-registry.registry.marathon.mesos:5000/org-default/pampas-blogshowcase-front-27c4462ec7c3e5f194ae6d934874b5191527760214124",
//					HealthCheck: &apistructs.HealthCheck{
//						Kind: "TCP",
//						Port: 12300,
//					},
//					Env: map[string]string{
//						"BACKEND_URL": "http://${BLOG_WEB_HOST}:${BLOG_WEB_PORT}",
//					},
//					Labels: map[string]string{
//						"HAPROXY_GROUP":   "external",
//						"HAPROXY_0_VHOST": "zjt-test-blog.test.terminus.io",
//						"IS_ENDPOINT":     "true",
//					},
//					Depends: []string{
//						"blog-web",
//					},
//				},
//				{
//					Name: "blog-web",
//					Resources: apistructs.Resources{
//						Cpu:  0.1,
//						Mem:  384,
//						Disk: 0,
//					},
//					Scale: 1,
//					Ports: []int{
//						12300,
//					},
//					Image: "docker-registry.registry.marathon.mesos:5000/org-default/pampas-blogblog-service_user-service_blog-web-d0fda83887bef5c5f7322858e9905a571527760214121",
//					HealthCheck: &apistructs.HealthCheck{
//						Kind: "TCP",
//						Port: 12300,
//					},
//					Env: map[string]string{
//						"APP_DIR": "/blog-web",
//					},
//					Depends: []string{
//						"blog-service",
//						"user-service",
//					},
//					Labels: map[string]string{
//						"IS_ENDPOINT": "true",
//					},
//				},
//				{
//					Name: "blog-service",
//					Resources: apistructs.Resources{
//						Cpu:  0.1,
//						Mem:  512,
//						Disk: 0,
//					},
//					Scale: 1,
//					Ports: []int{
//						20880,
//					},
//					Image: "docker-registry.registry.marathon.mesos:5000/org-default/pampas-blogblog-service_user-service_blog-web-d0fda83887bef5c5f7322858e9905a571527760214121",
//					HealthCheck: &apistructs.HealthCheck{
//						Kind: "TCP",
//						Port: 20880,
//					},
//					Env: map[string]string{
//						"APP_DIR": "/blog-service/blog-service-impl",
//					},
//					Depends: []string{
//						"user-service",
//					},
//				},
//				{
//					Name: "user-service",
//					Resources: apistructs.Resources{
//						Cpu:  0.1,
//						Mem:  512,
//						Disk: 0,
//					},
//					Scale: 1,
//					Ports: []int{
//						20880,
//					},
//					Image: "docker-registry.registry.marathon.mesos:5000/org-default/pampas-blogblog-service_user-service_blog-web-d0fda83887bef5c5f7322858e9905a571527760214121",
//					HealthCheck: &apistructs.HealthCheck{
//						Kind: "TCP",
//						Port: 20880,
//					},
//					Env: map[string]string{
//						"APP_DIR": "/user-service/user-service-impl",
//					},
//				},
//			},
//		},
//	}
//	blogGlobalEnvs := map[string]string{
//		"MYSQL_HOST":        "a5b76982fe584dbb99cb75691fe420b1.88.mysql.addons.marathon.l4lb.thisdcos.directory",
//		"MYSQL_PORT":        "3306",
//		"MYSQL_DATABASE":    "blog",
//		"MYSQL_USERNAME":    "root",
//		"MYSQL_PASSWORD":    "5LdLMKp7BXhAE4f2",
//		"REDIS_HOST":        "464ee4073334409abb7052a7bcfec0c7.88.redis.addons.marathon.l4lb.thisdcos.directory",
//		"REDIS_PORT":        "6379",
//		"REDIS_PASSWORD":    "c07EbrChTCsvRuTK",
//		"ZOOKEEPER_HOST":    "38591d3a10b343f5aa3d2149b93b6477.88.zookeeper.addons.marathon.l4lb.thisdcos.directory",
//		"ZOOKEEPER_PORT":    "2181",
//		"JAVA_OPTS":         " -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=1617 -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false ",
//		"TERMINUS_APP_NAME": "PAMPAS_BLOG",
//		"TRACE_SAMPLE":      "1",
//	}
//	for i := range specObjBlog.Services {
//		for k, v := range blogGlobalEnvs {
//			specObjBlog.Services[i].Env[k] = v
//		}
//	}
//}
//
//func TestCreate(t *testing.T) {
//	ctx := context.Background()
//	_, err := marathon.Create(ctx, specObj)
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestCreatePvExplore(t *testing.T) {
//	ctx := context.Background()
//	str := `{
//  "executor": "MARATHONFORTERMINUS",
//  "version": "",
//  "name": "mysql-addon-fail",
//  "namespace": "test",
//  "services": [
//    {
//      "name": "mysql",
//      "image": "registry.cn-hangzhou.aliyuncs.com/terminus/dice-mysql:1.3.0",
//      "ports": [
//        3306
//      ],
//      "scale": 1,
//      "resources": {
//        "cpu": 0.01,
//        "mem": 512,
//        "disk": 0
//      },
//      "env": {
//        "MYSQL_DATABASE": "dice",
//        "MYSQL_ROOT_PASSWORD": "Hello1234"
//      },
//      "binds": [
//        {
//          "containerPath": "/var/lib/mysql",
//          "hostPath": "mysql-data"
//        },
//        {
//          "containerPath": "mysql-data",
//          "persistent": {
//            "type": "root",
//            "size": 1024
//          }
//        }
//      ],
//      "healthCheck": {
//        "Kind": "TCP",
//        "Port": 3306
//      }
//    }
//  ],
//  "serviceDiscoveryKind": "PROXY",
//  "serviceDiscoveryMode": "GLOBAL"
//}
//`
//
//	var obj apistructs.ServiceGroup
//	json.NewDecoder(strings.NewReader(str)).Decode(&obj)
//	//marathon.Destroy(ctx, obj)
//	_, err := marathon.Create(ctx, obj)
//	assert.NoError(t, err)
//}
//
//func TestDestroy(t *testing.T) {
//	ctx := context.Background()
//	err := marathon.Destroy(ctx, specObj)
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestUpdateSingle(t *testing.T) {
//	ctx := context.Background()
//	resp, err := marathon.Update(ctx, specObj)
//	if err != nil {
//		t.Error(err)
//	}
//	assert.Nil(t, resp)
//}
//
//func TestUpdateMulti(t *testing.T) {
//	ctx := context.Background()
//	resp, err := marathon.Update(ctx, specObjDice)
//	if err != nil {
//		t.Error(err)
//	}
//	assert.Nil(t, resp)
//}
//
//func TestCreateBlog(t *testing.T) {
//	ctx := context.Background()
//	resp, err := marathon.Create(ctx, specObjBlog)
//	assert.NoError(t, err)
//	assert.Nil(t, resp)
//}
//
//func TestUpdateBlog(t *testing.T) {
//	ctx := context.Background()
//	resp, err := marathon.Update(ctx, specObjBlog)
//	assert.NoError(t, err)
//	assert.Nil(t, resp)
//}
//
//func TestDestroyBlog(t *testing.T) {
//	ctx := context.Background()
//	err := marathon.Destroy(ctx, specObjBlog)
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestQueryStatusBlog(t *testing.T) {
//	ctx := context.Background()
//	status, err := marathon.Status(ctx, specObjBlog)
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(status)
//	assert.NotNil(t, status)
//}
//
//func TestInspectBlog(t *testing.T) {
//	ctx := context.Background()
//	ret, err := marathon.Inspect(ctx, specObjBlog)
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(ret)
//	t.Log(ret.(apistructs.ServiceGroup).StatusDesc)
//	for _, s := range ret.(apistructs.ServiceGroup).Services {
//		t.Log(s.Name, s.StatusDesc)
//	}
//	assert.NotNil(t, ret)
//}
//
//func TestExpandOneEnv(t *testing.T) {
//	ret := expandOneEnv("http://${CONSOLE_HOST}:${CONSOLE_PORT}", &map[string]string{
//		"CONSOLE_HOST": "localhost",
//		"CONSOLE_PORT": "8081",
//	})
//	assert.Equal(t, "http://localhost:8081", ret)
//}
//
//func TestParseMarathonVersion(t *testing.T) {
//	var ver, err = parseVersion("1.5.0")
//	assert.NoError(t, err)
//	assert.Equal(t, Ver{1, 5, 0}, ver)
//
//	ver, err = parseVersion("1.6.222")
//	assert.NoError(t, err)
//	assert.Equal(t, Ver{1, 6, 222}, ver)
//
//	ver, err = parseVersion("1.3")
//	assert.NoError(t, err)
//	assert.Equal(t, Ver{1, 3}, ver)
//
//	ver, err = parseVersion("1")
//	assert.NoError(t, err)
//	assert.Equal(t, Ver{1}, ver)
//
//	_, err = parseVersion("1..")
//	assert.Error(t, err)
//}
//
//func TestCompareVersion(t *testing.T) {
//	assert.True(t, lessThan(Ver{1, 3, 2}, Ver{1, 3, 3}))
//	assert.False(t, lessThan(Ver{1, 4, 0}, Ver{1, 4, 0}))
//	assert.False(t, lessThan(Ver{1, 4, 1}, Ver{1, 4, 0}))
//
//	assert.True(t, lessThan(Ver{1, 2}, Ver{1, 2, 1}))
//	assert.True(t, lessThan(Ver{1, 2, 9}, Ver{1, 3}))
//}
//
//func TestBuildMarathonGroupIdAndAppId(t *testing.T) {
//	assert.Equal(t, "/runtime/v1/kdjkla/thename", buildMarathonGroupId("/RunTime/V1", "kdjKla", "theName"))
//	assert.Equal(t, "/mygroupid/myservicename", buildMarathonAppId("/MyGroupId", "MyServiceName"))
//}
//
//func TestConvertPortToPortMapping(t *testing.T) {
//	mappings := convertPortToPortMapping([]int{8080, 8090, 9090}, "sticky.one", "test.vip.haha.com")
//	assert.Equal(t, []AppContainerPortMapping{
//		{
//			Labels: map[string]string{
//				"VIP_0": "sticky.one:8080",
//				"VIP_1": "test.vip.haha.com:8080",
//			},
//			Protocol:      "tcp",
//			ContainerPort: 8080,
//		},
//		{
//			Labels: map[string]string{
//				"VIP_2": "sticky.one:8090",
//				"VIP_3": "test.vip.haha.com:8090",
//			},
//			Protocol:      "tcp",
//			ContainerPort: 8090,
//		},
//		{
//			Labels: map[string]string{
//				"VIP_4": "sticky.one:9090",
//				"VIP_5": "test.vip.haha.com:9090",
//			},
//			Protocol:      "tcp",
//			ContainerPort: 9090,
//		},
//	}, mappings)
//}
//
//func TestConvertHealthCheck(t *testing.T) {
//	ahc, err := convertHealthCheck(apistructs.Service{
//		HealthCheck: &apistructs.HealthCheck{
//			Kind: "TCP",
//		},
//	}, Ver{1, 4, 7})
//	assert.NoError(t, err)
//	assert.Equal(t, AppHealthCheck{
//		GracePeriodSeconds:     0,
//		IntervalSeconds:        15,
//		MaxConsecutiveFailures: 9,
//		TimeoutSeconds:         10,
//		Protocol:               "TCP",
//	}, *ahc)
//
//	ahc2, err2 := convertHealthCheck(apistructs.Service{
//		HealthCheck: &apistructs.HealthCheck{
//			Kind: "TCP",
//		},
//	}, Ver{1, 6, 0})
//	assert.NoError(t, err2)
//	assert.Equal(t, AppHealthCheck{
//		GracePeriodSeconds:     0,
//		IntervalSeconds:        15,
//		MaxConsecutiveFailures: 9,
//		TimeoutSeconds:         10,
//		Protocol:               "MESOS_TCP",
//	}, *ahc2)
//}
//
//func TestParseAddHost(t *testing.T) {
//	assert.Nil(t, parseAddHost(nil))
//
//	assert.Nil(t, parseAddHost([]string{}))
//
//	assert.Equal(t, []AppContainerDockerParameter{
//		{"add-host", "dns.google.com:8.8.8.8"},
//	}, parseAddHost([]string{"8.8.8.8 dns.google.com"}))
//
//	assert.Equal(t, []AppContainerDockerParameter{
//		{"add-host", "baidu.com google.com:127.0.0.1"},
//	}, parseAddHost([]string{"127.0.0.1 baidu.com google.com"}))
//
//	assert.Equal(t, []AppContainerDockerParameter{
//		{"add-host", "dns.google.com:8.8.8.8"},
//		{"add-host", "baidu.com google.com:127.0.0.1"},
//	}, parseAddHost([]string{
//		"8.8.8.8 dns.google.com",
//		"127.0.0.1 baidu.com google.com"}))
//}
//
//func TestBuildVolumeCloudHostPath(t *testing.T) {
//	assert.Equal(t, "/netdata/volumes/haha/hehe", buildVolumeCloudHostPath("/netdata/volumes", "haha", "hehe"))
//	assert.Equal(t, "/netdata/volumes/haha/hehe", buildVolumeCloudHostPath("/netdata/volumes", "haha", "/hehe"))
//}
//
//func TestCancel(t *testing.T) {
//	marathon2 := &Marathon{
//		name: "MARATHONFORTERMINUSDEV",
//		addr: "https://dcos.dev.terminus.io/service/marathon",
//		options: map[string]string{
//			"ADDR":       "https://dcos.dev.terminus.io/service/marathon",
//			"CA_CRT":     "-----BEGIN CERTIFICATE-----\nMIIGATCCA+mgAwIBAgIJALoSf2TM2dHnMA0GCSqGSIb3DQEBCwUAMIGWMQswCQYD\nVQQGEwJDTjERMA8GA1UECAwIWkhFSklBTkcxETAPBgNVBAcMCEhBTkdaSE9VMREw\nDwYDVQQKDAhURVJNSU5VUzELMAkGA1UECwwCSVQxGDAWBgNVBAMMD2Rldi50ZXJt\naW51cy5pbzEnMCUGCSqGSIb3DQEJARYYamlhbmd0YW9AYWxpYmFiYS1pbmMuY29t\nMB4XDTE4MTIxMzAzMzE1MFoXDTI4MTIxMDAzMzE1MFowgZYxCzAJBgNVBAYTAkNO\nMREwDwYDVQQIDAhaSEVKSUFORzERMA8GA1UEBwwISEFOR1pIT1UxETAPBgNVBAoM\nCFRFUk1JTlVTMQswCQYDVQQLDAJJVDEYMBYGA1UEAwwPZGV2LnRlcm1pbnVzLmlv\nMScwJQYJKoZIhvcNAQkBFhhqaWFuZ3Rhb0BhbGliYWJhLWluYy5jb20wggIiMA0G\nCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDOSyoQrSzE+6m/WnSyqCtWabsjUyjg\naADlsC/JEAQFubq9hlb91xHl91rqOrgJAm65PyMvFNv7++vAYTf/ob7jcatZUTkq\n8AoYfvNpqoIOjO3TrVxJ2Rhrkyt0GnXO4S2dDB/MkyEIxigWJIvWbD5FEpQgxRsE\nWpbHO3RV5kfqWReK/LmwjC5Wm2OV3iPAvLAKPpFMvxCgkt0YcOOlF/U59q/DOM1y\ncNoyoY7qRT4PYE0i9hbvhkbi/Acq/MtPuWYaJoD8ZYOiBirdOy1pSZwI/Kqo4WME\nOagzbhNVfjklvfjvb6iOpv7x1Zlagf/8dv+iepplUPkg8DrxvI/Z5X+uugpjOnlw\nR/rd4NLgCDOHI3lgdzkkJZmVbHoXiZc/a5LZHsl1OE7eLIqcTSqT4h/X72jeH//T\nZlefz/b7PZ+VGB/e6rnEt3Nr1NVvWd2jYDuQszypOrL/VCbdYoG4PkvhAj+738Du\nlJV/XIspKsyjJ/dnjjxKnpBsit83Xv9tQyR2KCzjSDAA6tYHxsKrXROpm+6rsA+U\nndu9xVia4oFAUDodq9sPM7JRcMEGEHCYIJsv8LI7bK9lro2cM03N8p2F6VAn1fVV\ncLZsv07kzxPCo77gODweBHSBDk/iTr237Lo2+B1b/3km94cnontSWy+8HpNIkCIF\nZJ3G6VonulJhkQIDAQABo1AwTjAdBgNVHQ4EFgQUiRcvOWzwZJluTE15IeQr6CJj\nprowHwYDVR0jBBgwFoAUiRcvOWzwZJluTE15IeQr6CJjprowDAYDVR0TBAUwAwEB\n/zANBgkqhkiG9w0BAQsFAAOCAgEAJsWSjZSjM8L8dLDH+wJKVoS90K8sGf7E7F0L\ndHP+8ItaH+FC2Ur2S70msCZyy0bTrBQMVcqCqPckz19zESx9MwEokzyXK/CKVt+Z\nKqJy1gj+o7Q7mmoOItLtr8uk3AfU7uSqVmgcMEmEPZlbxJ+0rxWkn03YVvXw7I8y\n9pLGVBPlfBu5TB9GlFejJwHVJ2wc7hErzOOA1hG/Kv6GlesXUxwE36aaM3wfjNOI\nWGsvWPyp8h2hskZILYqqWur4KjUZqbdI/FZ5u/t+gPXcisEmKbJ13iwQ/t4nL+r7\n+m3FFKmH04pKWaBKFy/PPq9YLaPXygj67I+Pu4yf6fnQewPWz0/pxK8RXePeUUnC\nIF4AmAUAEkkADx/0niBtxzXOse7gLXGJFMDklyA/lgMeWEptD7rAN/DdaSLGG8Tf\nmhzu0cmjY9bHgvwBSBWAoI6My/b+XK046czl8qpGAahwmjwDIDjT3VRNWU8Oxy5c\nxj4NAEdJQRGMZWcuTz0basJcMtNyl7zOEQbqKXun4/NjMUk3zXC2PtwHonF5Fb27\nS/gf/WBOMZl4tWLZq3whK681nvhrUsTQkj+pBAhoGZbzAN1UrKCeECYoXbyrq1Ub\nsqp8cjlOaIx78qtVVY2o5ugUUTNLMoeyImsNmd+KR/XIuWum/26j0rffdA701VwQ\nizSO9v4=\n-----END CERTIFICATE-----",
//			"CLIENT_CRT": "-----BEGIN CERTIFICATE-----\nMIIEIzCCAgsCAQEwDQYJKoZIhvcNAQELBQAwgZYxCzAJBgNVBAYTAkNOMREwDwYD\nVQQIDAhaSEVKSUFORzERMA8GA1UEBwwISEFOR1pIT1UxETAPBgNVBAoMCFRFUk1J\nTlVTMQswCQYDVQQLDAJJVDEYMBYGA1UEAwwPZGV2LnRlcm1pbnVzLmlvMScwJQYJ\nKoZIhvcNAQkBFhhqaWFuZ3Rhb0BhbGliYWJhLWluYy5jb20wHhcNMTgxMjEzMDMz\nMTUwWhcNMjgxMjEwMDMzMTUwWjCBmzELMAkGA1UEBhMCQ04xETAPBgNVBAgMCFpI\nRUpJQU5HMREwDwYDVQQHDAhIQU5HWkhPVTERMA8GA1UECgwIVEVSTUlOVVMxCzAJ\nBgNVBAsMAklUMR0wGwYDVQQDDBRkY29zLmRldi50ZXJtaW51cy5pbzEnMCUGCSqG\nSIb3DQEJARYYamlhbmd0YW9AYWxpYmFiYS1pbmMuY29tMIGfMA0GCSqGSIb3DQEB\nAQUAA4GNADCBiQKBgQDNiCOwN8IHASgblSn3vBY9Ztb+IgUkg3n1kwAVwWfJRmaH\nB1mb3fjrQwnMSC/wSbtDY7uFTjVrWWOlIT9v7VNJhkHqc5WFSHHt1HQ3fWREsgg8\nrxBugwl6zlmuq3JE7Urw7BRs8QZa8wiOMKPX4nheyPpKgkk8b3Kj8ajEUiguhwID\nAQABMA0GCSqGSIb3DQEBCwUAA4ICAQAgU+amDrGftjOMdiS/fB5EnTT/eCfeRXLV\nTOleHQUKvgSz9eEus7yyH4JupT36fhaRDoApSrFspv5ZB1idkNvVcDhGfRCDq2cZ\nbS3pXzhaINvw42c3DjUBw8r9bwsD+xp7gY3wJWWwerxW/OXTP/fbShfdrbsdnEvY\nwdTAKFL9xQ/FZBaAqP+6z9l2fwR4mKUv4ADa6vcWJz5SI4TyQUv3JXADgregGwCr\nWo9XvBxH4oQPj1TgWDimHRuIJe8I43RKDkWv94RH96HXnARATwOHyiTuIhqDDTQZ\nQEjTOXI8rHyCG5a/SyH49At9PSAVokXAxvMYV9wF99E3HB+33vga/9RLBPKQWkbS\nA/kzo5LgrfVZ7Leh3/CmtTVSECweE/LGobbPBNZ4vE9j+CjoJwfG/CZm66Wk+Hl+\nAK0BTAdnIUO/SJl8RQWTvqcne7aW5Z2jqHl98QlYv9PA+RauNkkXkaBYT2mVYxSN\nxAN1FIOKE49k/wS6y244FJX/keIIxxuWaAD7RD07uvH5GxHio9Z77EEmBykMVDie\npbYPGNZL57V1nhvsdr2r70CI0TgxXewlqNNJmKzIG814C0yVwvCpk6V7JB5G9eK5\nlO66Fgxk6qdxfByJOKgG9ZCEq2Qtmt6N7Bkof108FK+7rjxZrdmJ9ube4kGu1ggP\nCEviIN00lA==\n-----END CERTIFICATE-----",
//			"CLIENT_KEY": "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDNiCOwN8IHASgblSn3vBY9Ztb+IgUkg3n1kwAVwWfJRmaHB1mb\n3fjrQwnMSC/wSbtDY7uFTjVrWWOlIT9v7VNJhkHqc5WFSHHt1HQ3fWREsgg8rxBu\ngwl6zlmuq3JE7Urw7BRs8QZa8wiOMKPX4nheyPpKgkk8b3Kj8ajEUiguhwIDAQAB\nAoGAfGaBS2CERN8TWpaPP04Nm/6J9Gm8+RvHDrd53rEgU3gUCHiPaUMSLbt2y7mJ\nooPOH3zW/FmZBa+mG0WjcuiPdqMIOohOayTnAKvBsE0JMSdW8v5+2Ku8ypG+bRDa\nB06/p/qVK935077tu72jCN9IkxVteSO7WEOaSqM75YeCJpkCQQD9cqPaIpuS14o3\ngxKk5LHcWO44AHl4Fk7mrHRemjIGgBAlv7TkF33eMGuyy4QqqV5xGK0FA/VB8Yfk\nMXWU6P5LAkEAz5n6LWhH3gChvqDLbltudUPz+yBAdQ8gh7Jl/XgS2EqMI5YZSAty\n7p2iQpT3WAAIzdoS1g0ptDP0JvsK5sj7NQJBAL36gdHQITeX81YbHQ2XE69sxdwa\nlvKqHiiQ2oXTJW5z7iatpcVXypSTTRdvsDleTZmO+pqI1f3BM7CcVlvxrjMCQHEB\nJfd1njkwKtszd8j4qCXY+YQnSC7wLwruhyn0JH3sBmCQoe5fnQ5abCrGH+WdDy3O\nmRY/UAYxiaN2X7bEjEkCQQCb5phFx0FVv/NJTKzNZ75nz2W1ejK8anC9epsQkjLX\n+kKMMjFr+MMXPhPrClSvl49OhoGxxMRarIy2XZ4dB4hK\n-----END RSA PRIVATE KEY-----",
//			"ENABLETAG":  "true",
//		},
//		client: httpclient.New().BasicAuth("admin", "Terminus1234"),
//	}
//
//	marathon2.client = httpclient.New(httpclient.WithHttpsCertFromJSON([]byte(marathon2.options["CLIENT_CRT"]),
//		[]byte(marathon2.options["CLIENT_KEY"]),
//		[]byte(marathon2.options["CA_CRT"])))
//
//	ctx := context.Background()
//	id, err := marathon2.Cancel(ctx, specObj)
//	assert.Nil(t, err)
//	// 当前 runtime 不存在，返回的 deployment ID 为空
//	assert.Equal(t, "", id)
//}
//
//func TestBuildVip(t *testing.T) {
//	oldRuntimeVip := buildMarathonStickyVipAppLevelPart(map[string]string{
//		"DICE_ORG_NAME":         "terminus",
//		"DICE_PROJECT_NAME":     "pmp",
//		"DICE_APPLICATION_NAME": "pmp",
//		"DICE_WORKSPACE":        "DEV",
//		"DICE_RUNTIME_NAME":     "DEV-feature/test",
//		"DICE_SERVICE_NAME":     "pmp-backend",
//	})
//	assert.Equal(t, "pmp-backend.devfeaturetest.pmp.pmp.terminus.runtimes", oldRuntimeVip)
//
//	vip := buildMarathonStickyVipAppLevelPart(map[string]string{
//		"DICE_ORG_NAME":         "terminus",
//		"DICE_PROJECT_NAME":     "pmp",
//		"DICE_APPLICATION_NAME": "pmp",
//		"DICE_WORKSPACE":        "DEV",
//		"DICE_RUNTIME_NAME":     "dev.feature.test",
//		"DICE_SERVICE_NAME":     "pmp-backend",
//	})
//	assert.Equal(t, "pmp-backend.devfeaturetest.pmp.pmp.terminus.runtimes", vip)
//}
//
//func TestSetFineGrainedCPU(t *testing.T) {
//	m := Marathon{
//		cpuNumQuota:       -1,
//		cpuSubscribeRatio: 2.0,
//	}
//	app1 := App{
//		Cpus: 0.1,
//	}
//	extra1 := map[string]string{
//		"CPU_SUBSCRIBE_RATIO": "4.0",
//	}
//
//	err := m.setFineGrainedCPU(&app1, extra1)
//	assert.Nil(t, err)
//	// 0.1 的申请cpu 计算出的最大 cpu 为 0.2
//	assert.Equal(t, app1.Container.Docker.Parameters, []AppContainerDockerParameter{{"cpu-quota", "20000"}})
//	// 实际分配的 cpu 是申请 cpu 除以超卖比，即 app1.Cpus / ratio = 0.1 / 4 = 0.025
//	assert.Equal(t, 0.025, app1.Cpus)
//
//	app2 := App{
//		Cpus: 0.25,
//	}
//	v := cpupolicy.AdjustCPUSize(0.25)
//	assert.Equal(t, v, 0.4)
//	err = m.setFineGrainedCPU(&app2, nil)
//	assert.Nil(t, err)
//	assert.Equal(t, app2.Container.Docker.Parameters, []AppContainerDockerParameter{{"cpu-quota", "40000"}})
//	// map 为空，超卖比从集群配置中读出，为2.0
//	assert.Equal(t, 0.125, app2.Cpus)
//
//	// quota 不限制的情况
//	m2 := Marathon{
//		cpuNumQuota: 0,
//	}
//	app3 := App{
//		Cpus: 0.5,
//	}
//	err = m2.setFineGrainedCPU(&app3, nil)
//	assert.Nil(t, err)
//	assert.Equal(t, app3.Container.Docker.Parameters, []AppContainerDockerParameter{{"cpu-quota", "0"}})
//	// map 为空，并且集群中没有配置超卖比，即超卖比为 1
//	assert.Equal(t, 0.5, app3.Cpus)
//
//	// 申请cpu值小于0.1的返回错误
//	app4 := App{
//		Cpus: 0.05,
//	}
//	err = m.setFineGrainedCPU(&app4, nil)
//	assert.NotNil(t, err)
//
//	m5 := Marathon{
//		cpuSubscribeRatio: 2.5,
//	}
//	app5 := App{
//		Cpus: 1,
//	}
//	extra5 := map[string]string{
//		"CPU_SUBSCRIBE_RATIO": "4.0",
//	}
//	err = m5.setFineGrainedCPU(&app5, extra5)
//	assert.Nil(t, err)
//	assert.Equal(t, 0.25, app5.Cpus)
//}
//
//func TestSetServiceDetailedResourceInfo(t *testing.T) {
//	service := &apistructs.Service{
//		Name: "xx",
//	}
//	queue := &Queue{
//		Queue: []QueueOffer{
//			{
//				Delay: QueueOfferDelay{
//					Overdue: true,
//				},
//				App: App{
//					Id: "X1",
//				},
//				ProcessedOffersSummary: ProcessedOffersSummary{
//					RejectSummaryLastOffers: []RejectSummaryLastOffer{{
//						Reason:    apistructs.INSUFFICIENTCPUS,
//						Processed: 7,
//						Declined:  7,
//					},
//					},
//				},
//			},
//			{
//				Delay: QueueOfferDelay{
//					Overdue: true,
//				},
//				App: App{
//					Id: "X2",
//				},
//				ProcessedOffersSummary: ProcessedOffersSummary{
//					RejectSummaryLastOffers: []RejectSummaryLastOffer{{
//						Reason:    apistructs.INSUFFICIENTMEMORY,
//						Processed: 5,
//						Declined:  6,
//					},
//					},
//				},
//			},
//			{
//				Delay: QueueOfferDelay{
//					Overdue: false,
//				},
//				App: App{
//					Id: "X3",
//				},
//				ProcessedOffersSummary: ProcessedOffersSummary{
//					RejectSummaryLastOffers: []RejectSummaryLastOffer{{
//						Reason:    apistructs.INSUFFICIENTMEMORY,
//						Processed: 4,
//						Declined:  4,
//					},
//					},
//				},
//			},
//		},
//	}
//	appID := "X1"
//	status := AppStatusWaiting
//
//	setServiceDetailedResourceInfo(service, queue, appID, status)
//	assert.Equal(t, true, service.StatusDesc.UnScheduledReasons.IsCPUInsufficient())
//	assert.Equal(t, "CPU资源不足", service.StatusDesc.UnScheduledReasons.String())
//
//	appID = "X2"
//	status = AppStatusWaiting
//	service2 := &apistructs.Service{
//		Name: "xx",
//	}
//	var emptyStatus apistructs.StatusDesc
//	setServiceDetailedResourceInfo(service2, queue, appID, status)
//	assert.Equal(t, emptyStatus, service2.StatusDesc)
//
//	appID = "X3"
//	status = AppStatusWaiting
//	service3 := &apistructs.Service{
//		Name: "xxyz",
//	}
//	setServiceDetailedResourceInfo(service3, queue, appID, status)
//	assert.Equal(t, emptyStatus, service3.StatusDesc)
//}
//
//func TestConstructConstrains1(t *testing.T) {
//	cons := constructConstrains(&apistructs.ScheduleInfo{
//		Location: [][]string{{"xxx", "yyy"}, {"zzz"}},
//	})
//
//	assert.Len(t, cons, 3)
//	cons2 := constructConstrains(&apistructs.ScheduleInfo{
//		IsPlatform: true,
//	})
//
//	assert.Len(t, cons2, 2)
//}
