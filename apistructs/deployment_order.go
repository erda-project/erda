package apistructs

const (
	TypePipeline           = "PIPELINE"
	TypeApplicationRelease = "APPLICATION_RELEASE"
	TypeProjectRelease     = "PROJECT_RELEASE"
)

type DeploymentOrderStatus string

type DeploymentOrderCreateRequest struct {
	OrgID     uint64 `json:"orgId,omitempty"`
	Type      string `json:"type,omitempty"`
	ReleaseId string `json:"releaseId"`
	Workspace string `json:"workspace,omitempty"`
	Operator  string `json:"operator"`
}

type DeploymentOrderListData struct {
	Total int                    `json:"total"`
	List  []*DeploymentOrderItem `json:"list"`
}

type DeploymentOrderItem struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	ReleaseID string                `json:"releaseId"`
	Params    string                `json:"params"`
	Type      string                `json:"type"`
	Status    DeploymentOrderStatus `json:"status"`
	Operator  string                `json:"operator"`
}

type DeploymentOrderParam struct {
	Env  []DeploymentOrderParamItem `json:"env"`
	File []DeploymentOrderParamItem `json:"file"`
}

type DeploymentOrderParamItem struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	IsEncrypt bool   `json:"isEncrypt"`
}

type DeploymentOrderStatusMap map[string]DeploymentOrderStatusItem

type DeploymentOrderStatusItem struct {
	AppID            uint64           `json:"appId"`
	DeploymentID     uint64           `json:"deploymentId"`
	DeploymentStatus DeploymentStatus `json:"deploymentStatus"`
	RuntimeID        uint64           `json:"runtimeId"`
}
