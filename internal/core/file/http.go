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

package file

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	infrahttpserver "github.com/erda-project/erda-infra/providers/httpserver"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/file/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/file/filetypes"
	"github.com/erda-project/erda/internal/core/legacy/services/apierrors"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

type HTTPHandler interface {
	UploadFile(rw http.ResponseWriter, r *http.Request)
	DownloadFile(rw http.ResponseWriter, r *http.Request)
	HeadFile(rw http.ResponseWriter, r *http.Request)
	DeleteFile(rw http.ResponseWriter, r *http.Request)
}

func (p *provider) UploadFile(rw http.ResponseWriter, r *http.Request) {
	// check the user login
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		errorresp.Error(rw, apierrors.ErrUploadFile.NotLogin())
		return
	}

	// check the size
	if r.ContentLength > int64(p.Cfg.Limit.FileMaxUploadSize) {
		errorresp.Error(rw, apierrors.ErrUploadTooLargeFile.InvalidParameter(errors.Errorf("max file size: %s", p.Cfg.Limit.FileMaxUploadSize.String())))
		return
	}

	// get the file
	if err := r.ParseMultipartForm(int64(p.Cfg.Limit.FileMaxMemorySize)); err != nil {
		errorresp.Error(rw, apierrors.ErrUploadFile.InvalidParameter(errors.Errorf("err: %s", err)))
		return
	}
	formFile, fileHeader, err := r.FormFile("file")
	if err != nil {
		errorresp.Error(rw, apierrors.ErrUploadFile.InternalError(err))
		return
	}
	defer formFile.Close()

	fileExtension := filepath.Ext(fileHeader.Filename)
	if !p.Cfg.Security.FileTypeCarryActiveContentAllowed && strutil.Exist(p.Cfg.Security.FileTypesCanCarryActiveContent, strutil.TrimPrefixes(fileExtension, ".")) {
		errorresp.Error(rw, apierrors.ErrUploadFile.InvalidParameter(errors.Errorf("cannot upload file with type: %s", fileExtension)))
		return
	}
	// get params
	const (
		paramFileFrom  = "fileFrom"
		paramFileFrom2 = "from"
		paramPublic    = "public"
		paramEncrypt   = "encrypt"
		paramExpiredIn = "expiredIn" // such as "300ms", "-1.5h" or "2h45m". "0" means it doesn't expire.
	)
	fileFrom := r.URL.Query().Get(paramFileFrom)
	if fileFrom == "" {
		fileFrom = r.URL.Query().Get(paramFileFrom2)
	}
	public, _ := strconv.ParseBool(r.URL.Query().Get(paramPublic))
	encrypt, _ := strconv.ParseBool(r.URL.Query().Get(paramEncrypt))
	expiredInStr := r.URL.Query().Get(paramExpiredIn)
	var expiredAt *time.Time
	if expiredInStr != "" {
		expiredIn, err := time.ParseDuration(expiredInStr)
		if err != nil {
			errorresp.Error(rw, apierrors.ErrUploadFile.InvalidParameter(fmt.Sprintf("invalid expiredIn: %s", expiredIn)))
			return
		}
		if expiredIn != 0 {
			t := time.Now().Add(expiredIn)
			expiredAt = &t
		}
	}

	// 处理文件元信息
	if unescapeFilename, err := url.QueryUnescape(fileHeader.Filename); err == nil {
		fileHeader.Filename = unescapeFilename
	}

	// 上传文件
	file, err := p.fileService.UploadFile(filetypes.FileUploadRequest{
		FileNameWithExt: fileHeader.Filename,
		ByteSize:        fileHeader.Size,
		FileReader:      formFile,
		From:            fileFrom,
		IsPublic:        public,
		Creator:         identityInfo.UserID,
		Encrypt:         encrypt,
		ExpiredAt:       expiredAt,
	})
	if err != nil {
		errorresp.Error(rw, err)
		return
	}
	httpserver.WriteData(rw, file, file.Creator)
	return
}

const (
	headerValueApplicationJSON = "application/json"
)

