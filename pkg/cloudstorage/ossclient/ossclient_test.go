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

package ossclient

//import (
//	"testing"
//
//	"github.com/sirupsen/logrus"
//	"github.com/stretchr/testify/assert"
//)
//
//const (
//	BUCKET = "terminus-dice"
//)
//
//var c *OssClient
//
//func TestNewClient(t *testing.T) {
//	_, err := New("oss-cn-hangzhou.aliyuncs.com", "xxx", "xx")
//	assert.Nil(t, err)
//}
//
//func TestClient_GetFileUrl(t *testing.T) {
//	url, err := c.GetFileUrl(BUCKET, "test.yml")
//	assert.Nil(t, err)
//	logrus.Info(">>", url)
//}
//
//func TestClient_UploadFile(t *testing.T) {
//	url, err := c.UploadFile(BUCKET, "yhp-test-alert.sh", "/Users/yuhaiping/Desktop/alert.sh")
//	assert.Nil(t, err)
//	logrus.Info(">> url:", url)
//}
//
//func TestOssClient_DownloadFile(t *testing.T) {
//	bs, err := c.DownloadFile(BUCKET, "test.yml")
//	assert.Nil(t, err)
//	logrus.Info(string(bs))
//}
