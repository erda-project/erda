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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	netutil "net/http/httputil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/smithy-go/logging"
	"github.com/stretchr/testify/assert"
)

func TestSignV4(t *testing.T) {
	//payload := `{"messages":[{"role":"user","content":"hi"}],"max_tokens":1024,"anthropic_version":"bedrock-2023-05-31"}`
	payload := `{"messages":[{"content":"test","role":"user"}],"temperature":1,"max_tokens":1024,"anthropic_version":"bedrock-2023-05-31"}`
	sum := sha256.Sum256([]byte(payload))
	h := hex.EncodeToString(sum[:])

	//modelID := "anthropic.claude-3-7-sonnet-20250219-v1:0"
	modelID := "anthropic.claude-3-sonnet-20240229-v1:0"
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("https://bedrock-runtime.us-east-1.amazonaws.com/model/%s/invoke-with-response-stream", modelID),
		bytes.NewBufferString(payload))

	req.Host = "bedrock-runtime.us-east-1.amazonaws.com"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(payload)))
	req.Header.Set("X-Amz-Content-Sha256", h)

	signer := v4.NewSigner()
	ctx := context.TODO()
	credCaches := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(os.Getenv("AWS_BEDROCK_AK"), os.Getenv("AWS_BEDROCK_SK"), ""),
	)
	cred, err := credCaches.Retrieve(ctx)
	assert.NoError(t, err)
	_ = signer.SignHTTP(ctx, cred, req, h, "bedrock", "us-east-1", time.Now(), func(options *v4.SignerOptions) {
		options.LogSigning = true
		options.Logger = logging.NewStandardLogger(os.Stdout)
	})

	// do http request
	req.Header.Set("Accept", "application/json")
	client := http.DefaultClient
	client.Transport = &DumpTransport{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	fmt.Println("Response Status:", resp.Status)
	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Printf("Response body: %s\n", string(respBody))
}

type DumpTransport struct{}

func (t *DumpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	raw, _ := netutil.DumpRequestOut(req, true)
	fmt.Printf("url.path: %s\n", req.URL.Path)
	fmt.Printf("url.rawpath: %s\n", req.URL.RawPath)
	fmt.Printf("\n--- OUTBOUND REQUEST (%d bytes) ---\n%s", len(raw), raw)
	// set back
	if req.Body != nil {
		req.Body = io.NopCloser(bytes.NewReader(raw[len(raw)-int(req.ContentLength):]))
	}
	d := http.DefaultTransport
	d.(*http.Transport).ForceAttemptHTTP2 = true
	return d.RoundTrip(req)
}
