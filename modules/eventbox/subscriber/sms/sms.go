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

package sms

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/eventbox/subscriber"
	"github.com/erda-project/erda/modules/eventbox/types"
)

type MobileSubscriber struct {
	accessKeyId         string
	accessSecret        string
	signName            string
	monitorTemplateCode string
	bundle              *bundle.Bundle
}

type MobileData struct {
	Template string            `json:"template"`
	Params   map[string]string `json:"params"`
	OrgID    int64             `json:"orgID"`
}

type Option func(*MobileSubscriber)

func New(accessKeyId, accessKeySecret, signName, monitorTemplateCode string, bundle *bundle.Bundle) subscriber.Subscriber {
	subscriber := &MobileSubscriber{
		accessKeyId:  accessKeyId,
		accessSecret: accessKeySecret,
		signName:     signName,
		bundle:       bundle,
	}
	return subscriber
}

func (d *MobileSubscriber) Publish(dest string, content string, time int64, msg *types.Message) []error {
	errs := []error{}
	var mobiles []string
	err := json.Unmarshal([]byte(dest), &mobiles)
	if err != nil {
		return []error{err}
	}
	var mobileData MobileData
	err = json.Unmarshal([]byte(content), &mobileData)
	if err != nil {
		return []error{err}
	}
	paramStr, err := json.Marshal(mobileData.Params)
	if err != nil {
		return []error{err}
	}

	org, err := d.bundle.GetOrg(mobileData.OrgID)
	if err != nil {
		logrus.Errorf("failed to get org info err:%s", err)
	}

	accessKeyID, accessSecret, signName := d.accessKeyId, d.accessSecret, d.signName
	if err == nil && org.Config != nil && org.Config.SMSKeyID != "" && org.Config.SMSKeySecret != "" {
		accessKeyID = org.Config.SMSKeyID
		accessSecret = org.Config.SMSKeySecret
		signName = org.Config.SMSSignName
	}

	sdkClient, err := sdk.NewClientWithAccessKey("cn-hangzhou", accessKeyID, accessSecret)
	if err != nil {
		return []error{err}
	}

	// 通知组的短信模版存在notifyitem里
	templateCode := mobileData.Template
	if msg.Sender == "analyzer-alert" {
		// 因为目前监控只用一个模板，并且不通过group的发送，所以监控的短信模版存在orgConfig里或者环境变量里
		templateCode = d.monitorTemplateCode
		if err == nil && org.Config != nil && org.Config.SMSMonitorTemplateCode != "" {
			templateCode = org.Config.SMSMonitorTemplateCode
		}
	}

	if templateCode == "" {
		return []error{fmt.Errorf("empty template_code")}
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-hangzhou"
	request.QueryParams["PhoneNumbers"] = strings.Join(mobiles, ",")
	request.QueryParams["SignName"] = signName
	request.QueryParams["TemplateCode"] = templateCode
	request.QueryParams["TemplateParam"] = string(paramStr)

	response, err := sdkClient.ProcessCommonRequest(request)
	if err != nil {
		logrus.Errorf("failed to send sms  %s", err)
		return []error{err}
	}
	if !response.IsSuccess() {
		logrus.Errorf("failed to send sms  %s", response.GetHttpContentString())
		return []error{fmt.Errorf("failed to send sms %s", response.GetHttpContentString())}
	}
	logrus.Infof("sms send success %s", response.GetHttpContentString())
	return errs
}

func (d *MobileSubscriber) Status() interface{} {
	return nil
}

func (d *MobileSubscriber) Name() string {
	return "SMS"
}
