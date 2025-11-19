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

package main

import (
	"embed"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/apps/ai-proxy/config"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/providers/ai-proxy" // import service hub dependencies
	"github.com/erda-project/erda/pkg/common"
)

//go:embed bootstrap.yml
var bootstrap string

//go:embed conf/routes
var routesFS embed.FS

//go:embed conf/templates
var templatesFS embed.FS

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
}

func main() {
	config.InjectEmbedFS(&routesFS, &templatesFS)
	common.Run(&servicehub.RunOptions{
		Content: bootstrap,
	})
}
