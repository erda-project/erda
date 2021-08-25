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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/qaparser/types"

	"github.com/erda-project/erda/modules/dop/dbclient"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

const (
	ErrMsgURLMissingPathApiID = "URL PATH 缺少参数: apiID"
)

func (e *Endpoints) GetTestTypes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	return httpserver.OkResp(types.TestTypeValues())
}

func (e *Endpoints) TestCallback(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	if r.ContentLength == 0 {
		return apierrors.ErrDoTestCallback.MissingParameter(apierrors.MissingRequestBody).ToResp(), nil
	}
	var req apistructs.TestCallBackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDoTestCallback.InvalidParameter(err).ToResp(), nil
	}

	qaID, err := storeTestResults(&req)
	if err != nil {
		return apierrors.ErrDoTestCallback.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(qaID)
}

func (e *Endpoints) GetRecords(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.TestRecordPagingRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingTestRecords.InvalidParameter(err).ToResp(), nil
	}

	if req.PageNo == 0 {
		req.PageNo = 0
	}
	if req.PageSize == 0 {
		req.PageSize = 15
	}

	pagingResult, err := dbclient.FindTPRecordPagingByAppID(req)
	if err != nil {
		return apierrors.ErrPagingTestRecords.InternalError(err).ToResp(), nil
	}

	for _, r := range pagingResult.List.([]*dbclient.TPRecordDO) {
		// erase sensitive information
		r.EraseSensitiveInfo()
	}

	return httpserver.OkResp(pagingResult)
}

func (e *Endpoints) GetTestRecord(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	recordID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrGetTestRecord.InvalidParameter(fmt.Errorf("invalid id: %v", err)).ToResp(), nil
	}

	record, err := dbclient.FindTPRecordById(recordID)
	if err != nil {
		return apierrors.ErrGetTestRecord.InternalError(err).ToResp(), nil
	}

	record.EraseSensitiveInfo()

	return httpserver.OkResp(record)
}

func storeTestResults(testResults *apistructs.TestCallBackRequest) (string, error) {
	tpRecord := convertTestRecords(testResults)
	if _, err := dbclient.InsertTPRecord(tpRecord); err != nil {
		return "", err
	}

	return strconv.FormatUint(tpRecord.ID, 10), nil
}

func convertTestRecords(results *apistructs.TestCallBackRequest) *dbclient.TPRecordDO {
	return &dbclient.TPRecordDO{
		Suites:          results.Suites,
		Totals:          results.Totals,
		ParserType:      types.TestParserType(results.Results.Type),
		ApplicationID:   results.Results.ApplicationID,
		ProjectID:       results.Results.ProjectID,
		BuildID:         results.Results.BuildID,
		Name:            results.Results.Name,
		Branch:          results.Results.Branch,
		GitRepo:         results.Results.GitRepo,
		OperatorID:      results.Results.OperatorID,
		TType:           apistructs.TestType(results.Results.Type),
		Workspace:       apistructs.DiceWorkspace(results.Results.Workspace),
		CommitID:        results.Results.CommitID,
		OperatorName:    results.Results.OperatorName,
		ApplicationName: results.Results.ApplicationName,
		Extra:           results.Results.Extra,
		UUID:            results.Results.UUID,
	}
}
