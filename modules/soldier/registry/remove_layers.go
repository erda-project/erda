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

package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pkg/colonyutil"
	"github.com/erda-project/erda/modules/pkg/dcos"
)

type RemoveLayersRequest struct {
	MasterIP      string `json:"masterID"`
	RegistryAppID string `json:"registryAppID"`
}

func (req *RemoveLayersRequest) setDefault() {
	if req.MasterIP == "" {
		req.MasterIP = "master.mesos"
	}
	if req.RegistryAppID == "" {
		req.RegistryAppID = "/devops/registry"
	}
}

func init() {
	p := os.Getenv("NETDATA_REGISTRY_PATH")
	if p == "" {
		logrus.Fatalln("no registry path")
	}
	err := os.Symlink(filepath.Join("/netdata", p), "/var/lib/registry")
	if err != nil {
		logrus.Fatalln(err)
	}
}

func RemoveLayers(w http.ResponseWriter, r *http.Request) {
	var req RemoveLayersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "400", err.Error())
		return
	}
	req.setDefault()

	mutex.Lock()
	defer mutex.Unlock()
	if cmd != nil {
		colonyutil.WriteData(w, false)
		return
	}

	if err := req.isRegistryHealthy(); err != nil {
		logrus.Errorln(err)
		colonyutil.WriteErr(w, "500", err.Error())
		return
	}

	cmd = exec.Command("registry", "garbage-collect", "/var/lib/registry/config-rw.yml", "--delete-untagged=true")
	go func() {
		defer func() {
			cmd = nil
		}()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		logrus.Infoln("start remove layers")
		if err := req.restartRegistry("ro"); err != nil {
			logrus.Errorln(err)
			return
		}
		defer func() {
			if err := req.restartRegistry("rw"); err != nil {
				logrus.Errorln(err)
				return
			}
		}()
		if err := cmd.Run(); err != nil {
			logrus.Errorln("remove layers failed:", err)
		} else {
			logrus.Infoln("remove layers succeed")
		}
	}()

	colonyutil.WriteData(w, true)
}

var (
	mutex sync.Mutex
	cmd   *exec.Cmd
)

func (req RemoveLayersRequest) isRegistryHealthy() error {
	app, err := dcos.GetApp(req.MasterIP, req.RegistryAppID)
	if err != nil {
		return err
	}
	if _, ok := app["healthChecks"]; ok {
		if tasksHealthy := int(app["tasksHealthy"].(float64)); tasksHealthy <= 0 {
			return fmt.Errorf("tasks healthy %d", tasksHealthy)
		}
	} else {
		if tasksRunning := int(app["tasksRunning"].(float64)); tasksRunning <= 0 {
			return fmt.Errorf("tasks running %d", tasksRunning)
		}
	}
	return nil
}

func (req RemoveLayersRequest) restartRegistry(s string) error {
	err := copyFile("/var/lib/registry/config.yml", "/var/lib/registry/config-"+s+".yml")
	if err != nil {
		return err
	}
	deploymentID, err := dcos.RestartApp(req.MasterIP, req.RegistryAppID)
	if err != nil {
		return err
	}
	logrus.Infoln("restart", s, req.RegistryAppID, deploymentID)
	for i := 1; i <= 10; i++ {
		time.Sleep(30 * time.Second)
		err = req.isRegistryHealthy()
		if err == nil {
			return nil
		}
		logrus.Warningf("restart %s sleep %d: %s\n", s, i, err.Error())
	}
	return err
}

func copyFile(dst, src string) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, b, 0644)
}
