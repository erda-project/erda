package password

import (
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/apistructs"
)

type LoginParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  *commonpb.UserInfo      `json:"user,omitempty"`
	Token *apistructs.OAuth2Token `json:"token,omitempty"`
}
