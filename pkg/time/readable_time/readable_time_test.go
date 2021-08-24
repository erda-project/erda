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

package readable_time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadableTime(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2018-12-05T16:54:57+08:00")
	assert.Nil(t, err)
	t2, err := time.Parse(time.RFC3339, "2018-12-05T16:54:59+08:00")
	assert.Nil(t, err)

	a := readableTime(t1, t2)
	assert.Equal(t, int64(2), a.Second)
}
