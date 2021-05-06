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

package register

// func TestRegister(t *testing.T) {
// 	r, err := New()
// 	assert.Nil(t, err)
// 	labels := map[types.LabelKey]interface{}{
// 		"/777": "888",
// 		"999":  "101010",
// 	}
// 	assert.Nil(t, r.Put("666", labels))
// 	labels1 := r.PrefixGet("666")

// 	assert.Equal(t, "888", labels1["/666"]["/777"])
// 	assert.Nil(t, r.Del("666"))
// 	assert.Len(t, r.PrefixGet("000"), 0)
// 	assert.Nil(t, r.Del("000"))
// }

// func TestRegisterHTTP(t *testing.T) {
// 	s, err := server.New()
// 	assert.Nil(t, err)
// 	reg, err := New()
// 	assert.Nil(t, err)
// 	reghttp := NewHTTP(reg)
// 	s.AddEndPoints(reghttp.GetHTTPEndPoints())
// 	go func() {
// 		assert.Nil(t, s.Start())
// 	}()
// 	defer s.Stop()

// 	time.Sleep(1 * time.Second)
// 	resp, err := httpclient.New().Put("127.0.0.1:9528").Path("/api/dice/eventbox/register").Header("Accept", "application/json").JSONBody(PutRequest{
// 		Key: "testregisterhttp_1",
// 		Labels: map[types.LabelKey]interface{}{
// 			"testregisterhttp1": "testregisterhttp2",
// 			"testregisterhttp3": "testregisterhttp4",
// 		},
// 	}).Do().DiscardBody()
// 	assert.Nil(t, err)
// 	if resp != nil {
// 		assert.True(t, resp.IsOK())
// 	}
// 	m := reg.PrefixGet("testregisterhttp_1")
// 	assert.Len(t, m, 1)
// 	_, ok := m["/testregisterhttp_1"]["/testregisterhttp1"]
// 	assert.True(t, ok)
// 	resp, err = httpclient.New().Delete("127.0.0.1:9528").Header("Accept", "application/json").Path("/api/dice/eventbox/register").JSONBody(DelRequest{"testregisterhttp_1"}).Do().DiscardBody()

// 	assert.Nil(t, err)
// 	if resp != nil {
// 		assert.True(t, resp.IsOK())
// 	}
// 	m = reg.PrefixGet("testregisterhttp_1")
// 	assert.Len(t, m, 0)
// }

// func TestRegisterDelete(t *testing.T) {
// 	s, _ := server.New()
// 	reg, _ := New()
// 	regHTTP := NewHTTP(reg)
// 	s.AddEndPoints(regHTTP.GetHTTPEndPoints())
// 	go func() {
// 		assert.Nil(t, s.Start())
// 	}()
// 	defer s.Stop()
// 	time.Sleep(1 * time.Second)
// 	resp, err := httpclient.New().Put("127.0.0.1:9528").Path("/api/dice/eventbox/register").Header("Accept", "application/json").JSONBody(PutRequest{
// 		Key: "/testregisterhttp/lvl1/lvl2",
// 		Labels: map[types.LabelKey]interface{}{
// 			"testregisterhttp1": "testregisterhttp2",
// 			"testregisterhttp3": "testregisterhttp4",
// 		},
// 	}).Do().DiscardBody()
// 	assert.Nil(t, err)
// 	if resp != nil {
// 		assert.True(t, resp.IsOK())
// 	}
// 	assert.True(t, len(reg.PrefixGet("/testregisterhttp/lvl1")) > 0)
// 	var buf bytes.Buffer
// 	resp, err = httpclient.New().Delete("127.0.0.1:9528").Path("/api/dice/eventbox/register").Header("Accept", "application/json").JSONBody(DelRequest{"/testregisterhttp/lvl1/lvl2"}).Do().Body(&buf)
// 	assert.Nil(t, err)
// 	if resp != nil {
// 		assert.True(t, resp.IsOK())
// 	}
// }
