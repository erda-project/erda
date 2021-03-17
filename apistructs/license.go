package apistructs

import "github.com/erda-project/erda/pkg/license"

// LicenseResponse 查询license响应数据
type LicenseResponse struct {
	Valid            bool             `json:"valid"`
	Message          string           `json:"message"`
	CurrentHostCount uint64           `json:"currentHostCount"`
	License          *license.License `json:"license"`
}
