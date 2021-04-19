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

package apitestsv2

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/jsonpath"
)

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

func TestJsonPath(t *testing.T) {
	var a map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(`{"success":true}`), &a))
	data, err := jsonpath.Get(a, "success")
	assert.NoError(t, err)
	spew.Dump(data)
}
