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

package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pkg/colonyutil"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type RemoveManifestsRequest struct {
	RegistryURL string   `json:"registryURL"`
	Images      []string `json:"images"`
}

type RemoveManifestsResponse struct {
	Succeed []string          `json:"succeed"`
	Failed  map[string]string `json:"failed"`
}

func (req RemoveManifestsRequest) removeManifests(s string) string {
	var name, tag string
	if i := strings.IndexByte(s, '/'); i != -1 {
		name = s[i+1:]
		if i := strings.LastIndexByte(name, ':'); i != -1 {
			name, tag = name[:i], name[i+1:]
		}
	}
	if name == "" {
		return "image name is empty"
	}
	if tag == "" {
		tag = "latest"
	}
	c := httpclient.New()
	res, err := c.Get(req.RegistryURL).Path(fmt.Sprintf("/v2/%s/manifests/%s", name, tag)).
		Header("Content-Type", "application/json").
		Header("Accept", "application/vnd.docker.distribution.manifest.v2+json").
		Do().DiscardBody()
	if err != nil {
		return "get manifests failed: " + err.Error()
	}
	if sc := res.StatusCode(); sc != http.StatusOK {
		if sc == http.StatusNotFound { // does not exist
			return ""
		} else {
			return "get manifests failed: status code is " + strconv.Itoa(sc)
		}
	}
	dcd := res.ResponseHeader("Docker-Content-Digest")
	if dcd == "" {
		return "get manifests failed: header Docker-Content-Digest is empty"
	}
	res, err = c.Delete(req.RegistryURL).Path(fmt.Sprintf("/v2/%s/manifests/%s", name, dcd)).
		Header("Content-Type", "application/json").
		Do().DiscardBody()
	if err != nil {
		return "delete manifests failed: " + err.Error()
	}
	if sc := res.StatusCode(); sc != http.StatusOK && sc != http.StatusAccepted {
		return "delete manifests failed: status code is " + strconv.Itoa(sc)
	}
	return ""
}

func RemoveManifests(w http.ResponseWriter, r *http.Request) {
	var req RemoveManifestsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		colonyutil.WriteErr(w, "400", err.Error())
		return
	}
	if req.RegistryURL == "" {
		req.RegistryURL = os.Getenv("REGISTRY_ADDR")
	}
	if req.RegistryURL == "" {
		colonyutil.WriteErr(w, "400", "no registry url")
		return
	}

	m := make(map[string]string, len(req.Images))
	for _, image := range req.Images {
		if _, ok := m[image]; !ok {
			m[image] = req.removeManifests(image)
		}
	}
	var res RemoveManifestsResponse
	for k, v := range m {
		if v == "" {
			logrus.Infof("%s: delete image succeed\n", k)
			res.Succeed = append(res.Succeed, k)
		} else {
			if res.Failed == nil {
				res.Failed = make(map[string]string)
			}
			logrus.Warningf("%s: %s\n", k, v)
			res.Failed[k] = v
		}
	}
	colonyutil.WriteData(w, res)
}

type State struct {
	Running   bool      `json:"running"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	LastError string    `json:"lastError"`
}

type StateReply struct {
	Success bool   `json:"success"`
	Data    *State `json:"data"`
	Err     *struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Ctx  interface{} `json:"ctx"`
	} `json:"err"`
}

func Readonly(w http.ResponseWriter, r *http.Request) {
	registryctlURL := r.FormValue("registryctlURL")
	if registryctlURL == "" {
		registryctlURL = os.Getenv("REGISTRY_ADDR")
		if registryctlURL != "" {
			i := strings.LastIndexByte(registryctlURL, ':')
			if i != -1 {
				registryctlURL = registryctlURL[0:i]
			}
			registryctlURL = registryctlURL + ":5050"
		}
	}
	if registryctlURL == "" {
		colonyutil.WriteErr(w, "400", "no registryctl url")
		return
	}

	var v StateReply
	c := httpclient.New()
	res, err := c.Get(registryctlURL).Path("/gc").Do().JSON(&v)
	if err != nil {
		colonyutil.WriteErr(w, "500", "get gc state failed: "+err.Error())
		return
	}
	if sc := res.StatusCode(); sc != http.StatusOK {
		colonyutil.WriteErr(w, "500", "get gc state failed: status code is "+strconv.Itoa(sc))
		return
	}
	if !v.Success {
		colonyutil.WriteErr(w, v.Err.Code, "get gc state failed: "+v.Err.Msg)
		return
	}
	colonyutil.WriteData(w, v.Data.Running)
}
