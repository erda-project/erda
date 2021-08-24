// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package pipeline

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
	"github.com/erda-project/erda/pkg/dumpstack"
)

type provider struct {
	CmsService    pb.CmsServiceServer           `autowired:"erda.core.pipeline.cms.CmsService"`
	PluginsManage *plugins_manage.PluginsManage `autowired:"erda.core.pipeline.aop.plugins"`
}

func (p *provider) Run(ctx context.Context) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000000",
	})
	logrus.SetOutput(os.Stdout)

	dumpstack.Open()
	logrus.Infoln(version.String())

	return p.Initialize()
}

func init() {
	servicehub.Register("pipeline", &servicehub.Spec{
		Services: []string{"pipeline"},
		Creator:  func() servicehub.Provider { return &provider{} },
	})
}
