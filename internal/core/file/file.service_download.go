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
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/file/db"
	"github.com/erda-project/erda/internal/core/legacy/services/apierrors"
	"github.com/erda-project/erda/pkg/kms/kmscrypto"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
	_ "github.com/erda-project/erda/pkg/mimetype"
)

const (
	headerContentType        = "Content-Type"
	headerContentDisposition = "Content-Disposition"
	HeaderContentLength      = "Content-Length" // The Content-Length entity header indicates the size of the entity-body, in bytes, sent to the recipient.

	headerContentTypePng       = "image/png"
	headerContentDefaultExtKey = "default"
)

var (
	// allowedContentTypes is a map of allowed file code(key), ext(key) and content type(value).
	// reference link: https://www.garykessler.net/library/file_sigs.html
	allowedContentTypes = map[string]map[string]string{
		"FFD8FFE0": {
			".jpg":                     headerContentTypePng,
			".jfif":                    headerContentTypePng,
			".jpeg":                    headerContentTypePng,
			".jpe":                     headerContentTypePng,
			headerContentDefaultExtKey: headerContentTypePng,
		}, // jpg
		"FFD8FFE1": {
			".jpg":                     headerContentTypePng,
			headerContentDefaultExtKey: headerContentTypePng,
		}, // jpg
		"FFD8FFE8": {
			".jpg":                     headerContentTypePng,
			headerContentDefaultExtKey: headerContentTypePng,
		}, // jpg
		"47494638": {
			".gif":                     headerContentTypePng,
			headerContentDefaultExtKey: headerContentTypePng,
		}, // gif
		"89504E47": {
			".png":                     headerContentTypePng,
			headerContentDefaultExtKey: headerContentTypePng,
		}, // png
		"504B0304": {
			".apk":                     "application/vnd.android.package-archive",
			".zip":                     "application/zip",
			".jar":                     "application/java-archive",
			".kmz":                     "application/vnd.google-earth.kmz",
			"kwd":                      "application/vnd.kde.kword",
			"epub":                     "application/epub+zip",
			headerContentDefaultExtKey: "application/octet-stream",
		}, // apk, zip, jar, kmz, kwd, epub
	}
)

// DownloadFile write file to writer `w`,  return corresponding file http response headers.
func (s *fileService) DownloadFile(w io.Writer, file db.File) (headers map[string]string, err error) {
	// check path
	if err := checkPath(file.FullRelativePath); err != nil {
		return nil, apierrors.ErrDownloadFile.InvalidParameter(err)
	}

	// check expired
	if file.ExpiredAt != nil && time.Now().After(*file.ExpiredAt) {
		return nil, apierrors.ErrDownloadFile.InvalidParameter("file already expired")
	}

	// storager
	storager := s.GetStorage(file.StorageType)
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
		dekDecryptResp, err := s.bdl.KMSDecrypt(apistructs.KMSDecryptRequest{
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
		fileBytes, err := io.ReadAll(reader)
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
		headerContentDisposition: s.headerValueDispositionInline(file.Ext, file.DisplayName),
		HeaderContentLength:      strconv.FormatInt(file.ByteSize, 10),
	}
	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)
	contentType := GetFileContentType(tee, file.Ext)
	if contentType != "" {
		headers[headerContentType] = contentType
	}

	// set headers to http ResponseWriter `w` before write into `w`.
	if rw, ok := w.(http.ResponseWriter); ok {
		for k, v := range headers {
			rw.Header().Set(k, v)
		}
	}

	if _, err := io.Copy(w, &buf); err != nil {
		return nil, apierrors.ErrDownloadFile.InternalError(err)
	}
	if _, err := io.Copy(w, reader); err != nil {
		return nil, apierrors.ErrDownloadFile.InternalError(err)
	}

	return
}

// GetFileContentType judge file content type by file header.
// If file header is found in allowedContentTypes, return content type, otherwise return application/octet-stream.
func GetFileContentType(r io.Reader, ext string) string {
	contentType := "application/octet-stream"
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)
	headerBuf := make([]byte, 20)
	n, err := tee.Read(headerBuf)
	if err != nil {
		return contentType
	}
	fileCode := bytesToHexString(headerBuf[:n])
	for k, extData := range allowedContentTypes {
		if strings.HasPrefix(strings.ToLower(fileCode), strings.ToLower(k)) {
			if contentType, ok := extData[ext]; ok {
				return contentType
			}
			return extData[headerContentDefaultExtKey]
		}
	}

	return contentType
}

func bytesToHexString(src []byte) string {
	res := bytes.Buffer{}
	if src == nil || len(src) <= 0 {
		return ""
	}
	temp := make([]byte, 0)
	i, length := 100, len(src)
	if length < i {
		i = length
	}
	for j := 0; j < i; j++ {
		sub := src[j] & 0xFF
		hv := hex.EncodeToString(append(temp, sub))
		if len(hv) < 2 {
			res.WriteString(strconv.FormatInt(int64(0), 10))
		}
		res.WriteString(hv)
	}
	return res.String()
}
