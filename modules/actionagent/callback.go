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

package actionagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/retry"
	"github.com/erda-project/erda/pkg/strutil"
)

func (agent *Agent) Callback() {
	cb := &Callback{}
	defer func() {
		cb.Errors = append(cb.Errors, agent.MergeErrors()...)
		agent.LockPushedMetaFileMap.Lock()
		defer agent.LockPushedMetaFileMap.Unlock()
		if err := agent.callbackToPipelinePlatform(cb); err != nil {
			for _, err := range cb.Errors {
				logrus.Println(err.Msg)
			}
			// 回调失败，直接打印错误日志
			logrus.Println(err)
		}
	}()

	// ${METAFILE}
	_, err := os.Stat(agent.EasyUse.ContainerMetaFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		agent.AppendError(err)
		return
	}
	r, err := os.Open(agent.EasyUse.ContainerMetaFile)
	if err != nil {
		agent.AppendError(err)
		return
	}
	mb, err := ioutil.ReadAll(r)
	if err != nil {
		if err == io.EOF {
			return
		}
		agent.AppendError(err)
		return
	}
	if err := cb.HandleMetaFile(mb); err != nil {
		agent.AppendError(err)
		return
	}
}

func (agent *Agent) callbackToPipelinePlatform(cb *Callback) (err error) {

	filterMetadata(cb, agent)
	defer func() {
		if err == nil {
			updatePushedMetadata(cb, agent)
		} else {
			err = fmt.Errorf("callback to pipeline platform failed, err: %v", err)
		}
	}()

	if agent.EasyUse.OpenAPIAddr == "" {
		return errors.New("unknown openapi addr, cannot callback")
	}

	// 如果全部为空，则不需要回调
	if len(cb.Metadata) == 0 && len(cb.Errors) == 0 && cb.MachineStat == nil {
		return nil
	}

	cbByte, _ := json.Marshal(cb)
	logrus.Debugf("begin callback meta: %s", string(cbByte))

	type config struct {
		OpenAPIToken   string `env:"DICE_OPENAPI_TOKEN" required:"true"`
		PipelineID     uint64 `env:"PIPELINE_ID"`
		PipelineTaskID uint64 `env:"PIPELINE_TASK_ID"`
		InternalClient string `env:"DICE_INTERNAL_CLIENT"`
		UserID         string `env:"DICE_USER_ID"`
	}

	var cfg config
	if err := envconf.Load(&cfg); err != nil {
		return err
	}

	// 兜底方案从 env 中获取回调函数的必要参数
	if agent.Arg.PipelineID == 0 {
		agent.Arg.PipelineID = cfg.PipelineID
	}
	if agent.Arg.PipelineTaskID == 0 {
		agent.Arg.PipelineTaskID = cfg.PipelineTaskID
	}
	cb.PipelineID = agent.Arg.PipelineID
	cb.PipelineTaskID = agent.Arg.PipelineTaskID

	// 序列化新的 []byte
	b, err := json.Marshal(cb)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}

	var cbReq apistructs.PipelineCallbackRequest
	cbReq.Type = string(apistructs.PipelineCallbackTypeOfAction)
	cbReq.Data = b

	return retry.DoWithInterval(func() error {
		var resp apistructs.PipelineCallbackResponse

		r, err := httpclient.New(httpclient.WithCompleteRedirect()).
			Post(agent.EasyUse.OpenAPIAddr).
			Path("/api/pipelines/actions/callback").
			Header("Authorization", cfg.OpenAPIToken).
			JSONBody(&cbReq).
			Do().
			JSON(&resp)
		if err != nil {
			return err
		}
		if !r.IsOK() || !resp.Success {
			return errors.Errorf("status-code %d, resp %#v", r.StatusCode(), resp)
		}
		return nil
	}, 5, time.Second*5)
}

type Callback apistructs.ActionCallback

// append fields to metadata and limit metadataField
func (c *Callback) AppendMetadataFields(fields []*apistructs.MetadataField) {

	if fields == nil {
		return
	}

	for _, field := range fields {

		if field == nil {
			continue
		}

		var name = field.Name
		var value = field.Value
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)

		c.Metadata = append(c.Metadata, apistructs.MetadataField{Name: name, Value: value})
	}

	c.limitMetadataField()
}

// 1) decode as: apistructs.Metadata
// 2) decode as: line(k=v)
func (c *Callback) HandleMetaFile(b []byte) error {
	// 1)
	err := json.NewDecoder(bytes.NewReader(b)).Decode(c)
	if err == nil || err == io.EOF {
		return nil
	}

	// 2)
	lines := strutil.Lines(string(b), true)
	for _, line := range lines {
		kv := strings.SplitN(line, "=", 2)
		var k string
		var v string
		if len(kv) > 0 {
			k = strings.TrimSpace(kv[0])
		}
		if len(kv) > 1 {
			v = strings.TrimSpace(kv[1])
		}
		c.Metadata = append(c.Metadata, apistructs.MetadataField{Name: k, Value: v})
	}

	c.limitMetadataField()

	return nil
}

// limit metadata
// key length <= 128
// value length <= 1024000
// metadata length <= 100
func (c *Callback) limitMetadataField() {
	var result apistructs.Metadata
	for i, meta := range c.Metadata {
		if i >= 100 {
			logrus.Warnf("skip meta (too many metadata, max size 100), index: %d, name: %s", i+1, meta.Name)
			continue
		}
		if len(meta.Name) > 128 {
			logrus.Warnf("skip meta (meta name is too long, max length 128), name: %s", meta.Name)
			continue
		}
		if len(meta.Value) > 1024000 {
			logrus.Warnf("skip meta (meta value is too long, max length 1024000), name: %s", meta.Name)
			continue
		}
		result = append(result, apistructs.MetadataField{Name: meta.Name, Value: meta.Value})
	}
	c.Metadata = result
}

func filterMetadata(cb *Callback, agent *Agent) {
	// 不推送已经推送过的
	var filteredMetadata apistructs.Metadata
	for _, wait := range cb.Metadata {
		pushedV, ok := agent.PushedMetaFileMap[wait.Name]
		if ok && pushedV == wait.Value {
			continue
		}
		filteredMetadata = append(filteredMetadata, wait)
	}
	cb.Metadata = filteredMetadata
}

func updatePushedMetadata(cb *Callback, agent *Agent) {
	for _, meta := range cb.Metadata {
		agent.PushedMetaFileMap[meta.Name] = meta.Value
	}
}
