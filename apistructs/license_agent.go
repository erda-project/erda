package apistructs

type RegisterLicenseRequest struct {
	Scope   int64  `json:"scope"`
	License string `json:"license"`
}

type RegisterLicenseResponse struct {
	Id int64 `json:"id"`
}
