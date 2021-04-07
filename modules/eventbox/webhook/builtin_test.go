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

package webhook

// func TestCreateBuiltinHooks(t *testing.T) {
// 	impl, err := NewWebHookImpl()
// 	assert.Nil(t, err)
// 	assert.Nil(t, createIfNotExist(impl, &CreateHookRequest{
// 		Name:   "test-builtin-hook",
// 		Events: []string{"test-event"},
// 		URL:    "http://xxx/a",
// 		Active: true,
// 		HookLocation: HookLocation{
// 			Org:         "-1",
// 			Project:     "-1",
// 			Application: "-1",
// 		},
// 	}))
// }

// func TestMakeSureBuiltinHooks(t *testing.T) {
// 	impl, err := NewWebHookImpl()
// 	assert.Nil(t, err)
// 	assert.Nil(t, MakeSureBuiltinHooks(impl))
// }
