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

package bundle

import (
	"bytes"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) CreateNoticeRequest(req *apistructs.NoticeCreateRequest,
	orgID uint64) (*apistructs.NoticeCreateResponse, error) {
	cmdbURL, err := b.urls.CoreServices()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	var ncresp apistructs.NoticeCreateResponse
	httpClient := b.hc
	resp, err := httpClient.Post(cmdbURL).Path("/api/notices").
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, fmt.Sprintf("%d", orgID)).
		Header(httputil.UserHeader, req.UserID).
		JSONBody(&req).
		Do().
		JSON(&ncresp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to create notice, status code: %d, body: %v",
				resp.StatusCode(),
				string(resp.Body()),
			))
	}
	return &ncresp, nil
}

func (b *Bundle) UpdateNotice(req *apistructs.NoticeUpdateRequest, noticeID, orgID uint64, userID string) (
	*apistructs.NoticeUpdateResponse, error) {
	cmdbURL, err := b.urls.CoreServices()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	httpClient := b.hc

	var ncresp apistructs.NoticeUpdateResponse
	resp, err := httpClient.Put(cmdbURL).Path(fmt.Sprintf("/api/notices/%d", noticeID)).
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, fmt.Sprintf("%d", orgID)).
		Header(httputil.UserHeader, userID).
		JSONBody(&req).
		Do().
		JSON(&ncresp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to update notice, status code: %d, body: %v",
				resp.StatusCode(),
				string(resp.Body()),
			))
	}

	return &ncresp, nil
}

func (b *Bundle) DeleteNotice(noticeID, orgID uint64, userID string) (*apistructs.NoticeDeleteResponse, error) {
	cmdbURL, err := b.urls.CoreServices()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	httpClient := b.hc

	var ncresp apistructs.NoticeDeleteResponse
	resp, err := httpClient.Delete(cmdbURL).Path(fmt.Sprintf("/api/notices/%d", noticeID)).
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, fmt.Sprintf("%d", orgID)).
		Header(httputil.UserHeader, userID).
		Do().
		JSON(&ncresp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to update notice, status code: %d, body: %v",
				resp.StatusCode(),
				string(resp.Body()),
			))
	}

	return &ncresp, nil
}

func (b *Bundle) PublishORUnPublishNotice(orgID uint64, noticeID uint64, userID, publishType string) error {
	cmdbURL, err := b.urls.CoreServices()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	var buf bytes.Buffer
	resp, err := b.hc.Put(cmdbURL).Path(fmt.Sprintf("/api/notices/%d/actions/%s", noticeID, publishType)).
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, fmt.Sprintf("%d", orgID)).
		Header(httputil.UserHeader, userID).
		Do().
		Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to %s notice, status code: %d, body: %v",
				publishType,
				resp.StatusCode(),
				buf.String(),
			))
	}
	return nil
}

func (b *Bundle) ListNoticeByOrgID(orgID uint64, userID string) (*apistructs.NoticeListResponse, error) {
	cmdbURL, err := b.urls.CoreServices()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	var noteList apistructs.NoticeListResponse
	resp, err := b.hc.Get(cmdbURL).Path("/api/notices").
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, fmt.Sprintf("%d", orgID)).
		Header(httputil.UserHeader, userID).
		Do().
		JSON(&noteList)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to list notice, status code: %d, body: %v",
				resp.StatusCode(),
				string(resp.Body()),
			))
	}

	return &noteList, nil
}
