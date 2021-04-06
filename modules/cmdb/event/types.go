package event

import (
	"github.com/erda-project/erda/apistructs"
)

// StandardResponse 标准 api 返回结构
type StandardResponse struct {
	apistructs.Header
	Data interface{} `json:"data,omitempty"`
}
