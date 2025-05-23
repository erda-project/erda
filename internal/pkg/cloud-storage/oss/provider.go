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

package oss

import (
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type config struct {
	Bucket string `file:"bucket"`
	Region string `file:"region"`
	Prefix string `file:"prefix"`
	MaxKey int32  `file:"maxKey"`
}

type provider struct {
	Cfg    *config
	client *oss.Client
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(p.Cfg.Region)
	p.client = oss.NewClient(cfg)
	logrus.Infof("init oss provider success")
	return nil
}
func init() {
	servicehub.Register("oss-provider", &servicehub.Spec{
		Services:     []string{"oss-provider"},
		Dependencies: []string{},
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
