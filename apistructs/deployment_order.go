package apistructs

import "time"

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
	AutoRun   bool   `json:"autoRun"`
}

type DeploymentOrderCreateResponse struct {
	Id              string                `json:"id"`
	Name            string                `json:"name"`
	Type            string                `json:"type"`
	ReleaseId       string                `json:"releaseId"`
	ProjectId       uint64                `json:"projectId"`
	ProjectName     string                `json:"projectName"`
	ApplicationId   int64                 `json:"applicationId"`
	ApplicationName string                `json:"applicationName"`
	Status          DeploymentOrderStatus `json:"status"`
}

type DeploymentOrderDeployRequest struct {
	DeploymentOrderId string
	Workspace         string `json:"workspace,omitempty"`
	Operator          string `json:"operator"`
}

type DeploymentOrderDetail struct {
	DeploymentOrderItem
	ReleaseVersion   string              `json:"releaseVersion"`
	ApplicationsInfo []*ApplicationsInfo `json:"applicationsInfo"`
}

type ApplicationsInfo struct {
	Name           string           `json:"name"`
	Param          string           `json:"param"`
	ReleaseVersion string           `json:"releaseVersion"`
	Branch         string           `json:"branch"`
	CommitId       string           `json:"commitId"`
	DiceYaml       string           `json:"diceYaml"`
	Status         DeploymentStatus `json:"status"`
}

type DeploymentOrderListData struct {
	Total int                    `json:"total"`
	List  []*DeploymentOrderItem `json:"list"`
}

type DeploymentOrderItem struct {
	ID                string                `json:"id"`
	Name              string                `json:"name"`
	ReleaseID         string                `json:"releaseId"`
	ReleaseVersion    string                `json:"releaseVersion"`
	Params            string                `json:"params,omitempty"`
	Type              string                `json:"type"`
	ApplicationStatus string                `json:"applicationStatus"`
	Status            DeploymentOrderStatus `json:"status"`
	Operator          string                `json:"operator"`
	CreatedAt         time.Time             `json:"createdAt"`
	UpdatedAt         time.Time             `json:"updatedAt"`
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

type DeploymentOrderParamData struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	ConfigType string `json:"configType"`
}

type DeploymentOrderStatusMap map[string]DeploymentOrderStatusItem

type DeploymentOrderStatusItem struct {
	AppID            uint64           `json:"appId"`
	DeploymentID     uint64           `json:"deploymentId"`
	DeploymentStatus DeploymentStatus `json:"deploymentStatus"`
	RuntimeID        uint64           `json:"runtimeId"`
}

type DeploymentOrderCancelRequest struct {
	DeploymentOrderId string
	Operator          string `json:"operator"`
	Force             bool   `json:"force"`
}
