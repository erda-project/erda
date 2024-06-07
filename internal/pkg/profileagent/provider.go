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

package profileagent

import (
	"os"

	"github.com/grafana/pyroscope-go"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
)

var (
	defaultProflieTypes = []pyroscope.ProfileType{
		pyroscope.ProfileCPU,
		pyroscope.ProfileAllocObjects,
		pyroscope.ProfileAllocSpace,
		pyroscope.ProfileInuseObjects,
		pyroscope.ProfileInuseSpace,
	}
)

type config struct {
	CollectorURL    string            `file:"collector_url" env:"DICE_COLLECTOR_URL" default:"http://collector:7076"`
	EnableOptions   bool              `file:"enable_options" env:"ENABLE_OPTIONS" default:"false"`
	CustomTags      map[string]string `file:"custom_tags" env:"CUSTOM_TAGS"`
	ApplicationName string            `file:"application_name" env:"APPLICATION_NAME"`
}

// provider .
type provider struct {
	Cfg *config
	Log logs.Logger
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.ApplicationName == "" {
		p.Cfg.ApplicationName = os.Getenv("DICE_SERVICE")
	}
	tags := map[string]string{
		"DICE_APPLICATION_ID":   os.Getenv("DICE_APPLICATION_ID"),
		"DICE_APPLICATION_NAME": os.Getenv("DICE_APPLICATION_NAME"),
		"DICE_CLUSTER_NAME":     os.Getenv("DICE_CLUSTER_NAME"),
		"DICE_ORG_ID":           os.Getenv("DICE_ORG_ID"),
		"DICE_ORG_NAME":         os.Getenv("DICE_ORG_NAME"),
		"DICE_PROJECT_ID":       os.Getenv("DICE_PROJECT_ID"),
		"DICE_PROJECT_NAME":     os.Getenv("DICE_PROJECT_NAME"),
		"DICE_SERVICE":          os.Getenv("DICE_SERVICE"),
		"DICE_WORKSPACE":        os.Getenv("DICE_WORKSPACE"),
		"POD_IP":                os.Getenv("POD_IP"),
	}
	for k, v := range p.Cfg.CustomTags {
		tags[k] = v
	}
	var profileTypes = append([]pyroscope.ProfileType{}, defaultProflieTypes...)
	if p.Cfg.EnableOptions {
		profileTypes = append(defaultProflieTypes, pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration)
	}
	pyroscope.Start(pyroscope.Config{
		ApplicationName: p.Cfg.ApplicationName,
		ServerAddress:   p.Cfg.CollectorURL,
		Logger:          p.Log,
		Tags:            tags,
		ProfileTypes:    profileTypes,
	})
	return nil
}

func init() {
	servicehub.Register("profile-agent", &servicehub.Spec{
		Services:     []string{"profile-agent"},
		Dependencies: []string{},
		Description:  "start profile agent",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
