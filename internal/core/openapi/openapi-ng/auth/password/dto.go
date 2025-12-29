package password

import (
	"github.com/erda-project/erda/apistructs"
	identity "github.com/erda-project/erda/internal/core/user/common"
)

type LoginParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  *identity.UserInfo      `json:"user,omitempty"`
	Token *apistructs.OAuth2Token `json:"token,omitempty"`
}
