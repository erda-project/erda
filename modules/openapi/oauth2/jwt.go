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

package oauth2

import (
	"encoding/base64"
	"encoding/json"
	errs "errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/utils/uuid"

	"github.com/erda-project/erda/apistructs"
)

const (
	JWTKey = "openapi_oauth2_token_secret"
)

// JWTAccessClaims jwt claims
type JWTAccessClaims struct {
	jwt.StandardClaims
	Payload apistructs.OpenapiOAuth2TokenPayload `json:"payload"`
}

// Valid claims verification
func (a *JWTAccessClaims) Valid() error {
	if a.ExpiresAt == 0 {
		return nil
	}
	if time.Unix(a.ExpiresAt, 0).Before(time.Now()) {
		return errors.ErrInvalidAccessToken
	}
	return nil
}

// NewJWTAccessGenerate create to generate the jwt access token instance
func NewJWTAccessGenerate(key []byte, method jwt.SigningMethod) *JWTAccessGenerate {
	return &JWTAccessGenerate{
		SignedKey:    key,
		SignedMethod: method,
	}
}

// JWTAccessGenerate generate the jwt access token
type JWTAccessGenerate struct {
	SignedKey    []byte
	SignedMethod jwt.SigningMethod
}

// Token based on the UUID generated token
func (a *JWTAccessGenerate) Token(data *oauth2.GenerateBasic, isGenRefresh bool) (string, string, error) {
	// payload from request body
	var payload apistructs.OpenapiOAuth2TokenPayload
	if err := json.NewDecoder(data.Request.Body).Decode(&payload); err != nil && err != io.EOF {
		return "", "", fmt.Errorf("failed to json decode payload from request body, err: %v", err)
	}

	// expire_in
	if payload.AccessTokenExpiredIn != "" {
		expiredIn, err := time.ParseDuration(payload.AccessTokenExpiredIn)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse accessTokenExpiredIn from request body, err: %v", err)
		}
		// oauth2 框架没有为 client_credentials 类型提供 per_token 级别的 handler，只能在每个 token generator 里修改
		data.TokenInfo.SetAccessExpiresIn(expiredIn)
	}

	// jwt payload
	claims := &JWTAccessClaims{
		StandardClaims: jwt.StandardClaims{
			Audience: data.Client.GetID(),
			Subject:  data.UserID,
			ExpiresAt: func() int64 {
				if data.TokenInfo.GetAccessExpiresIn() == 0 {
					return 0
				}
				return data.TokenInfo.GetAccessCreateAt().Add(data.TokenInfo.GetAccessExpiresIn()).Unix()
			}(),
		},
	}
	claims.Payload = payload

	token := jwt.NewWithClaims(a.SignedMethod, claims)
	var key interface{}
	if a.isEs() {
		v, err := jwt.ParseECPrivateKeyFromPEM(a.SignedKey)
		if err != nil {
			return "", "", err
		}
		key = v
	} else if a.isRsOrPS() {
		v, err := jwt.ParseRSAPrivateKeyFromPEM(a.SignedKey)
		if err != nil {
			return "", "", err
		}
		key = v
	} else if a.isHs() {
		key = a.SignedKey
	} else {
		return "", "", errs.New("unsupported sign method")
	}

	access, err := token.SignedString(key)
	if err != nil {
		return "", "", err
	}
	refresh := ""

	if isGenRefresh {
		refresh = base64.URLEncoding.EncodeToString(uuid.NewSHA1(uuid.Must(uuid.NewRandom()), []byte(access)).Bytes())
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}

	return access, refresh, nil
}

func (a *JWTAccessGenerate) isEs() bool {
	return strings.HasPrefix(a.SignedMethod.Alg(), "ES")
}

func (a *JWTAccessGenerate) isRsOrPS() bool {
	isRs := strings.HasPrefix(a.SignedMethod.Alg(), "RS")
	isPs := strings.HasPrefix(a.SignedMethod.Alg(), "PS")
	return isRs || isPs
}

func (a *JWTAccessGenerate) isHs() bool {
	return strings.HasPrefix(a.SignedMethod.Alg(), "HS")
}

func ParseJWTAccess(access string) (*JWTAccessClaims, error) {
	var claims JWTAccessClaims
	token, err := jwt.ParseWithClaims(access, &claims, func(t *jwt.Token) (key interface{}, err error) {
		key = []byte(JWTKey)
		return
	})
	if err != nil {
		logrus.Errorf("failed to parse jwt access token, access: %s, err: %v", access, err)
		return nil, fmt.Errorf("invalid access token")
	}

	if !token.Valid {
		return nil, err
	}
	return &claims, nil
}
