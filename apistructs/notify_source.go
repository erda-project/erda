package apistructs

type NotifySource struct {
	ID         int64       `json:"-"`
	Name       string      `json:"name"`
	SourceType string      `json:"sourceType"`
	SourceID   string      `json:"sourceId"`
	Params     interface{} `json:"params"`
}

type DeleteNotifySourceRequest struct {
	SourceType string `json:"sourceType"`
	SourceID   string `json:"sourceId"`
	OrgID      int64  `json:"-"`
}
