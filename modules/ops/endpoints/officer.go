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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/colonyutil"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) RegistryReadonly(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	s := vars["clusterName"]
	clusterinfo, err := e.bdl.QueryClusterInfo(s)
	if err != nil {
		errstr := fmt.Sprintf("RegistryReadonly queryclusterinfo err: %v, cluster: %v", err, s)
		logrus.Errorf(errstr)
		return mkResponseErr("400", errstr)
	}
	u := discover.Soldier()
	if clusterinfo.MustGet(apistructs.DICE_IS_EDGE) == "true" {
		u = clusterinfo.MustGetPublicURL("soldier")
	}
	var v apistructs.RegistryReadonlyResponse
	res, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(u).
		Path("/registry/readonly").
		Param("registryctlURL", ""). // omit this param
		Do().JSON(&v)
	if err != nil {
		errstr := fmt.Sprintf("call soldier failed: %v", err)
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	if res.StatusCode() != http.StatusOK {
		errstr := fmt.Sprintf("call soldier failed: statuscode: %d", res.StatusCode())
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	if !v.Success {
		errstr := fmt.Sprintf("call soldier failed: %v", v.Error.Msg)
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	return mkResponseData(v.Data)
}

func (e *Endpoints) RegistryRemoveManifests(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.RegistryManifestsRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return mkResponseErr("400", err.Error())
	}
	s := vars["clusterName"]
	clusterinfo, err := e.bdl.QueryClusterInfo(s)
	if err != nil {
		errstr := fmt.Sprintf("RegistryRemoveManifests queryclusterinfo err: %v, cluster: %v", err, s)
		logrus.Errorf(errstr)
		return mkResponseErr("400", errstr)
	}

	u := discover.Soldier()
	if clusterinfo.MustGet(apistructs.DICE_IS_EDGE) == "true" {
		u = clusterinfo.MustGetPublicURL("soldier")
	}
	req.RegistryURL = clusterinfo.MustGet(apistructs.REGISTRY_ADDR)
	var v apistructs.RegistryManifestsRemoveResponse
	res, err := httpclient.New().Post(u).Path("/registry/remove/manifests").JSONBody(req).Do().JSON(&v)
	if err != nil {
		errstr := fmt.Sprintf("RegistryRemoveManifests call soldier failed: %v", err)
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	if res.StatusCode() != http.StatusOK {
		errstr := fmt.Sprintf("call soldier failed: statuscode: %d", res.StatusCode())
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	if !v.Success {
		errstr := fmt.Sprintf("call soldier failed: %v", v.Error.Msg)
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	return mkResponseData(v.Data)
}

func (e *Endpoints) RegistryRemoveLayers(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	s := vars["clusterName"]
	clusterinfo, err := e.bdl.QueryClusterInfo(s)
	if err != nil {
		errstr := fmt.Sprintf("failed queryclusterinfo err: %v, cluster: %v", err, s)
		logrus.Errorf(errstr)
		return mkResponseErr("400", errstr)
	}
	u := discover.Soldier()
	if clusterinfo.MustGet(apistructs.DICE_IS_EDGE) == "true" {
		u = clusterinfo.MustGetPublicURL("soldier")
	}

	registryAppID := "/devops/registry"
	var v apistructs.RegistryReadonlyResponse // reuse
	res, err := httpclient.New().Post(u).
		Path("/registry/remove/layers").
		JSONBody(map[string]interface{}{
			"registryAppID": registryAppID}).
		Do().JSON(&v)
	if err != nil {
		errstr := fmt.Sprintf("failed to call soldier: %v", err)
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	if res.StatusCode() != http.StatusOK {
		errstr := fmt.Sprintf("call soldier failed: statuscode: %d", res.StatusCode())
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	if !v.Success {
		errstr := fmt.Sprintf("call soldier failed: %v", v.Error.Msg)
		logrus.Errorf(errstr)
		return mkResponseErr("502", errstr)
	}
	return mkResponseData(v.Data)
}

var scriptPath string = "/app/scripts.tar.gz"

// GetScriptInfo 获取脚本信息
func (e *Endpoints) GetScriptInfo(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	if !strutil.Contains(r.Header.Get("Client-ID"), "soldier") {
		return mkResponseErr("401", "unknown client access")
	}
	md5, err := colonyutil.CheckMd5(scriptPath)
	if err != nil {
		return mkResponseErr("500", err.Error())
	}
	pathList := strutil.Split(scriptPath, "/")
	fileName := pathList[len(pathList)-1]
	size, err := colonyutil.GetFileSize(scriptPath)
	if err != nil {
		return mkResponseErr("500", err.Error())
	}
	return mkResponseData(apistructs.ScriptInfo{
		Md5:             md5,
		Name:            fileName,
		Size:            size,
		ScriptBlackList: []string{},
	})
}

// ServeScript 提供脚本下载
func (e *Endpoints) ServeScript(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if !strutil.Contains(r.Header.Get("Client-ID"), "soldier") {
		colonyutil.WriteErr(w, "401", "unknown client access")
		return nil
	}
	name := vars["Name"]
	if name != colonyutil.GetFileName(scriptPath) {
		colonyutil.WriteErr(w, "401", "script filename mismatch!")
		return nil
	}
	http.ServeFile(w, r, scriptPath)
	return nil
}

func mkResponseErr(code, msg string) (httpserver.Responser, error) {
	return mkResponse(map[string]interface{}{
		"success": false,
		"err": map[string]interface{}{
			"code": code,
			"msg":  msg,
			"ctx":  nil,
		},
	})
}

func mkResponseData(v interface{}) (httpserver.Responser, error) {
	return mkResponse(map[string]interface{}{
		"success": true,
		"data":    v,
	})
}
