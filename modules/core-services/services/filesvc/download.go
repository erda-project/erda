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

package filesvc

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/kms/kmscrypto"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
	_ "github.com/erda-project/erda/pkg/mimetype"
)

const (
	headerContentType        = "Content-Type"
	headerContentDisposition = "Content-Disposition"
	HeaderContentLength      = "Content-Length" // The Content-Length entity header indicates the size of the entity-body, in bytes, sent to the recipient.
)

var FileTypesCanCarryActiveContent = []string{".html", ".js", ".xml"}

// DownloadFile write file to writer `w`,  return corresponding file http response headers.
func (svc *FileService) DownloadFile(w io.Writer, file dao.File) (headers map[string]string, err error) {
	// check path
	if err := checkPath(file.FullRelativePath); err != nil {
		return nil, apierrors.ErrDownloadFile.InvalidParameter(err)
	}

	// check expired
	if file.ExpiredAt != nil && time.Now().After(*file.ExpiredAt) {
		return nil, apierrors.ErrDownloadFile.InvalidParameter("file already expired")
	}

	// storager
	storager := svc.GetStorage(file.StorageType)
	reader, err := storager.Read(file.FullRelativePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, apierrors.ErrDownloadFile.NotFound()
		}
		return nil, apierrors.ErrDownloadFile.InternalError(err)
	}
	// 解密 信封加密 文件数据
	if file.Extra.Encrypt {
		// 调用 KMS 解密 DEK
		dekDecryptResp, err := svc.bdl.KMSDecrypt(apistructs.KMSDecryptRequest{
			DecryptRequest: kmstypes.DecryptRequest{
				KeyID:            file.Extra.KMSKeyID,
				CiphertextBase64: file.Extra.DEKCiphertextBase64,
			},
		})
		if err != nil {
			return nil, apierrors.ErrDownloadFileDecrypt.InternalError(err)
		}
		DEK, err := base64.StdEncoding.DecodeString(dekDecryptResp.PlaintextBase64)
		if err != nil {
			return nil, apierrors.ErrDownloadFileDecrypt.InternalError(err)
		}
		// 获取文件内容
		fileBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, apierrors.ErrDownloadFileDecrypt.InternalError(err)
		}
		filePlaintext, err := kmscrypto.AesGcmDecrypt(DEK, fileBytes, generateAesGemAdditionalData(file.From))
		if err != nil {
			return nil, apierrors.ErrDownloadFileDecrypt.InternalError(err)
		}
		reader = bytes.NewBuffer(filePlaintext)
	}

	headers = map[string]string{
		headerContentDisposition: headerValueDispositionInline(file.Ext, file.DisplayName),
		HeaderContentLength:      strconv.FormatInt(file.ByteSize, 10),
	}

	contentType := mime.TypeByExtension(file.Ext)
	if contentType != "" {
		headers[headerContentType] = contentType
	}

	// set headers to http ResponseWriter `w` before write into `w`.
	if rw, ok := w.(http.ResponseWriter); ok {
		for k, v := range headers {
			rw.Header().Set(k, v)
		}
	}

	if _, err := io.Copy(w, reader); err != nil {
		return nil, apierrors.ErrDownloadFile.InternalError(err)
	}

	return
}
