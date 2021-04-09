package template

type templateDTO struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Scope       string         `json:"scope"`
	ScopeID     string         `json:"scopeId"`
	ViewConfig  *ViewConfigDTO `json:"viewConfig"`
	CreatedAt   int64          `json:"createdAt"`
	UpdatedAt   int64          `json:"updatedAt"`
	Version     string         `json:"version"`
	Type        int64          `json:"type"`
}

type templateUpdate struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	ViewConfig  *ViewConfigDTO `json:"viewConfig"`
}

type templateOverview struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
	ScopeID     string `json:"scopeId"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	Version     string `json:"version"`
	Type        int64  `json:"type"`
}

type templateResp struct {
	TemplateDTO []*templateOverview `json:"list"`
	Total       int                 `json:"total"`
}

type templateSearch struct {
	ID       string `query:"id"`
	Scope    string `query:"scope" validate:"required"`
	ScopeID  string `query:"scopeId" validate:"required"`
	PageNo   int64  `query:"pageNo" validate:"gte=1" default:"20"`
	PageSize int64  `query:"pageSize" validate:"gte=1" default:"20"`
	Type     int64  `query:"type"`
	Name     string `query:"name"`
}

type templateType struct {
	Type int64
}
