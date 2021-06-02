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

package bundle

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// DownloadDiceFile 根据 uuid 返回文件流
func (b *Bundle) DownloadDiceFile(uuid string) (io.ReadCloser, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	respBody, resp, err := hc.Get(host).
		Path(fmt.Sprintf("/api/files/%s", uuid)).
		Header("Internal-Client", "bundle").
		Do().StreamBody()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		bodyBytes, _ := ioutil.ReadAll(respBody)
		var downloadResp apistructs.FileDownloadFailResponse
		if err := json.Unmarshal(bodyBytes, &downloadResp); err == nil {
			return nil, toAPIError(resp.StatusCode(), downloadResp.Error)
		}
		return nil, fmt.Errorf("failed to download dice file, uuid: %s, responseBody: %s", uuid, string(bodyBytes))
	}
	return respBody, nil
}

// DeleteDiceFile 根据 uuid 删除文件
func (b *Bundle) DeleteDiceFile(uuid string) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc
	respBody, resp, err := hc.Delete(host).
		Path(fmt.Sprintf("/api/files/%s", uuid)).
		Header("Internal-Client", "bundle").
		Do().StreamBody()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		bodyBytes, _ := ioutil.ReadAll(respBody)
		var downloadResp apistructs.FileDownloadFailResponse
		if err := json.Unmarshal(bodyBytes, &downloadResp); err == nil {
			return toAPIError(resp.StatusCode(), downloadResp.Error)
		}
		return fmt.Errorf("failed to download dice file, uuid: %s, responseBody: %s", uuid, string(bodyBytes))
	}
	return nil
}

// UploadFile 上传文件
func (b *Bundle) UploadFile(req apistructs.FileUploadRequest, clientTimeout ...int64) (*apistructs.File, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	multiparts := map[string]httpclient.MultipartItem{
		"file": {
			Reader:   req.FileReader,
			Filename: req.FileNameWithExt,
		},
	}
	if len(clientTimeout) <= 0 {
		clientTimeout = append(clientTimeout, 30)
	}
	request := httpclient.New(httpclient.WithCompleteRedirect(),
		httpclient.WithTimeout(5*time.Second, time.Duration(clientTimeout[0])*time.Second)).Post(host).
		Path("/api/files").
		Param("fileFrom", req.From).
		Param("public", strconv.FormatBool(req.IsPublic)).
		Param("encrypt", strconv.FormatBool(req.Encrypt)).
		Header(httputil.UserHeader, req.Creator).
		Header(httputil.InternalHeader, "bundle").
		MultipartFormDataBody(multiparts)
	if req.ExpiredAt != nil {
		request = request.Param("expiredIn", req.ExpiredAt.Sub(time.Now()).String())
	}
	var resp apistructs.FileUploadResponse
	httpResp, err := request.Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, fmt.Errorf("fail to upload file, status code: %d, body: %s", httpResp.StatusCode(), string(httpResp.Body()))
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error.Msg)
	}
	if resp.Data == nil || len(resp.Data.DownloadURL) <= 0 {
		return nil, fmt.Errorf("fail to upload file, no url in response")
	}
	return resp.Data, nil
}
