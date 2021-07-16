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

package event

// var eventboxAddr = "eventbox.marathon.l4lb.thisdcos.directory:9528"
//
// func TestWebhookCreate(t *testing.T) {
// 	server, err := NewWebhook(eventboxAddr)
// 	assert.Nil(t, err)
//
// 	spec := apistructs.WebhookCreateRequest{
// 		Name:   "test-hook",
// 		Active: true,
// 		Events: []string{"test-event"},
// 		URL:    "https://eventbox.test.terminus.io/aaa",
// 		HookLocation: apistructs.HookLocation{
// 			Org:         "-1",
// 			Project:     "-1",
// 			Application: "-1",
// 		},
// 	}
//
// 	err = server.Create(spec)
// 	assert.Nil(t, err)
//
// 	hooks, err := server.List()
// 	assert.Nil(t, err)
//
// 	for _, hook := range hooks {
// 		if hook.Name == spec.Name {
// 			t.Logf("successfully found the hook(%s) created before", spec.Name)
// 			return
// 		}
// 	}
//
// 	t.Errorf("failed to create webhooks")
// }
//
// func TestWebhookList(t *testing.T) {
// 	server, err := NewWebhook(eventboxAddr)
// 	assert.Nil(t, err)
//
// 	hooks, err := server.List()
// 	assert.Nil(t, err)
//
// 	t.Logf("TestWebhookList result: %+v", hooks)
// }
