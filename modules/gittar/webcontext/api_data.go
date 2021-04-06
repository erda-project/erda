package webcontext

type ApiData struct {
	Success bool        `json:"success"`
	Err     interface{} `json:"err"`
	Data    interface{} `json:"data"`
	UserIDs []string    `json:"userIDs,omitempty"`
}
