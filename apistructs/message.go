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
