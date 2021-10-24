package exception

type Erda_event struct {
	EventId        string            `json:"event_id"`
	Timestamp      int64             `json:"timestamp"`
	RequestId      string            `json:"request_id"`
	ErrorId        string            `json:"error_id"`
	Stacks         []string          `json:"stacks"`
	Tags           map[string]string `json:"tags"`
	MetaData       map[string]string `json:"meta_data"`
	RequestContext map[string]string `json:"request_context"`
	RequestHeaders map[string]string `json:"request_headers"`
}


