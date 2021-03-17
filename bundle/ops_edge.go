package bundle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
)

const (
	edgeAppURL           = "/api/edge/app"
	edgeSiteURL          = "/api/edge/site"
	edgeConfigsetURL     = "/api/edge/configset"
	edgeCfgSetItemURL    = "/api/edge/configset/item"
	schedulerClusterInfo = "/api/clusterinfo/%s"
	edgeInstanceInfo     = "/api/instanceinfo"
	edgeSiteInitURL      = "/api/edge/site/init"
)

func (b *Bundle) ListEdgeApp(req *apistructs.EdgeAppListPageRequest) (*apistructs.EdgeAppListResponse, error) {
	var (
		httpReqRes httpserver.Resp
		res        apistructs.EdgeAppListResponse
		buffer     bytes.Buffer
		reqParam   map[string]string
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	hcReq := b.hc.
		Get(host).
		Path(edgeAppURL)

	// Ops Edge 默认分页, Size 20, PageNo 1
	if req.PageNo == 0 && req.PageSize == 0 {
		reqParam = map[string]string{
			"orgID": strconv.Itoa(int(req.OrgID)),
		}
	} else {
		reqParam = map[string]string{
			"pageNo":   strconv.Itoa(req.PageNo),
			"pageSize": strconv.Itoa(req.PageSize),
			"orgID":    strconv.Itoa(int(req.OrgID)),
		}
	}

	for key, value := range reqParam {
		hcReq.Param(key, value)
	}

	httpResp, err := hcReq.Do().Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (list edge application) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	resJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resJson, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (b *Bundle) GetEdgeApp(appID int64) (*apistructs.EdgeAppInfo, error) {
	var (
		res        apistructs.EdgeAppInfo
		buffer     bytes.Buffer
		httpReqRes httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	httpResp, err := b.hc.
		Get(host).
		Path(fmt.Sprintf(edgeAppURL+"/%d", appID)).Do().Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (list edge application) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	resJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resJson, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (b *Bundle) CreateEdgeApp(req *apistructs.EdgeAppCreateRequest) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Post(host).
		Path(edgeAppURL).
		JSONBody(req).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) DeleteEdgeApp(appID int64) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Delete(host).
		Path(fmt.Sprintf(edgeAppURL+"/%d", appID)).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) GetEdgeAppStatus(req *apistructs.EdgeAppStatusListRequest) (*apistructs.EdgeAppStatusResponse, error) {
	var (
		res        apistructs.EdgeAppStatusResponse
		buffer     bytes.Buffer
		httpReqRes httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	reqClient := b.hc.
		Get(host).
		Path(fmt.Sprintf(edgeAppURL+"/status/%d", req.AppID))

	if req.NotPaging {
		reqClient.Param("notPaging", "true")
	} else {
		reqClient.Param("pageNo", strconv.Itoa(req.PageNo))
		reqClient.Param("pageSize", strconv.Itoa(req.PageSize))
	}

	httpResp, err := reqClient.Do().Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (list edge application) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	resJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resJson, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (b *Bundle) ListEdgeSite(req *apistructs.EdgeSiteListPageRequest) (*apistructs.EdgeSiteListResponse, error) {
	var (
		httpReqRes httpserver.Resp
		res        apistructs.EdgeSiteListResponse
		buffer     bytes.Buffer
		reqParam   map[string]string
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	hcReq := b.hc.
		Get(host).
		Path(edgeSiteURL)

	if req.NotPaging {
		reqParam = map[string]string{
			"orgID":     strconv.Itoa(int(req.OrgID)),
			"notPaging": strconv.FormatBool(req.NotPaging),
		}
		// 如果不指定采用Ops默认分页参数
	} else if req.PageNo == 0 && req.PageSize == 0 {
		reqParam = map[string]string{
			"orgID": strconv.Itoa(int(req.OrgID)),
		}
	} else {
		reqParam = map[string]string{
			"pageNo":   strconv.Itoa(req.PageNo),
			"pageSize": strconv.Itoa(req.PageSize),
			"orgID":    strconv.Itoa(int(req.OrgID)),
		}
	}

	for key, value := range reqParam {
		hcReq.Param(key, value)
	}

	httpResp, err := hcReq.Do().Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (list edge site) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	resJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resJson, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (b *Bundle) GetEdgeSite(siteID int64) (*apistructs.EdgeSiteInfo, error) {
	var (
		res        apistructs.EdgeSiteInfo
		buffer     bytes.Buffer
		httpReqRes httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	httpResp, err := b.hc.
		Get(host).
		Path(fmt.Sprintf(edgeSiteURL+"/%d", siteID)).Do().Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (list edge application) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	resJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resJson, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (b *Bundle) CreateEdgeSite(req *apistructs.EdgeSiteCreateRequest) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Post(host).
		Path(edgeSiteURL).
		JSONBody(req).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) UpdateEdgeSite(req *apistructs.EdgeSiteUpdateRequest, siteID int64) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Put(host).
		Path(fmt.Sprintf(edgeSiteURL+"/%d", siteID)).
		JSONBody(req).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) DeleteEdgeSite(siteID int64) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Delete(host).
		Path(fmt.Sprintf(edgeSiteURL+"/%d", siteID)).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) ListEdgeConfigset(req *apistructs.EdgeConfigSetListPageRequest) (*apistructs.EdgeConfigSetListResponse, error) {
	var (
		httpReqRes httpserver.Resp
		res        apistructs.EdgeConfigSetListResponse
		buffer     bytes.Buffer
		reqParam   map[string]string
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	hcReq := b.hc.
		Get(host).
		Path(edgeConfigsetURL)

	if req.NotPaging {
		reqParam = map[string]string{
			"orgID":     strconv.Itoa(int(req.OrgID)),
			"notPaging": strconv.FormatBool(req.NotPaging),
		}
	} else if req.PageNo == 0 && req.PageSize == 0 {
		reqParam = map[string]string{
			"orgID": strconv.Itoa(int(req.OrgID)),
		}
	} else {
		reqParam = map[string]string{
			"pageNo":   strconv.Itoa(req.PageNo),
			"pageSize": strconv.Itoa(req.PageSize),
			"orgID":    strconv.Itoa(int(req.OrgID)),
		}
	}

	for key, value := range reqParam {
		hcReq.Param(key, value)
	}

	httpResp, err := hcReq.Do().Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (list config set) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	resJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resJson, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (b *Bundle) CreateEdgeConfigset(req *apistructs.EdgeConfigSetCreateRequest) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Post(host).
		Path(edgeConfigsetURL).
		JSONBody(req).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) UpdateEdgeConfigset(req *apistructs.EdgeConfigSetUpdateRequest, siteID int64) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Put(host).
		Path(fmt.Sprintf(edgeConfigsetURL+"/%d", siteID)).
		JSONBody(req).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) DeleteEdgeConfigset(siteID int64) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Delete(host).
		Path(fmt.Sprintf(edgeConfigsetURL+"/%d", siteID)).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) ListEdgeCfgSetItem(req *apistructs.EdgeCfgSetItemListPageRequest) (*apistructs.EdgeCfgSetItemListResponse, error) {
	var (
		httpReqRes httpserver.Resp
		res        apistructs.EdgeCfgSetItemListResponse
		buffer     bytes.Buffer
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	hcReq := b.hc.
		Get(host).
		Path(edgeCfgSetItemURL)

	reqParam := map[string]string{
		"pageNo":      strconv.Itoa(req.PageNo),
		"pageSize":    strconv.Itoa(req.PageSize),
		"configSetID": strconv.Itoa(int(req.ConfigSetID)),
	}

	if req.Search != "" {
		reqParam["search"] = req.Search
	}

	for key, value := range reqParam {
		hcReq.Param(key, value)
	}

	httpResp, err := hcReq.Do().Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (list config set item) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	resJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resJson, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (b *Bundle) GetEdgeCfgSetItem(itemID int64) (*apistructs.EdgeCfgSetItemInfo, error) {
	var (
		res        apistructs.EdgeCfgSetItemInfo
		buffer     bytes.Buffer
		httpReqRes httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	httpResp, err := b.hc.
		Get(host).
		Path(fmt.Sprintf(edgeCfgSetItemURL+"/%d", itemID)).Do().Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (get configset item) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	resJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resJson, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (b *Bundle) CreateEdgeCfgSetItem(req *apistructs.EdgeCfgSetItemCreateRequest) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Post(host).
		Path(edgeCfgSetItemURL).
		JSONBody(req).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) UpdateEdgeCfgSetItem(req *apistructs.EdgeCfgSetItemUpdateRequest, cfgSetItemID int64) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Put(host).
		Path(fmt.Sprintf(edgeCfgSetItemURL+"/%d", cfgSetItemID)).
		JSONBody(req).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) DeleteEdgeCfgSetItem(siteID int64) error {
	var (
		resp httpserver.Resp
	)

	host, err := b.urls.Ops()
	if err != nil {
		return err
	}

	httpResp, err := b.hc.
		Delete(host).
		Path(fmt.Sprintf(edgeCfgSetItemURL+"/%d", siteID)).
		Do().
		JSON(&resp)

	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Err)
	}

	return nil
}

func (b *Bundle) ListEdgeCluster(orgID uint64) ([]map[string]interface{}, error) {
	var (
		edgeClusters = make([]map[string]interface{}, 0)
		edgeCloudKey = "IS_EDGE_CLOUD"
	)

	res, err := b.ListClusters("", orgID)
	if err != nil {
		return edgeClusters, err
	}

	for _, value := range res {
		res, err := b.GetClusterInfo(value.Name)
		if err != nil {
			logrus.Errorf("get cluster %s info error: %v", value.Name, err)
			continue
		}
		if _, ok := res[edgeCloudKey]; ok {
			edgeClusters = append(edgeClusters, map[string]interface{}{
				"name":  value.Name,
				"value": value.ID,
			})
		}
	}

	return edgeClusters, nil
}

func (b *Bundle) ListEdgeSelectSite(orgID int64, valueType string) ([]map[string]interface{}, error) {
	sites := make([]map[string]interface{}, 0)

	res, err := b.ListEdgeSite(&apistructs.EdgeSiteListPageRequest{
		OrgID:     orgID,
		NotPaging: true,
	})

	if err != nil {
		return nil, err
	}

	if valueType == apistructs.EdgeListValueTypeID {
		for _, value := range res.List {
			sites = append(sites, map[string]interface{}{
				"name":  value.Name,
				"value": value.ID,
			})
		}
	} else if valueType == apistructs.EdgeListValueTypeName {
		for _, value := range res.List {
			sites = append(sites, map[string]interface{}{
				"name":  value.Name,
				"value": value.Name,
			})
		}
	}

	return sites, nil
}

func (b *Bundle) ListEdgeSelectConfigSet(orgID int64, valueType string) ([]map[string]interface{}, error) {
	configSets := make([]map[string]interface{}, 0)

	res, err := b.ListEdgeConfigset(&apistructs.EdgeConfigSetListPageRequest{
		OrgID:     orgID,
		NotPaging: true,
	})

	if err != nil {
		return nil, err
	}

	if valueType == apistructs.EdgeListValueTypeID {
		for _, value := range res.List {
			configSets = append(configSets, map[string]interface{}{
				"name":  value.Name,
				"value": value.ID,
			})
		}
	} else if valueType == apistructs.EdgeListValueTypeName {
		for _, value := range res.List {
			configSets = append(configSets, map[string]interface{}{
				"name":  value.Name,
				"value": value.Name,
			})
		}
	}

	return configSets, nil
}

func (b *Bundle) GetEdgeSiteInitShell(siteID int64) (map[string]interface{}, error) {
	var (
		httpReqRes httpserver.Resp
		buffer     bytes.Buffer
	)

	host, err := b.urls.Ops()
	if err != nil {
		return nil, err
	}

	httpResp, err := b.hc.
		Get(host).
		Path(fmt.Sprintf(edgeSiteInitURL+"/%d", siteID)).
		Do().
		Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request ops api (get site init shell) not found")
	}

	fmt.Println(httpResp.Body())

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	if shell, ok := httpReqRes.Data.(map[string]interface{}); ok {
		return shell, nil
	} else {
		return nil, fmt.Errorf("response data format error")
	}
}

func (b *Bundle) GetClusterInfo(clusterName string) (map[string]interface{}, error) {
	var (
		httpReqRes httpserver.Resp
		buffer     bytes.Buffer
	)

	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}

	httpResp, err := b.hc.
		Get(host).
		Path(fmt.Sprintf(schedulerClusterInfo, clusterName)).
		Do().
		Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request scheduler api (get cluster info) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	if !httpReqRes.Success {
		return nil, fmt.Errorf(httpReqRes.Err.Msg)
	}

	if info, ok := httpReqRes.Data.(map[string]interface{}); ok {
		return info, nil
	} else {
		return nil, fmt.Errorf("response data format error")
	}
}

func (b *Bundle) GetEdgeInstanceInfo(orgID int64, appName, site string) ([]apistructs.InstanceInfoData, error) {
	var (
		httpReqRes httpserver.Resp
		infosData  []apistructs.InstanceInfoData
		buffer     bytes.Buffer
	)

	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}

	httpResp, err := b.hc.
		Get(host).
		Path(edgeInstanceInfo).
		Param("orgID", strconv.Itoa(int(orgID))).
		Param("edgeApplicationName", appName).
		Param("edgeSite", site).
		Do().
		Body(&buffer)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if httpResp.IsNotfound() {
		return nil, fmt.Errorf("request scheduler api (get cluster info) not found")
	}

	if err = json.Unmarshal(buffer.Bytes(), &httpReqRes); err != nil {
		return nil, err
	}

	infoJson, err := json.Marshal(httpReqRes.Data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(infoJson, &infosData)
	if err != nil {
		return nil, err
	}

	return infosData, nil
}
