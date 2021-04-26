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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (e *Endpoints) createPipelineQueue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// validate request
	if r.ContentLength == 0 {
		return apierrors.ErrCreatePipelineQueue.MissingParameter("request body").ToResp(), nil
	}

	// decode request
	var req apistructs.PipelineQueueCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreatePipelineQueue.InvalidParameter(fmt.Errorf("failed to unmarshal request body, err: %v", err)).ToResp(), nil
	}

	// check authentication
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrCreatePipelineQueue.AccessDenied().ToResp(), nil
	}

	// do create
	queue, err := e.queueManage.CreatePipelineQueue(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(queue)
}

func (e *Endpoints) getPipelineQueue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// validate request
	queueIDStr := vars[pathQueueID]
	queueID, err := strconv.ParseUint(queueIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetPipelineQueue.InvalidParameter(fmt.Errorf("invalid queueID: %s, err: %v", queueIDStr, err)).ToResp(), nil
	}

	// check authentication
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrGetPipelineQueue.AccessDenied().ToResp(), nil
	}

	// do get
	queue, err := e.queueManage.GetPipelineQueue(queueID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// set usage
	queue.Usage = e.reconciler.QueueManager.QueryQueueUsage(queue)

	return httpserver.OkResp(queue)
}

func (e *Endpoints) pagingPipelineQueues(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// handle request
	var req apistructs.PipelineQueuePagingRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingPipelineQueues.InvalidParameter(err).ToResp(), nil
	}

	// check authentication
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrPagingPipelineQueues.AccessDenied().ToResp(), nil
	}

	// do paging
	queues, err := e.queueManage.PagingPipelineQueues(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(queues)
}

func (e *Endpoints) updatePipelineQueue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// validate request
	if r.ContentLength == 0 {
		return httpserver.OkResp(nil)
	}

	// handle request
	var req apistructs.PipelineQueueUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdatePipelineQueue.InvalidParameter(err).ToResp(), nil
	}

	// get queue id
	queueIDStr := vars[pathQueueID]
	queueID, err := strconv.ParseUint(queueIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrUpdatePipelineQueue.InvalidParameter(fmt.Errorf("invalid queueID: %s, err: %v", queueIDStr, err)).ToResp(), nil
	}
	req.ID = queueID

	// check authentication
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrUpdatePipelineQueue.AccessDenied().ToResp(), nil
	}

	// do update
	queue, err := e.queueManage.UpdatePipelineQueue(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// update queue in manager
	e.reconciler.QueueManager.IdempotentAddQueue(queue)

	return httpserver.OkResp(queue)
}

func (e *Endpoints) deletePipelineQueue(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// get queue id
	queueIDStr := vars[pathQueueID]
	queueID, err := strconv.ParseUint(queueIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrDeletePipelineQueue.InvalidParameter(fmt.Errorf("invalid queueID: %s, err: %v", queueIDStr, err)).ToResp(), nil
	}

	// check authentication
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrPagingPipelineQueues.AccessDenied().ToResp(), nil
	}

	// do update
	if err := e.queueManage.DeletePipelineQueue(queueID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}
