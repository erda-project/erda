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

//import (
//	"os"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/pkg/filehelper"
//)
//
//func TestDownloadDiceFile(t *testing.T) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	b := New(WithCMDB())
//	r, err := b.DownloadDiceFile("1e5c1abc142446699437fe3ba7e70245")
//	assert.NoError(t, err)
//	assert.NoError(t, filehelper.CreateFile2("/Users/sfwn/111.jpg", r, 0644))
//}
//
//func BenchmarkFileEOF(b *testing.B) {
//	os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
//	bdl := New(WithAllAvailableClients())
//	fi, err := os.Stat("/tmp/bug.log")
//	assert.NoError(b, err)
//	f, err := os.Open("/tmp/bug.log")
//	defer f.Close()
//
//	for i := 0; i < b.N; i++ {
//		assert.NoError(b, err)
//		r, err := bdl.UploadFile(apistructs.FileUploadRequest{
//			FileNameWithExt: "bug.log",
//			ByteSize:        fi.Size(),
//			FileReader:      f,
//			From:            "bundle-test",
//			IsPublic:        true,
//			Encrypt:         false,
//			Creator:         "my",
//			ExpiredAt:       &[]time.Time{time.Now().Add(time.Hour * 24)}[0],
//		})
//		assert.NoError(b, err)
//		if err != nil {
//			b.Logf("upload failed, err: %v", err)
//		}
//
//		_, err = bdl.DownloadDiceFile(r.UUID)
//		assert.NoError(b, err)
//		if err != nil {
//			b.Logf("upload failed, err: %v", err)
//		}
//	}
//}
