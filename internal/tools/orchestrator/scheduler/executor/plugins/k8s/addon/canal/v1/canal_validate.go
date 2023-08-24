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

package v1

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func (r *Canal) Default() {
	if r.Spec.Version == "" {
		r.Spec.Version = "v1.1.5"
	}

	if r.Spec.Replicas == 0 {
		r.Spec.Replicas = 1
	}

	if r.Spec.CanalOptions == nil {
		r.Spec.CanalOptions = make(map[string]string)
	}
	if r.Spec.CanalOptions["canal.admin.manager"] != "" {
		//TODO：admin和server各自独立部署，在admin中配置server

		// if r.Spec.AdminOptions["spring.datasource.address"] == "" {
		// 	r.Spec.AdminOptions["spring.datasource.address"] = "127.0.0.1:3306"
		// }
		if r.Spec.AdminOptions["spring.datasource.database"] == "" {
			r.Spec.AdminOptions["spring.datasource.database"] = "canal_manager"
		}
		// if r.Spec.AdminOptions["spring.datasource.username"] == "" {
		// 	r.Spec.AdminOptions["spring.datasource.username"] = "canal"
		// }
		// if r.Spec.AdminOptions["spring.datasource.password"] == "" {
		// 	r.Spec.AdminOptions["spring.datasource.password"] = "canal"
		// }
		if r.Spec.AdminOptions["canal.adminUser"] == "" {
			r.Spec.AdminOptions["canal.adminUser"] = "admin"
		}
		if r.Spec.AdminOptions["canal.adminPasswd"] == "" {
			r.Spec.AdminOptions["canal.adminPasswd"] = "admin"
		}

		if r.Spec.CanalOptions["canal.admin.port"] == "" {
			r.Spec.CanalOptions["canal.admin.port"] = "11110"
		}
		if r.Spec.CanalOptions["canal.admin.user"] == "" {
			r.Spec.CanalOptions["canal.admin.user"] = r.Spec.AdminOptions["canal.adminUser"]
		}
		if r.Spec.CanalOptions["canal.admin.passwd"] == "" {
			r.Spec.CanalOptions["canal.admin.passwd"] = r.Spec.AdminOptions["canal.adminPasswd"]
		}
	} else {
		if r.Spec.CanalOptions["canal.port"] == "" {
			r.Spec.CanalOptions["canal.port"] = "11111"
		}
		if r.Spec.CanalOptions["canal.metrics.pull.port"] == "" {
			r.Spec.CanalOptions["canal.metrics.pull.port"] = "11112"
		}

		if r.Spec.CanalOptions["canal.auto.scan"] == "" {
			r.Spec.CanalOptions["canal.auto.scan"] = "true"
		}
		if _, ok := r.Spec.CanalOptions["canal.destinations"]; !ok {
			r.Spec.CanalOptions["canal.destinations"] = ""
		}

		if r.Spec.CanalOptions["canal.instance.gtidon"] == "" {
			r.Spec.CanalOptions["canal.instance.gtidon"] = "true"
		}
		if r.Spec.CanalOptions["canal.instance.connectionCharset"] == "" {
			r.Spec.CanalOptions["canal.instance.connectionCharset"] = "UTF-8"
		}

		if r.Spec.Replicas > 1 {
			if r.Spec.CanalOptions["canal.instance.global.spring.xml"] == "" {
				r.Spec.CanalOptions["canal.instance.global.spring.xml"] = "classpath:spring/default-instance.xml"
			}
		}
	}

	if r.Spec.Image == "" {
		r.Spec.Image = "registry.erda.cloud/erda-addons/canal:" + strings.TrimPrefix(r.Spec.Version, "v")
	}
	if r.Spec.ImagePullPolicy == "" {
		r.Spec.ImagePullPolicy = corev1.PullIfNotPresent
	}
}

func Between(i, min, max int) bool {
	return min <= i && i <= max
}

func (r *Canal) Validate() (err error) {
	if !strings.HasPrefix(r.Spec.Version, "v") {
		return fmt.Errorf("version invalid: %s", r.Spec.Version)
	}

	if !Between(r.Spec.Replicas, 1, 9) {
		return fmt.Errorf("replicas not in [1, 9]: %d", r.Spec.Replicas)
	}

	if len(r.Spec.CanalOptions) == 0 {
		return fmt.Errorf("canal properties required")
	}
	if r.Spec.CanalOptions["canal.admin.manager"] != "" {
		if r.Spec.CanalOptions["canal.admin.manager"] != "127.0.0.1:8089" {
			return fmt.Errorf("canal.admin.manager must 127.0.0.1:8089")
		}

		if r.Spec.AdminOptions["spring.datasource.address"] == "" {
			return fmt.Errorf("spring.datasource.address required")
		}
		// if r.Spec.AdminOptions["spring.datasource.database"] == "" {
		// 	return fmt.Errorf("spring.datasource.database required")
		// }
		if r.Spec.AdminOptions["spring.datasource.username"] == "" {
			return fmt.Errorf("spring.datasource.username required")
		}
		if r.Spec.AdminOptions["spring.datasource.password"] == "" {
			return fmt.Errorf("spring.datasource.password required")
		}

		if r.Spec.AdminOptions["canal.adminUser"] != r.Spec.CanalOptions["canal.admin.user"] {
			return fmt.Errorf("canal.adminUser/canal.admin.user mismatch")
		}
		if r.Spec.AdminOptions["canal.adminPasswd"] != r.Spec.CanalOptions["canal.admin.passwd"] {
			return fmt.Errorf("canal.adminPasswd/canal.admin.passwd mismatch")
		}
	} else {
		if r.Spec.CanalOptions["canal.auto.scan"] == "true" {
			if r.Spec.CanalOptions["canal.destinations"] != "" {
				return fmt.Errorf("canal.destinations not required")
			}
		} else if r.Spec.CanalOptions["canal.auto.scan"] == "false" {
			if r.Spec.CanalOptions["canal.destinations"] == "" {
				return fmt.Errorf("canal.destinations required")
			}
		} else {
			return fmt.Errorf("canal.auto.scan invalid")
		}

		// if strings.Contains(r.Spec.CanalOptions["canal.destinations"], ",") {
		// 	return fmt.Errorf("multi canal.destinations unsupported")
		// }
		// if r.Spec.CanalOptions["canal.instance.master.address"] == "" {
		// 	return fmt.Errorf("canal.instance.master.address required")
		// }
		// if r.Spec.CanalOptions["canal.instance.dbUsername"] == "" {
		// 	return fmt.Errorf("canal.instance.dbUsername required")
		// }
		// if r.Spec.CanalOptions["canal.instance.dbPassword"] == "" {
		// 	return fmt.Errorf("canal.instance.dbPassword required")
		// }

		//canal.instance.rds.accesskey
		//canal.instance.rds.secretkey
		//canal.instance.rds.instanceId

		//canal.instance.filter.regex
		//canal.instance.filter.black.regex

		//canal.instance.master.journal.name
		//canal.instance.master.position

		if r.Spec.Replicas > 1 {
			if r.Spec.CanalOptions["canal.zkServers"] == "" {
				return fmt.Errorf("canal.zkServers required")
			}
		}
	}

	return nil
}
