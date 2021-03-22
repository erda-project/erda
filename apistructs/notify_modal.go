package apistructs

type EditOrCreateModalData struct {
	Name     string   `json:"name"`   //通知名称
	Target   int      `json:"target"` //选中的通知群组id
	Items    []string `json:"items"`  //选中的通知模版id
	Id       int      `json:"id"`
	Channels []string `json:"channels"` //通知方式
}
type InParams struct {
	ScopeType string `json:"scopeType"`
	ScopeId   string `json:"scopeId"`
}
type NotifyDetailResponse struct {
	Header
	Data DetailResponse `json:"data"`
}

type DetailResponse struct {
	Id         int64  `json:"id"`
	NotifyID   string `json:"notifyId"`
	NotifyName string `json:"notifyName"`
	Target     string `json:"target"`
	GroupType  string `json:"groupType"`
}

type AllTemplatesResponse struct {
	Header
	Data []*TemplateRes `json:"data"`
}

type TemplateRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
