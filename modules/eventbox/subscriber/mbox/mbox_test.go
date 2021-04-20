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
