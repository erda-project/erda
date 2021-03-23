package bundle

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

func (b *Bundle) GetExtensionVersion(req apistructs.ExtensionVersionGetRequest) (*apistructs.ExtensionVersion, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.ExtensionVersionGetResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/extensions/%v/%v", req.Name, req.Version)).
		Param("yamlFormat", strconv.FormatBool(req.YamlFormat)).
		Header("Internal-Client", "bundle").
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return &getResp.Data, nil
}

func (b *Bundle) QueryExtensionVersions(req apistructs.ExtensionVersionQueryRequest) ([]apistructs.ExtensionVersion, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.ExtensionVersionQueryResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/extensions/%v", req.Name)).
		Param("all", req.All).
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

func (b *Bundle) QueryExtensions(req apistructs.ExtensionQueryRequest) ([]apistructs.Extension, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.ExtensionQueryResponse
	resp, err := hc.Get(host).Path("/api/extensions").
		Param("all", req.All).
		Param("type", req.Type).
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

func (b *Bundle) SearchExtensions(req apistructs.ExtensionSearchRequest) (map[string]apistructs.ExtensionVersion, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.ExtensionSearchResponse
	resp, err := hc.Post(host).Path("/api/extensions/actions/search").
		Header("Internal-Client", "bundle").
		JSONBody(req).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return getResp.Data, nil
}

func (b *Bundle) GetPublishItem(publishItemID int64) (*apistructs.PublishItem, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.PublishItemResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/publish-items/%d", publishItemID)).
		Header("Internal-Client", "bundle").
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return &(getResp.Data), nil
}

func (b *Bundle) RenderPipelineTemplate(request *apistructs.PipelineTemplateRenderRequest) (*apistructs.PipelineTemplateRender, error) {

	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.PipelineTemplateRenderResponse
	resp, err := hc.Post(host).Path(fmt.Sprintf("/api/pipeline-templates/%s/actions/render", request.Name)).
		Header("Internal-Client", "bundle").
		JSONBody(request).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}

	return &getResp.Data, nil
}

func (b *Bundle) RenderPipelineTemplateBySpec(request *apistructs.PipelineTemplateRenderSpecRequest) (*apistructs.PipelineTemplateRender, error) {

	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.PipelineTemplateRenderResponse
	resp, err := hc.Post(host).Path(fmt.Sprintf("/api/pipeline-templates/local/actions/render-spec")).
		Header("Internal-Client", "bundle").
		JSONBody(request).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}

	return &getResp.Data, nil
}

func (b *Bundle) GetPipelineTemplateVersion(request *apistructs.PipelineTemplateVersionGetRequest) (*apistructs.PipelineTemplateVersion, error) {

	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.PipelineTemplateVersionGetResponse
	resp, err := hc.Post(host).Path(fmt.Sprintf("/api/pipeline-templates/%s/actions/query-version", request.Name)).
		Header("Internal-Client", "bundle").
		JSONBody(request).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}

	return &getResp.Data, nil
}
