package bundle

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

// DoRemoteAction 代理完成阿里云client action
func (b *Bundle) DoRemoteAliyunAction(orgId, clusterName, endpointType string, endpointMap map[string]string, request requests.AcsRequest, response responses.AcsResponse) error {
	req := &apistructs.RemoteActionRequest{
		OrgID:                orgId,
		ClusterName:          clusterName,
		Product:              request.GetProduct(),
		Version:              request.GetVersion(),
		ActionName:           request.GetActionName(),
		LocationServiceCode:  request.GetLocationServiceCode(),
		LocationEndpointType: request.GetLocationEndpointType(),
		EndpointType:         endpointType,
		EndpointMap:          endpointMap,
	}
	err := requests.InitParams(request)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Scheme = request.GetScheme()
	if req.Scheme == "" {
		req.Scheme = "https"
	}
	req.QueryParams = request.GetQueryParams()
	req.Headers = request.GetHeaders()
	req.FormParams = request.GetFormParams()
	host, err := b.urls.Ops()
	if err != nil {
		return errors.WithStack(err)
	}
	resp, err := b.hc.Post(host).Path("/api/aliyun-client").JSONBody(req).Do().RAW()
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()
	err = responses.Unmarshal(response, resp, request.GetAcceptFormat())
	if err != nil {
		return errors.WithStack(err)
	}
	logrus.Debugf("DoRemoteAction request:%+v, response status:%d, body:%s", req,
		response.GetHttpStatus(),
		response.GetHttpContentString())
	return nil
}

// GetOrgAccount 获取云账号信息
func (b *Bundle) GetOrgAccount(orgId, vendor string) (*apistructs.CloudAccount, error) {
	host, err := b.urls.Ops()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var resp apistructs.CloudAccountResponse
	r, err := b.hc.Get(host).Path("/api/internal-cloud-account").Param("orgID", orgId).Param("vendor", vendor).Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "find cloud account failed"})
	}
	return &resp.Data, nil
}
