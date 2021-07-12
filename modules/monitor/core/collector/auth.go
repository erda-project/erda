// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package collector

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func (p *provider) basicAuth() interface{} {
	return middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Validator: func(username string, password string, context echo.Context) (bool, error) {
			if username == p.Cfg.Auth.Username && password == p.Cfg.Auth.Password {
				return true, nil
			}
			return false, nil
		},
		Skipper: func(context echo.Context) bool {
			if p.Cfg.Auth.Force {
				return false
			}
			// 兼容旧版本，没有添加认证的客户端，这个版本先跳过
			authorizationHeader := context.Request().Header.Get("Authorization")
			return authorizationHeader == ""
		},
	})
}
