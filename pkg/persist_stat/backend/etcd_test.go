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

package backend

import (
	"testing"
	"time"

	// "github.com/erda-project/erda/pkg/jsonstore"

	"path/filepath"
	"strconv"

	"github.com/stretchr/testify/assert"
)

// func TestEtcdBackend(t *testing.T) {
// 	js, err := jsonstore.New()
// 	assert.Nil(t, err)
// 	stat := NewEtcd(js)
// 	assert.Nil(t, stat.Clear(time.Now()))
// 	stat.Emit("tag1", 1)
// 	stat.Emit("tag2", 2)
// 	stat.Emit("tag2", 3)
// 	stat.Emit("tag1", 4)

// 	time.Sleep(61 * time.Second)

// 	r, err := stat.Last1Day()
// 	assert.Equal(t, int64(5), r["tag1"])
// 	assert.Equal(t, int64(5), r["tag2"])
// }

func TestClearDir(t *testing.T) {
	now := time.Now()
	dir := clearDir("aa", 2, now)
	before := now.Add(-3 * 24 * time.Hour)
	d := filepath.Join("aa", strconv.Itoa(before.Year()), strconv.Itoa(int(before.Month())), strconv.Itoa(before.Day()))
	assert.Equal(t, d, dir)

}
