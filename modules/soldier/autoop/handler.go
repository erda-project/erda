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

package autoop

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5" // #nosec G501
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/colonyutil"
	"github.com/erda-project/erda/modules/soldier/auth"
	"github.com/erda-project/erda/modules/soldier/settings"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var (
	token   string
	lastMd5 string
)

func getToken() (string, error) {
	if token != "" {
		return token, nil
	}
	c := auth.Client{
		AccessTokenValiditySeconds:  433200,
		ClientId:                    "soldier",
		ClientName:                  "soldier",
		ClientSecret:                "soldier",
		RefreshTokenValiditySeconds: 433200,
		UserId:                      rand.Intn(100),
	}
	t, err := c.GetToken(5)
	if err != nil {
		return "", fmt.Errorf("auth failed: %s", err.Error())
	}
	logrus.Infof("get token succeed")
	token = t
	return token, nil
}

func resetToken(statusCode int) {
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		token = ""
	}
}

func download() (bool, error) {
	token, err := getToken()
	if err != nil {
		return false, err
	}
	var v apistructs.GetScriptInfoResponse
	res, err := httpclient.New().Get(settings.OpenAPIURL).Path("/api/script/info").Header("Authorization", token).Do().JSON(&v)
	if err != nil {
		return false, fmt.Errorf("get script info failed: %s", err.Error())
	}
	resetToken(res.StatusCode())
	if !res.IsOK() {
		return false, fmt.Errorf("get script info failed: status code is %d", res.StatusCode())
	}
	if !v.Success {
		return false, fmt.Errorf("get script info failed: %s %s", v.Error.Code, v.Error.Msg)
	}
	mutex.Lock()
	same := lastMd5 == v.Data.Md5
	mutex.Unlock()
	if same {
		return false, nil
	}
	var b bytes.Buffer
	res, err = httpclient.New().Get(settings.OpenAPIURL).Path("/api/script/"+v.Data.Name).Header("Authorization", token).Do().Body(&b)
	if err != nil {
		return false, fmt.Errorf("get script failed: %s", err.Error())
	}
	resetToken(res.StatusCode())
	if !res.IsOK() {
		return false, fmt.Errorf("get script failed: status code is %d", res.StatusCode())
	}
	// #nosec G401
	if a := md5.Sum(b.Bytes()); hex.EncodeToString(a[:]) != v.Data.Md5 || int64(b.Len()) != v.Data.Size {
		return false, fmt.Errorf("get script verify failed")
	}
	err = os.RemoveAll(newScriptsPath)
	if err != nil {
		return false, fmt.Errorf("remove new-scripts failed: %s", err.Error())
	}
	g, err := gzip.NewReader(&b)
	if err != nil {
		return false, fmt.Errorf("read gzip failed: %s", err.Error())
	}
	t := tar.NewReader(g)
	for {
		h, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, fmt.Errorf("read tar failed: %s", err.Error())
		}
		if h.Typeflag == tar.TypeDir {
			err = os.MkdirAll(filepath.Join(newScriptsPath, h.Name), h.FileInfo().Mode())
			if err != nil {
				return false, fmt.Errorf("mkdir failed: %s: %s", h.Name, err.Error())
			}
		} else {
			err = func() error {
				f, err := os.OpenFile(filepath.Join(newScriptsPath, h.Name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, h.FileInfo().Mode())
				if err != nil {
					return fmt.Errorf("open file failed: %s: %s", h.Name, err.Error())
				}
				defer f.Close()
				_, err = io.Copy(f, t)
				if err != nil {
					return fmt.Errorf("copy file failed: %s: %s", h.Name, err.Error())
				}
				return nil
			}()
			if err != nil {
				return false, err
			}
		}
	}
	mutex.Lock()
	lastMd5 = v.Data.Md5
	mutex.Unlock()
	return true, nil
}

var (
	clusterInfo *apistructs.ClusterInfo
	clusterTime = time.Now()
	diceCluster = os.Getenv("DICE_CLUSTER_NAME")
)

// TODO push
func readCluster() (*apistructs.ClusterInfo, error) {
	if clusterInfo != nil && time.Since(clusterTime) < time.Minute {
		return clusterInfo, nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	token, err := getToken()
	if err != nil {
		return nil, err
	}
	var v apistructs.GetClusterResponse
	res, err := httpclient.New().Get(settings.OpenAPIURL).Path("/api/clusters/read/"+diceCluster).Header("Authorization", token).Do().JSON(&v)
	if err != nil {
		return nil, fmt.Errorf("read cluster failed: %s", err.Error())
	}
	resetToken(res.StatusCode())
	if !res.IsOK() {
		return nil, fmt.Errorf("read cluster failed: status code is %d", res.StatusCode())
	}
	if !v.Success {
		return nil, fmt.Errorf("read cluster failed: %s %s", v.Error.Code, v.Error.Msg)
	}
	if v.Data.System == nil {
		return nil, fmt.Errorf("read cluster failed: system nil")
	}
	logrus.Infoln("read cluster succeed:", v.Data.Name)
	clusterInfo = &v.Data
	clusterTime = time.Now()
	return clusterInfo, nil
}

// CronActions download and load autoop script by hand
func CronActions(w http.ResponseWriter, r *http.Request) {
	reload := r.FormValue("reload") == "true"
	if err := Cron(reload); err != nil {
		colonyutil.WriteErr(w, "500", err.Error())
		return
	}
	mutex.Lock()
	n := len(actions)
	mutex.Unlock()
	colonyutil.WriteData(w, n)
}

// RunAction Manually run an automated operation script
func RunAction(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	a, ok := actions[name]
	if ok {
		m := NetdataEnv()
		for k := range a.Env {
			if v := r.FormValue("ENV_" + k); v != "" {
				m[k] = v
			}
		}
		if err := a.Run(m); err != nil {
			colonyutil.WriteErr(w, "500", err.Error())
			return
		}
	}
	colonyutil.WriteData(w, ok)
}

// CancelAction Cancels a running automation operation script
func CancelAction(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	a, ok := actions[name]
	if ok && a.CancelFunc != nil {
		a.CancelFunc()
	}
	colonyutil.WriteData(w, ok)
}
