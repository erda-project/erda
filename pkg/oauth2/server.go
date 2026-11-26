// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oauth2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"

	"github.com/erda-project/erda/pkg/oauth2/clientstore/mysqlclientstore"
	"github.com/erda-project/erda/pkg/oauth2/tokenstore/mysqltokenstore"
)

type OAuth2Server struct {
	srv         *server.Server
	clientStore *mysqlclientstore.ClientStore
}

func (o *OAuth2Server) Server() *server.Server {
	return o.srv
}

func NewOAuth2Server() *OAuth2Server {
	manager := manage.NewDefaultManager()
	manager.SetClientTokenCfg(&manage.Config{
		AccessTokenExp:    time.Hour * 1,
		RefreshTokenExp:   0,
		IsGenerateRefresh: false,
	})

	clientStore, err := mysqlclientstore.NewClientStore()
	if err != nil {
		panic(err)
	}
	manager.MustClientStorage(clientStore, nil)
	manager.MustTokenStorage(mysqltokenstore.NewTokenStore(mysqltokenstore.WithTokenStoreGCInterval(time.Second * 30)))

	// jwt token generate
	manager.MapAccessGenerate(NewJWTAccessGenerate([]byte(JWTKey), jwt.SigningMethodHS512))

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(false) // POST
	srv.SetClientInfoHandler(server.ClientFormHandler)

	// logger
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		logrus.Errorf("oauth2 server internal err: %v", err)
		re = &errors.Response{Error: err}
		return
	})
	srv.SetResponseErrorHandler(func(re *errors.Response) {
		logrus.Errorf("oauth2 server response err: %v", re.Error.Error())
	})

	return &OAuth2Server{
		srv:         srv,
		clientStore: clientStore,
	}
}

// {{openapi}}/oauth2/token?grant_type=client_credentials&client_id=<client-id>&client_secret=<client-secret>&scope=read
func (o *OAuth2Server) Token(w http.ResponseWriter, r *http.Request) {
	o.rewriteExpiredPreviousSecret(r)
	err := o.srv.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// rewriteExpiredPreviousSecret keeps the old client secret valid during a rotation window.
// The oauth2 client store only exposes one secret, so requests using previous_secret are
// rewritten to the current secret before the library validates the client.
func (o *OAuth2Server) rewriteExpiredPreviousSecret(r *http.Request) {
	if o == nil || o.clientStore == nil {
		return
	}
	clientID, clientSecret, source := extractClientCredentials(r)
	if clientID == "" || clientSecret == "" {
		return
	}
	stored, err := o.clientStore.GetItemByID(clientID)
	if err != nil {
		return
	}
	if stored == nil || stored.PreviousSecret == "" || stored.PreviousSecret != clientSecret {
		return
	}
	if stored.PreviousSecretExpiresAt == nil || time.Now().After(*stored.PreviousSecretExpiresAt) {
		return
	}
	switch source {
	case "basic":
		r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+stored.Secret)))
	case "query":
		values := r.URL.Query()
		values.Set("client_secret", stored.Secret)
		r.URL.RawQuery = values.Encode()
	case "form":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return
		}
		values.Set("client_secret", stored.Secret)
		rewritten := []byte(values.Encode())
		r.Body = io.NopCloser(bytes.NewReader(rewritten))
		r.ContentLength = int64(len(rewritten))
		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(rewritten)), nil
		}
	}
}

func extractClientCredentials(r *http.Request) (clientID, clientSecret, source string) {
	if username, password, ok := r.BasicAuth(); ok {
		return username, password, "basic"
	}
	if id := r.URL.Query().Get("client_id"); id != "" {
		return id, r.URL.Query().Get("client_secret"), "query"
	}
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/x-www-form-urlencoded") && r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return "", "", ""
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return "", "", ""
		}
		return values.Get("client_id"), values.Get("client_secret"), "form"
	}
	return "", "", ""
}

// {{openapi}}/oauth2/invalidate_token?access_token=xxx
func (o *OAuth2Server) InvalidateToken(w http.ResponseWriter, r *http.Request) {
	ti, err := o.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// remote from store
	err = o.srv.Manager.RemoveAccessToken(ti.GetAccess())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(o, w, ti)
}

// {{openapi}}/oauth2/validate_token?access_token=xxx
func (o *OAuth2Server) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ti, err := o.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(o, w, ti)
}

func writeJSON(o *OAuth2Server, w http.ResponseWriter, tokenInfo oauth2.TokenInfo) {
	tokenData := o.srv.GetTokenData(tokenInfo)
	expiresAt := tokenInfo.GetAccessCreateAt().Add(tokenInfo.GetAccessExpiresIn())
	tokenData["expiresAt"] = expiresAt.Format(time.RFC3339)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(&tokenData)
	w.Write(b)
}
