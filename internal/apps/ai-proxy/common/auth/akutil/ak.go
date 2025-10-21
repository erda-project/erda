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

package akutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type AKUtil struct {
	cache cachetypes.Manager
}

func New(cache cachetypes.Manager) *AKUtil {
	return &AKUtil{cache: cache}
}

func (util *AKUtil) AkToClient(ak string) (*clientpb.Client, error) {
	_, allClientsV, err := util.cache.ListAll(context.Background(), cachetypes.ItemTypeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get client by ak: %v", err)
	}
	var matchedClients []*clientpb.Client
	for _, client := range allClientsV.([]*clientpb.Client) {
		if client.AccessKeyId == ak {
			matchedClients = append(matchedClients, client)
		}
	}
	total := len(matchedClients)
	if total == 0 {
		return nil, fmt.Errorf("client not found by ak: %s", ak)
	}
	if total > 1 {
		return nil, fmt.Errorf("multiple clients found by ak: %s", ak)
	}
	return matchedClients[0], nil
}

func (util *AKUtil) GetAkFromHeader(ctx context.Context, req any) (string, bool) {
	// get from HTTP
	v, ok := util.GetAkFromHTTPHeader(req)
	if ok {
		return v, true
	}
	// get from GRPC
	v, ok = util.GetAkFromGRPCHeader(ctx)
	if ok {
		return v, true
	}
	return "", false
}

func (util *AKUtil) GetAkFromGRPCHeader(ctx context.Context) (string, bool) {
	v := apis.GetHeader(ctx, httputil.HeaderKeyAuthorization)
	v = vars.TrimBearer(v)
	return v, v != ""
}

func (util *AKUtil) GetAkFromHTTPHeader(req any) (string, bool) {
	var v string
	if req != nil {
		if r, ok := req.(*http.Request); ok {
			v = r.Header.Get(httputil.HeaderKeyAuthorization)
		}
	}
	v = vars.TrimBearer(v)
	return v, v != ""
}

func CheckAdmin(ctx context.Context, req any, cacheManager cachetypes.Manager) (bool, error) {
	ak, ok := New(cacheManager).GetAkFromHeader(ctx, req)
	if !ok || ak == "" {
		return false, handlers.ErrAkNotFound
	}
	return ak == os.Getenv(vars.EnvAIProxyAdminAuthKey), nil
}

func CheckAkOrToken(ctx context.Context, req any, cacheManager cachetypes.Manager) (*clienttokenpb.ClientToken, *clientpb.Client, error) {
	ak, ok := New(cacheManager).GetAkFromHeader(ctx, req)
	if !ok {
		return nil, nil, handlers.ErrAkNotFound
	}
	return GetClientInfo(ak, cacheManager)
}

func GetClientInfo(ak string, cacheManager cachetypes.Manager) (*clienttokenpb.ClientToken, *clientpb.Client, error) {
	// check by token
	if strings.HasPrefix(ak, client_token.TokenPrefix) {
		clientToken, client, err := New(cacheManager).TokenToClient(ak)
		if err != nil {
			return nil, nil, handlers.HTTPError(err, http.StatusUnauthorized)
		}
		return clientToken, client, nil
	}
	// check by ak
	client, err := New(cacheManager).AkToClient(ak)
	if err != nil {
		return nil, nil, handlers.HTTPError(err, http.StatusUnauthorized)
	}
	return nil, client, nil
}

func AutoCheckAndSetClientInfo(clientId string, clientTokenId string, req any, skipSet bool) error {
	reqValue, ok := structValue(req)
	if !ok {
		return nil
	}

	clientIdField, ok := findStringFieldCaseInsensitive(reqValue, "ClientID", "client_id")
	if ok {
		currentClientId := clientIdField.String()
		if currentClientId != "" && currentClientId != clientId {
			return handlers.ErrClientIdParamMismatch
		}
		if !skipSet && clientIdField.CanSet() {
			clientIdField.SetString(clientId)
		}
	}
	// check client token id
	if clientTokenId != "" {
		clientTokenIdField, ok := findStringFieldCaseInsensitive(reqValue, "ClientTokenID", "client_token_id")
		if ok {
			currentClientTokenId := clientTokenIdField.String()
			if currentClientTokenId != "" && currentClientTokenId != clientTokenId {
				return handlers.ErrTokenIdParamMismatch
			}
			if !skipSet && clientTokenIdField.CanSet() {
				clientTokenIdField.SetString(clientTokenId)
			}
		}
	}
	return nil
}

func structValue(req any) (reflect.Value, bool) {
	if req == nil {
		return reflect.Value{}, false
	}
	value := reflect.ValueOf(req)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return reflect.Value{}, false
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	return value, true
}

func findStringFieldCaseInsensitive(structValue reflect.Value, names ...string) (reflect.Value, bool) {
	for _, name := range names {
		field := structValue.FieldByName(name)
		if field.IsValid() && field.Kind() == reflect.String {
			return field, true
		}
	}

	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structType.Field(i)
		if fieldType.PkgPath != "" {
			continue
		}
		for _, name := range names {
			if strings.EqualFold(fieldType.Name, name) {
				fieldValue := structValue.Field(i)
				if fieldValue.Kind() == reflect.String {
					return fieldValue, true
				}
			}
		}
	}
	return reflect.Value{}, false
}
