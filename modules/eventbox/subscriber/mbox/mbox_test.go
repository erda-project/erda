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

package sms

// func TestMBoxSend(t *testing.T) {

// 	subscriber := New(bundle.New(bundle.WithCMDB()))
// 	params := map[string]string{
// 		"notice": "notice",
// 		"link":   "link",
// 		"title":  "机器告警通知",
// 	}
// 	request := map[string]interface{}{
// 		"template": "测试测试",
// 		"params":   params,
// 		"orgID":    1,
// 		"module":   "monitor",
// 	}

// 	ids := []string{"1", "2"}
// 	dest, _ := json.Marshal(ids)
// 	msg := &types.Message{
// 		Content: request,
// 		Time:    time.Now().Unix(),
// 	}
// 	content, _ := json.Marshal(request)
// 	errs := subscriber.Publish(string(dest), string(content), msg.Time, msg)
// 	if len(errs) > 0 {
// 		panic(errs[0])
// 	}
// }
