package bundle

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

func (b *Bundle) QueryPublishItems(req *apistructs.QueryPublishItemRequest) (*apistructs.QueryPublishItemData, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var getResp apistructs.QueryPublishItemResponse
	resp, err := hc.Get(host).Path("/api/publish-items").
		Param("pageSize", strconv.FormatInt(req.PageSize, 10)).
		Param("type", req.Type).
		Param("ids", req.Ids).
		Param("q", req.Q).
		Param("pageNo", strconv.FormatInt(req.PageNo, 10)).
		Param("publisherId", strconv.FormatInt(req.PublisherId, 10)).
		Param("name", req.Name).
		Header("Internal-Client", "bundle").
		Header("Org-ID", strconv.FormatInt(req.OrgID, 10)).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return &getResp.Data, nil
}
