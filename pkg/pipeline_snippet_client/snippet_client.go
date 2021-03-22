package pipeline_snippet_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httputil"
)

var snippetClientMap map[string]*apistructs.DicePipelineSnippetClient

var httpClient = httpclient.New(
	httpclient.WithTimeout(time.Second, time.Second*3),
)

func SetSnippetClientMap(clientMap map[string]*apistructs.DicePipelineSnippetClient) {
	snippetClientMap = clientMap
}

type querySnippetYamlResp struct {
	apistructs.Header
	Data string `json:"data"`
}

type batchQuerySnippetYamlResp struct {
	apistructs.Header
	Data []apistructs.BatchSnippetConfigYml `json:"data"`
}

func BatchGetSnippetPipelineYml(snippetConfig []apistructs.SnippetConfig) ([]apistructs.BatchSnippetConfigYml, error) {

	var configs = map[string][]apistructs.SnippetConfig{}
	for _, v := range snippetConfig {
		configs[v.Source] = append(configs[v.Source], v)
	}

	var results []apistructs.BatchSnippetConfigYml
	for key, v := range configs {
		clientConfig := snippetClientMap[key]
		if clientConfig == nil {
			return nil, fmt.Errorf("getSnippetPipelineYml error: can not find snippet host %v client", key)
		}

		var buffer bytes.Buffer
		r, err := httpClient.Post(clientConfig.Host).
			Path(clientConfig.Extra.UrlPathPrefix+"/actions/batch-query-snippet-yml").
			Header(httputil.InternalHeader, "pipeline_snippet_client").
			JSONBody(&v).
			Do().Body(&buffer)
		if err != nil {
			return nil, apierrors.ErrInvoke.InternalError(err)
		}

		if !r.IsOK() {
			return nil, apierrors.ErrInvoke.InternalError(fmt.Errorf("query snippet yml fail: please check snippet is exist, httpcode: %v, body: %s", r.StatusCode(), buffer.String()))
		}

		var resp batchQuerySnippetYamlResp
		err = json.NewDecoder(&buffer).Decode(&resp)
		if err != nil {
			return nil, apierrors.ErrInvoke.InternalError(fmt.Errorf("body: %s, decode error %v", buffer.String(), err))
		}

		if !resp.Success {
			return nil, apierrors.ErrInvoke.InternalError(fmt.Errorf("http client error: httpcode %v", r.StatusCode()))
		}
		if resp.Data != nil {
			results = append(results, resp.Data...)
		}
	}
	return results, nil
}

func GetSnippetPipelineYml(snippetConfig apistructs.SnippetConfig) (string, error) {
	clientConfig := snippetClientMap[snippetConfig.Source]
	if clientConfig == nil {
		return "", fmt.Errorf("getSnippetPipelineYml error: can not find snippet %s client", snippetConfig.Name)
	}

	var buffer bytes.Buffer
	r, err := httpClient.Post(clientConfig.Host).
		Path(clientConfig.Extra.UrlPathPrefix+"/actions/query-snippet-yml").
		Header(httputil.InternalHeader, "pipeline_snippet_client").
		JSONBody(&snippetConfig).
		Do().Body(&buffer)
	if err != nil {
		return "", apierrors.ErrInvoke.InternalError(err)
	}

	if !r.IsOK() {
		return "", apierrors.ErrInvoke.InternalError(fmt.Errorf("query snippet yml fail: please check snippet is exist, httpcode: %v, body: %s", r.StatusCode(), buffer.String()))
	}

	var resp querySnippetYamlResp
	err = json.NewDecoder(&buffer).Decode(&resp)
	if err != nil {
		return "", apierrors.ErrInvoke.InternalError(fmt.Errorf("body: %s, decode error %v", buffer.String(), err))
	}

	if resp.Success {
		return resp.Data, nil
	}

	return "", apierrors.ErrInvoke.InternalError(fmt.Errorf("http client error: httpcode %v", r.StatusCode()))
}
