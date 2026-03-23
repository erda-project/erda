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

package context

import (
	"fmt"
	"net/http"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func findModel(req *http.Request, client *clientpb.Client) (*modelpb.Model, error) {
	identifier, err := findModelIdentifier(req)
	if err != nil {
		return nil, fmt.Errorf("failed to find model: %v", err)
	}
	if identifier == "" {
		return nil, fmt.Errorf("missing model")
	}

	ctx := req.Context()
	trace, instance, err := routeToModelInstance(ctx, client.Id, identifier, req.Header)
	if err != nil {
		return nil, err
	}

	// render model by template for downstream usage
	model, err := cachehelpers.GetRenderedModelByID(ctx, instance.ModelWithProvider.Id)
	if err != nil {
		return nil, err
	}

	ctxhelper.PutPolicyTrace(ctx, trace)

	return model, nil
}
