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
