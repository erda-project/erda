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

package vms

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dyvmsapi"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/eventbox/subscriber"
	"github.com/erda-project/erda/modules/eventbox/types"
)

// VoiceSubscriber 语音通知分发
type VoiceSubscriber struct {
	accessKeyID             string
	accessSecret            string
	monitorTtsCode          string
	monitorCalledShowNumber string
	bundle                  *bundle.Bundle
}

// VoiceData 语音通知数据
type VoiceData struct {
	Template         string            `json:"template"`
	CalledShowNumber string            `json:"calledShowNumber"`
	Params           map[string]string `json:"params"`
	OrgID            int64             `json:"orgID"`
}

type Option func(*VoiceSubscriber)

// New 新建一个语音通知分发的实例
func New(accessKeyID, accessKeySecret, monitorTtsCode, monitorCalledShowNumber string, bundle *bundle.Bundle) subscriber.Subscriber {
	subscriber := &VoiceSubscriber{
		accessKeyID:             accessKeyID,
		accessSecret:            accessKeySecret,
		monitorTtsCode:          monitorTtsCode,
		monitorCalledShowNumber: monitorCalledShowNumber,
		bundle:                  bundle,
	}
	return subscriber
}

// Publish 推送语音
func (d *VoiceSubscriber) Publish(dest string, content string, time int64, msg *types.Message) []error {
	errs := []error{}
	var mobiles []string
	err := json.Unmarshal([]byte(dest), &mobiles)
	if err != nil {
		return []error{err}
	}
	var voiceData VoiceData
	err = json.Unmarshal([]byte(content), &voiceData)
	if err != nil {
		return []error{err}
	}
	paramStr, err := json.Marshal(voiceData.Params)
	if err != nil {
		return []error{err}
	}

	org, err := d.bundle.GetOrg(voiceData.OrgID)
	if err != nil {
		logrus.Errorf("failed to get org info err:%s", err)
	}

	accessKeyID, accessSecret := d.accessKeyID, d.accessSecret
	if err == nil && org.Config != nil && org.Config.VMSKeyID != "" && org.Config.VMSKeySecret != "" {
		accessKeyID = org.Config.VMSKeyID
		accessSecret = org.Config.VMSKeySecret
	}

	// 语音模版属于公共号码池外呼的时候，被叫显号必须是空
	// 属于专属号码外呼的时候，被叫显号不能为空
	// 通知组的语音模版和被叫显号存在notifyitem里
	ttsCode, calledShowNumber := voiceData.Template, voiceData.CalledShowNumber
	if msg.Sender == "analyzer-alert" {
		// 因为目前监控只用一个模板，所以监控的语音模版和被呼号码存在orgConfig里或者环境变量里
		ttsCode, calledShowNumber = d.monitorTtsCode, d.monitorCalledShowNumber
		if err == nil && org.Config != nil && org.Config.VMSMonitorTtsCode != "" {
			ttsCode, calledShowNumber = org.Config.VMSMonitorTtsCode, org.Config.VMSMonitorCalledShowNumber
		}
	}

	if ttsCode == "" {
		return []error{fmt.Errorf("empty tts_code")}
	}

	client, err := dyvmsapi.NewClientWithAccessKey("cn-hangzhou", accessKeyID, accessSecret)
	if err != nil {
		return []error{err}
	}

	wg := sync.WaitGroup{}
	for i := range mobiles {
		wg.Add(1)
		go func(mobile string) {
			defer wg.Done()
			request := dyvmsapi.CreateSingleCallByTtsRequest()
			request.Scheme = "https"
			request.TtsParam = string(paramStr)
			request.CalledNumber = mobile
			request.TtsCode = ttsCode
			request.CalledShowNumber = calledShowNumber

			response, err := client.SingleCallByTts(request)
			if err != nil {
				logrus.Errorf("failed to send voice to %s: %s", mobile, err)
				errs = append(errs, fmt.Errorf("failed to send voice to %s: %s", mobile, err))
			}
			if !response.IsSuccess() {
				logrus.Errorf("failed to send voice to %s: %s", mobile, response.GetHttpContentString())
				errs = append(errs, fmt.Errorf("failed to send voice to %s: %s", mobile, err))
			}
		}(mobiles[i])
	}
	wg.Wait()

	if len(errs) == 0 {
		logrus.Info("voice send success")
	}

	return errs
}

func (d *VoiceSubscriber) Status() interface{} {
	return nil
}

func (d *VoiceSubscriber) Name() string {
	return "VMS"
}
