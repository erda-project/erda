package bundle

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

func (b *Bundle) GetNexusOrgDockerCredentialByImage(orgID uint64, image string) (*apistructs.NexusUser, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.NexusUserGetResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/orgs/%d/actions/get-nexus-docker-credential-by-image", orgID)).
		Param("image", image).
		Header("Internal-Client", "bundle").
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return getResp.Data, nil
}
