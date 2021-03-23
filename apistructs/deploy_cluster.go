package apistructs

type DeployClusterJump struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
}

type DeployClusterRequest struct {
	OrgID    int               `json:"orgID"`
	Jump     DeployClusterJump `json:"jump"`
	Config   Sysconf           `json:"config"`
	DeployID string            `json:"deployID"`
}
