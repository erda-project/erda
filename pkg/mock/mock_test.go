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

package mock

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	s := randString(Integer)
	i, err := strconv.Atoi(s)
	assert.NoError(t, err)
	fmt.Println(s, i)

	s = randString(String)
	fmt.Println(s)
}

func TestGetTime(t *testing.T) {
	t.Log("s:", getTime(TimeStamp))
	t.Log("s-hour:", getTime(TimeStampHour))
	t.Log("s-after-hour:", getTime(TimeStampAfterHour))
	t.Log("s-day:", getTime(TimeStampDay))
	t.Log("s-after-day:", getTime(TimeStampAfterDay))
	t.Log("ms:", getTime(TimeStampMs))
	t.Log("ms-hour:", getTime(TimeStampMsHour))
	t.Log("ms-after-hour:", getTime(TimeStampMsAfterHour))
	t.Log("ms-day:", getTime(TimeStampMsDay))
	t.Log("ms-after-day:", getTime(TimeStampMsAfterDay))
	t.Log("ns:", getTime(TimeStampNs))
	t.Log("ns-hour:", getTime(TimeStampNsHour))
	t.Log("ns-after-hour:", getTime(TimeStampNsAfterHour))
	t.Log("ns-day:", getTime(TimeStampNsDay))
	t.Log("ns-after-day:", getTime(TimeStampNsAfterDay))
}
