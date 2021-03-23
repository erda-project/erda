package aliyunclient

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/pkg/errors"
)

type RemoteActionRequest struct {
	ClusterName          string
	Product              string
	Version              string
	ActionName           string
	LocationServiceCode  string
	LocationEndpointType string
	QueryParams          map[string]string
	Headers              map[string]string
	FormParams           map[string]string
}

func NewRemoteActionRequest(acsReq requests.AcsRequest) (*RemoteActionRequest, error) {
	req := &RemoteActionRequest{
		Product:              acsReq.GetProduct(),
		Version:              acsReq.GetVersion(),
		ActionName:           acsReq.GetActionName(),
		LocationServiceCode:  acsReq.GetLocationServiceCode(),
		LocationEndpointType: acsReq.GetLocationEndpointType(),
	}
	err := requests.InitParams(acsReq)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.QueryParams = acsReq.GetQueryParams()
	req.Headers = acsReq.GetHeaders()
	req.FormParams = acsReq.GetFormParams()
	return req, nil
}
