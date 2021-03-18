package apistructs

type EventBoxRequest struct {
	Sender  string                 `json:"sender"`
	Content interface{}            `json:"content"`
	Labels  map[string]interface{} `json:"labels"`
}

type EventBoxResponse struct {
	Header
}
type EventBoxGroupNotifyRequest struct {
	Sender        string
	GroupID       int64
	NotifyItem    *NotifyItem
	Channels      string
	NotifyContent *GroupNotifyContent
	Params        map[string]string
}
