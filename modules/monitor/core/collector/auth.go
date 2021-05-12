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
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/secret"
	"github.com/erda-project/erda/pkg/secret/validator"
)

var bdl = bundle.New(bundle.WithCMDB())

type staticSKProvider struct {
	SecretKey string `file:"secretKey"`
}

func (c *collector) authSignedRequest() httpserver.Interceptor {
	return func(handler func(ctx httpserver.Context) error) func(ctx httpserver.Context) error {
		return func(ctx httpserver.Context) error {
			ak, ok := validator.GetAccessKeyID(ctx.Request())
			if !ok {
				return handler(ctx)
			}

			var sk string
			switch c.Cfg.SignAuth.SKProvider {
			case "static":
				sk, ok = c.Cfg.SignAuth.Config["secret_key"]
				if !ok {
					return echo.NewHTTPError(http.StatusUnauthorized, "no secret_key in config with static sk_provider")
				}
			default:
				aksk, err := bdl.GetAkSkByAk(ak)
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "unable to get aksk")
				}
				sk = aksk.Sk
			}
			vd := validator.NewHMACValidator(secret.AkSkPair{AccessKeyID: ak, SecretKey: sk})
			if res := vd.Verify(ctx.Request()); !res.Ok {
				return echo.NewHTTPError(http.StatusUnauthorized, res.Message)
			}
			return handler(ctx)
		}
	}
}

func (c *collector) basicAuth() interface{} {
	return middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Validator: func(username string, password string, context echo.Context) (bool, error) {
			if username == c.Cfg.Auth.Username && password == c.Cfg.Auth.Password {
				return true, nil
			}
			return false, nil
		},
		Skipper: func(context echo.Context) bool {
			if c.Cfg.Auth.Force {
				return false
			}
			// 兼容旧版本，没有添加认证的客户端，这个版本先跳过
			authorizationHeader := context.Request().Header.Get("Authorization")
			return authorizationHeader == ""
		},
	})
}
