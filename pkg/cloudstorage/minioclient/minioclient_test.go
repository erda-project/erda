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

package minioclient

//import (
//	"testing"
//
//	"github.com/sirupsen/logrus"
//	"github.com/stretchr/testify/assert"
//)
//
//func before() *MinioClient {
//	c, err := New("127.0.0.1:9009", "accesskey", "secretkey")
//	if err != nil {
//		panic(err)
//	}
//	return c
//}
//
//func TestNewClient(t *testing.T) {
//	c, err := New("127.0.0.1:9000", "accesskey", "secretkey")
//	assert.Nil(t, err)
//	assert.Nil(t, c.HealthCheck())
//}
//
//func TestClient_GetFileUrl(t *testing.T) {
//	c := before()
//	url, err := c.GetFileUrl("test1", "001.png")
//	assert.Nil(t, err)
//	assert.NotEmpty(t, url)
//}
//
//func TestClient_UploadFile(t *testing.T) {
//	c := before()
//	url, err := c.UploadFile("test1", "d1.xml", "../testdata/d1.xml")
//	assert.Nil(t, err)
//	assert.NotEmpty(t, url)
//}
//
//func TestMinioClient_DownloadFile(t *testing.T) {
//	c := before()
//	bs, err := c.DownloadFile("test1", "TEST-surefile.xml")
//	assert.Nil(t, err)
//	logrus.Info(string(bs))
//}