func (p *provider) DownloadFile(rw http.ResponseWriter, r *http.Request) {
	var err error
	statusCode := http.StatusInternalServerError
	defer func() {
		if err != nil {
			// set Content-Type before statusCode, or we will get text/plain
			rw.Header().Set(headerContentType, headerValueApplicationJSON)
			rw.WriteHeader(statusCode)
			var jsonResp pb.FileDownloadFailResponse
			jsonResp.Success = false
			jsonResp.Error = &commonpb.ResponseError{}
			jsonResp.Error.Code = "ErrDownloadFile"
			if apiErr, ok := err.(*errorresp.APIError); ok {
				jsonResp.Error.Msg = apiErr.Error()
				if jsonResp.Error.Msg == "" {
					jsonResp.Error.Msg = apiErr.Code()
				}
			} else {
				jsonResp.Error.Msg = err.Error()
			}
			if wErr := json.NewEncoder(rw).Encode(&jsonResp); wErr != nil {
				logrus.Errorf("failed to write http response while download file failed, downloadErr: %v, writeRespErr: %v", err, wErr)
			}
			// err 已经在 response body 中返回，不需要 writerHandler 再次处理
			err = nil
		}
	}()

	// get from path variable
	uuid, _ := infrahttpserver.Var(r, "uuid")
	if uuid == "" {
		// get from query param
		uuid = r.URL.Query().Get("file")
		if uuid == "" {
			statusCode = http.StatusBadRequest
			err = errors.Errorf("no file specified")
			return
		}
	}

	file, err := p.db.GetFileByUUID(uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = apierrors.ErrDownloadFile.NotFound()
			return
		}
		err = apierrors.ErrDownloadFile.InvalidParameter(uuid)
		return
	}

	// 校验用户登录
	if !file.Extra.IsPublic && !p.Cfg.Security.DisableFileDownloadPermissionValidate {
		_, err = user.GetIdentityInfo(r)
		if err != nil {
			err = apierrors.ErrDownloadFile.NotLogin()
			return
		}
	}

	if _, err = p.fileService.DownloadFile(rw, file); err != nil {
		return
	}

	return
}

func (p *provider) HeadFile(rw http.ResponseWriter, r *http.Request) {
	// get from path variable
	uuid, _ := infrahttpserver.Var(r, "uuid")
	if uuid == "" {
		// get from query param
		uuid = r.URL.Query().Get("file")
		if uuid == "" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	file, err := p.db.GetFileByUUID(uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// 校验用户登录
	if file.Extra.IsPublic == false {
		_, err = user.GetIdentityInfo(r)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	rw.Header().Set(HeaderContentLength, strconv.FormatInt(file.ByteSize, 10))

	return
}

// DeleteFile 删除文件
func (p *provider) DeleteFile(rw http.ResponseWriter, r *http.Request) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		errorresp.Error(rw, apierrors.ErrDeleteFile.NotLogin())
		return
	}
	// 只有内部调用允许删除文件
	if !CheckInternalPermission(identityInfo) {
		errorresp.Error(rw, apierrors.ErrDeleteFile.AccessDenied())
		return
	}

	// 获取文件
	uuid, _ := infrahttpserver.Var(r, "uuid")
	file, err := p.db.GetFileByUUID(uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errorresp.Error(rw, apierrors.ErrDeleteFile.NotFound())
			return
		}
		errorresp.Error(rw, apierrors.ErrDeleteFile.InvalidParameter(err))
		return
	}

	// delete
	if err := p.fileService.DeleteFile(file); err != nil {
		errorresp.Error(rw, apierrors.ErrDeleteFile.InternalError(err))
		return
	}

	httpserver.WriteData(rw, nil)
	return
}

func CheckInternalPermission(identityInfo apistructs.IdentityInfo) bool {
	if identityInfo.IsInternalClient() {
		return true
	}
	return isReservedInternalServiceAccount(identityInfo.UserID)
}

// isReservedInternalServiceAccount 是否为内部服务账号
func isReservedInternalServiceAccount(userID string) bool {
	// TODO: ugly code
	// all (1000,5000) users is reserved as internal service account
	if v, err := strutil.Atoi64(userID); err == nil {
		if v > 1000 && v < 5000 && userID != apistructs.SupportID {
			return true
		}
	}
	return false
}
