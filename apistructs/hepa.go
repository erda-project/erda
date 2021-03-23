package apistructs

// AuthType
const (
	AT_KEY_AUTH   = "key-auth"
	AT_OAUTH2     = "oauth2"
	AT_SIGN_AUTH  = "sign-auth"
	AT_HMAC_AUTH  = "hmac-auth"
	AT_ALIYUN_APP = "aliyun-app"
)

// AclType
const (
	ACL      = "acl"
	ACL_NONE = ""
	ACL_ON   = "on"
	ACL_OFF  = "off"
)

// Scene
const (
	OPENAPI_SCENE = "openapi"
	WEBAPI_SCENE  = "webapi"
	UNITY_SCENE   = "unity"
)

type EndpointInfoResponse struct {
	Header
	Data PackageInfoDto `json:"data"`
}

type ClientInfoResponse struct {
	Header
	Data ClientInfoDto `json:"data"`
}

type TenantGroupResponse struct {
	Header
	Data string `json:"data"`
}

type ClientInfoDto struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type LimitType struct {
	Day    *int `json:"qpd,omitempty"`
	Hour   *int `json:"qph,omitempty"`
	Minute *int `json:"qpm,omitempty"`
	Second *int `json:"qps,omitempty"`
}

type ChangeLimitsReq struct {
	Limits []LimitType `json:"limits"`
}

type OpenapiInfoDto struct {
	ApiId       string `json:"apiId"`
	CreateAt    string `json:"createAt"`
	DiceApp     string `json:"diceApp"`
	DiceService string `json:"diceService"`
	Origin      string `json:"origin"`
	Mutable     bool   `json:"mutable"`
	OpenapiDto
}

type OpenapiDto struct {
	ApiPath             string   `json:"apiPath"`
	RedirectType        string   `json:"redirectType"`
	RedirectAddr        string   `json:"redirectAddr"`
	RedirectPath        string   `json:"redirectPath"`
	RedirectApp         string   `json:"redirectApp"`
	RedirectService     string   `json:"redirectService"`
	RedirectRuntimeId   string   `json:"redirectRuntimeId"`
	RedirectRuntimeName string   `json:"redirectRuntimeName"`
	Method              string   `json:"method,omitempty"`
	AllowPassAuth       bool     `json:"allowPassAuth"`
	Description         string   `json:"description"`
	Hosts               []string `json:"hosts"`
}

type PackageDto struct {
	Name             string   `json:"name"`
	BindDomain       []string `json:"bindDomain"`
	AuthType         string   `json:"authType"`
	AclType          string   `json:"aclType"`
	Scene            string   `json:"scene"`
	Description      string   `json:"description"`
	NeedBindCloudapi bool     `json:"needBindCloudapi"`
}

type PackageInfoDto struct {
	Id       string `json:"id"`
	CreateAt string `json:"createAt"`
	PackageDto
}
