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

package aws_bedrock

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/smithy-go/logging"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/common/message_converter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/common/openai_extended"
	set_resp_body_chunk_splitter "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/set-resp-body-chunk-splitter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

const (
	APIVendor api_style.APIVendor = "aws-bedrock"

	defaultBedrockVersion = "bedrock-2023-05-31"
)

type BedrockRequest struct {
	message_converter.BaseAnthropicRequest
	AnthropicVersion string `json:"anthropic_version"`
}

type BedrockDirector struct {
	StreamMessageInfo message_converter.AnthropicStreamMessageInfo
}

func NewDirector() *BedrockDirector {
	return &BedrockDirector{}
}

func (f *BedrockDirector) AwsBedrockDirector(pr *httputil.ProxyRequest, apiStyleConfig api_style.APIStyleConfig) error {
	// handle path for stream
	if ctxhelper.MustGetIsStream(pr.Out.Context()) {
		pr.Out.URL.Path = strings.ReplaceAll(pr.Out.URL.Path, "/invoke", "/invoke-with-response-stream")
		// use bedrock stream splitter
		ctxhelper.PutRespBodyChunkSplitter(pr.Out.Context(), &set_resp_body_chunk_splitter.BedrockStreamSplitter{})
	}
	// openai format request
	var openaiReq openai_extended.OpenAIRequestExtended
	if err := json.NewDecoder(pr.Out.Body).Decode(&openaiReq); err != nil {
		panic(fmt.Errorf("failed to decode request body as openai format, err: %v", err))
	}
	// set options for bedrock
	openaiReq.Options = openai_extended.Options{
		ImageURLForceBase64: &[]bool{true}[0], // force convert http image url to base64
	}
	// convert to: anthropic format request
	baseAnthropicReq := message_converter.ConvertOpenAIRequestToBaseAnthropicRequest(openaiReq)
	bedrockReq := BedrockRequest{
		BaseAnthropicRequest: baseAnthropicReq,
		AnthropicVersion:     defaultBedrockVersion,
	}

	anthropicReqBytes, err := json.Marshal(&bedrockReq)
	if err != nil {
		panic(fmt.Errorf("failed to marshal anthropic request: %v", err))
	}
	if err := body_util.SetBody(pr.Out, anthropicReqBytes); err != nil {
		panic(fmt.Errorf("failed to set request body: %v", err))
	}

	// sign aws request with the converted body
	if err := SignAwsRequest(pr, anthropicReqBytes); err != nil {
		panic(fmt.Errorf("failed to sign aws request: %v", err))
	}
	return nil
}

// SignAwsRequest signs the request with AWS SigV4, clearing existing signatures first
func SignAwsRequest(pr *httputil.ProxyRequest, bodyBytes []byte) error {
	// get ak/sk
	provider := ctxhelper.MustGetModelProvider(pr.Out.Context())
	ak := provider.Metadata.Secret["ak"].GetStringValue()
	sk := provider.Metadata.Secret["sk"].GetStringValue()
	if ak == "" || sk == "" {
		return fmt.Errorf("missing provider.metadata.secret.{ak,sk}")
	}
	credCaches := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(ak, sk, ""))
	cred, err := credCaches.Retrieve(pr.Out.Context())
	if err != nil {
		return fmt.Errorf("failed to retrieve aws credentials: %v", err)
	}
	location := provider.Metadata.Public["location"].GetStringValue()
	if location == "" {
		return fmt.Errorf("missing provider.metadata.public.location")
	}

	// payload hash
	var payloadHash string
	sum := sha256.Sum256(bodyBytes)
	payloadHash = hex.EncodeToString(sum[:])
	pr.Out.Header.Set("X-Amz-Content-Sha256", payloadHash)

	// remove headers not required for AWS SigV4 signing
	// Keep only known safe headers
	keepHeaders := map[string]bool{
		"host":                 true,
		"content-type":         true,
		"content-length":       true,
		"accept":               true,
		"x-amz-date":           true,
		"x-amz-content-sha256": true,
	}
	for k := range pr.Out.Header {
		if !keepHeaders[strings.ToLower(k)] {
			pr.Out.Header.Del(k)
		}
	}

	// do aws sign v4
	signer := v4.NewSigner()
	if err := signer.SignHTTP(pr.Out.Context(), cred, pr.Out, payloadHash, "bedrock", location, time.Now(),
		func(options *v4.SignerOptions) {
			options.LogSigning = true
			options.Logger = logging.NewStandardLogger(os.Stdout)
		},
	); err != nil {
		return fmt.Errorf("failed to sign request: %v", err)
	}
	return nil
}

func (f *BedrockDirector) OnBodyChunk(resp *http.Response, chunk []byte, index int64) ([]byte, error) {
	// non-stream
	if !ctxhelper.MustGetIsStream(resp.Request.Context()) {
		// convert all at once
		var bedrockResp message_converter.AnthropicResponse
		if err := json.Unmarshal(chunk, &bedrockResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %s, err: %v", string(chunk), err)
		}
		openaiResp, err := bedrockResp.ConvertToOpenAIFormat(ctxhelper.MustGetModel(resp.Request.Context()).Metadata.Public["model_id"].GetStringValue())
		if err != nil {
			return nil, fmt.Errorf("failed to convert bedrock anthropic response body to openai format, err: %v", err)
		}
		return json.Marshal(openaiResp)
	}
	// stream
	var chunkWriter bytes.Buffer
	openaiChunks, err := f.pipeBedrockStream(resp.Request.Context(), io.NopCloser(bytes.NewBuffer(chunk)), &chunkWriter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bedrock eventstream, err: %v", err)
	}
	var chunkDataList [][]byte
	for _, openaiChunk := range openaiChunks {
		b, err := json.Marshal(openaiChunk)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal openai chunk, err: %v", err)
		}
		chunkData := vars.ConcatChunkDataPrefix(b)
		chunkDataList = append(chunkDataList, chunkData)
	}
	result := bytes.Join(chunkDataList, []byte("\n\n"))
	return result, nil
}

func (f *BedrockDirector) OnComplete(resp *http.Response) ([]byte, error) {
	if ctxhelper.MustGetIsStream(resp.Request.Context()) {
		// append [DONE] chunk
		doneChunk := vars.ConcatChunkDataPrefix([]byte("[DONE]"))
		return doneChunk, nil
	}
	return nil, nil
}
