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

package apistructs

// MessageLabel alias as string
type MessageLabel = string

const (
	// DingdingLabel "DINGDING": ["<url-1>", "<url-2>"]
	DingdingLabel MessageLabel = "DINGDING"

	// DingdingMarkdownLabel "MARKDOWN": {"title": "title"}
	DingdingMarkdownLabel MessageLabel = "MARKDOWN"

	// HTTPLabel "HTTP": ["<url-1>", "<url-2>"]
	HTTPLabel MessageLabel = "HTTP"

	// HTTPHeaderLabel "HTTP-HEADERS": {"k1": "v1", "k2": "v2"}
	HTTPHeaderLabel MessageLabel = "HTTP-HEADERS"

	// DingdingATLabel
	// "AT":
	// {
	//  "atMobiles": [
	//     "1825718XXXX"
	//   ],
	//   "isAtAll": false
	// }
	DingdingATLabel MessageLabel = "AT"

	// DingdingWorkNoticeLabel see also 'https://open-doc.dingtalk.com/microapp/serverapi2/pgoxpy'
	// "DINGDING-WORKNOTICE":
	// [{
	//   "url": "<worknotice-url>",
	//   "agent_id": "<agentid>",
	//   "userid_list": ["<id1>", "<id2>"]
	// }, ...]
	DingdingWorkNoticeLabel MessageLabel = "DINGDING-WORKNOTICE"

	// MySQLLabel "MYSQL": "<table-name>"
	MySQLLabel MessageLabel = "MYSQL"
)

// MessageCreateRequest see also `bundle/messages.go'
type MessageCreateRequest struct {
	Sender  string                       `json:"sender"`
	Content interface{}                  `json:"content"`
	Labels  map[MessageLabel]interface{} `json:"labels"`
}
