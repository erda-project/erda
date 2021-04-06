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

package filters

// func TestRegisterFilter(t *testing.T) {
// 	r, err := register.New()
// 	assert.Nil(t, err)
// 	filter := NewRegisterFilter(r)
// 	m := types.Message{
// 		Sender:  "self",
// 		Content: "2333",
// 		Labels: map[types.LabelKey]interface{}{
// 			types.LabelKey(constant.RegisterLabelKey): []string{"aaa"},
// 			"other": "value",
// 		},
// 		Time: 0,
// 	}
// 	assert.Nil(t, r.Put("aaa", map[types.LabelKey]interface{}{
// 		"bbb": "1",
// 		"ccc": "2",
// 	}))
// 	derr := filter.Filter(&m)

// 	assert.True(t, derr.IsOK())
// 	if !derr.IsOK() {
// 		fmt.Printf("%+v\n", derr) // debug print

// 	}

// 	assert.Equal(t, "1", m.Labels["/bbb"])
// 	assert.Equal(t, "value", m.Labels["other"])
// 	assert.Equal(t, []string{"aaa"}, m.Labels[types.LabelKey(constant.RegisterLabelKey)])

// }
