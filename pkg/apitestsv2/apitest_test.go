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

func TestAssert(t *testing.T) {
	var (
		v   interface{} // value
		op  string      // operator
		e   string      // expect
		ret bool
		err error
	)

	// 测试 =
	v = map[string]string{
		"aa": "bb",
		"cc": "dd",
	}
	op = "="
	e = `map[aa:bb cc:dd]`
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = []int{1, 2, 3, 4}
	op = "="
	e = `[1 2 3 4]`
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = []string{"q1", "q2", "test"}
	op = "="
	e = `[q1 q2 test]`
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = 200.00
	op = "="
	e = "200.0000"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 !=
	v = "200.00"
	op = "!="
	e = "200.01"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 >=
	v = "200.001"
	op = ">="
	e = "100"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = 200
	op = ">="
	e = "200"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 <=
	v = "200"
	op = "<="
	e = "300"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = "200"
	op = "<="
	e = "200"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 >
	v = "200"
	op = ">"
	e = "100"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 <
	v = "200"
	op = "<"
	e = "300"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 contains
	ss := `{"result":{"tenantId":null,"extra":null,"userId":11339,"token":"96c21d03d04928450d3861806ca763e409df9ce5eb4d9ced04504b1b8a35aa3b","expireTime":2592000},"success":true}`
	err = json.Unmarshal([]byte(ss), &v)
	op = "contains"
	e = "\"userId\":11339"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = 123456788888
	op = "contains"
	e = "123"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = "hello123 test dice!!"
	op = "contains"
	e = "[a-z]+[0-9]+"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 not_contains
	v = "1e2bc0278349bc5a4ca7ab2ac5daf6fc34d99563ec89edcc87f64a219622c364"
	op = "not_contains"
	e = "2bcdsf"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 belong
	v = 25
	op = "belong"
	e = "[-30,25]"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = "2592000"
	op = "belong"
	e = "{2592000, 23,34}"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = 25
	op = "belong"
	e = "[2,235]"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = 20
	op = "belong"
	e = "(10,30)"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = 10
	op = "belong"
	e = "[10,30)"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = 2592000
	op = "belong"
	e = "{[10,2600000),300,20}"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = 300
	op = "belong"
	e = "{300,20,10,30}"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 not_belong
	v = 259
	op = "not_belong"
	e = "[2,235]"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 empty
	v = make(map[string]string)
	op = "empty"
	e = ""
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	v = []int{}
	op = "empty"
	e = ""
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 not_empty
	v = "test"
	op = "not_empty"
	e = "test"
	ret, err = doAssert(v, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 exist
	var val interface{}
	err = json.Unmarshal([]byte(`{"data":{"key1":{"key2": "test"},"key":{"key2": "test"}}}`), &val)
	assert.Equal(t, err, nil)
	op = "exist"
	e = `data.key.key2`
	ret, err = doAssert(val, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)

	// 测试 not_exist
	var val1 interface{}
	err = json.Unmarshal([]byte(`{"data":{"key1":{"key2": "test"},"key":{"key2": "test"}}}`), &val1)
	assert.Equal(t, err, nil)
	op = "not_exist"
	e = `data.key11`
	ret, err = doAssert(val1, op, e)
	assert.Equal(t, err, nil)
	assert.Equal(t, ret, true)
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

func TestJsonPath(t *testing.T) {
	var a map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(`{"success":true}`), &a))
	data, err := jsonpath.Get(a, "success")
	assert.NoError(t, err)
	spew.Dump(data)
}
