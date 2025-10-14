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
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type AKUtil struct {
	Dao dao.DAO
}

func New(dao dao.DAO) *AKUtil {
	return &AKUtil{Dao: dao}
}

func (util *AKUtil) AkToClient(ak string) (*clientpb.Client, error) {
	pagingResp, err := util.Dao.ClientClient().Paging(context.Background(), &clientpb.ClientPagingRequest{
		PageSize:     1,
		PageNum:      1,
		AccessKeyIds: []string{ak},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get client by ak: %v", err)
	}
	if pagingResp.Total == 0 {
		return nil, fmt.Errorf("client not found by ak: %s", ak)
	}
	if pagingResp.Total > 1 {
		return nil, fmt.Errorf("multiple clients found by ak: %s", ak)
	}
	return pagingResp.List[0], nil
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

func CheckAdmin(ctx context.Context, req any, dao dao.DAO) (bool, error) {
	ak, ok := New(dao).GetAkFromHeader(ctx, req)
	if !ok || ak == "" {
		return false, handlers.ErrAkNotFound
	}
	return ak == os.Getenv(vars.EnvAIProxyAdminAuthKey), nil
}

func CheckAkOrToken(ctx context.Context, req any, dao dao.DAO) (*clienttokenpb.ClientToken, *clientpb.Client, error) {
	ak, ok := New(dao).GetAkFromHeader(ctx, req)
	if !ok {
		return nil, nil, handlers.ErrAkNotFound
	}
	return GetClientInfo(ak, dao)
}

func GetClientInfo(ak string, dao dao.DAO) (*clienttokenpb.ClientToken, *clientpb.Client, error) {
	// check by token
	if strings.HasPrefix(ak, client_token.TokenPrefix) {
		clientToken, client, err := New(dao).TokenToClient(ak)
		if err != nil {
			return nil, nil, handlers.HTTPError(err, http.StatusUnauthorized)
		}
		return clientToken, client, nil
	}
	// check by ak
	client, err := New(dao).AkToClient(ak)
	if err != nil {
		return nil, nil, handlers.HTTPError(err, http.StatusUnauthorized)
	}
	return nil, client, nil
}

func AutoCheckAndSetClientId(clientId string, req any, skipSet bool) error {
	// use reflect to set req's clientId field if have
	clientIdField := reflect.ValueOf(req).Elem().FieldByName("ClientId")
	if clientIdField != (reflect.Value{}) {
		// compare if ClientId already exists
		currentClientId := clientIdField.String()
		if currentClientId != "" && currentClientId != clientId {
			return handlers.ErrAkNotMatch
		}
		if !skipSet {
			clientIdField.SetString(clientId)
		}
	}
	return nil
}
