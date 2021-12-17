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

package expression

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"

	logs "github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/expression/pb"
	alertdb "github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/core/monitor/expression/model"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	SystemOrgExpression   string `file:"system_org_expression"`
	SystemMicroExpression string `file:"system_micro_expression"`
	SystemTemplate        string `file:"system_template"`
}

type provider struct {
	Cfg                  *config
	Log                  logs.Logger
	Register             transport.Register `autowired:"service-register" optional:"true"`
	t                    i18n.Translator
	DB                   *gorm.DB `autowired:"mysql-client"`
	alertDB              *alertdb.AlertExpressionDB
	metricDB             *alertdb.MetricExpressionDB
	customizeAlertRuleDB *alertdb.CustomizeAlertRuleDB
	expressionService    *expressionService
}

func (p *provider) Init(ctx servicehub.Context) error {
	log := ctx.Logger()
	p.Log = log
	p.alertDB = &alertdb.AlertExpressionDB{
		DB: p.DB,
	}
	p.metricDB = &alertdb.MetricExpressionDB{
		DB: p.DB,
	}
	p.customizeAlertRuleDB = &alertdb.CustomizeAlertRuleDB{
		DB: p.DB,
	}
	p.expressionService = &expressionService{
		p: p,
	}
	p.t = ctx.Service("i18n").(i18n.I18n).Translator("alert")
	SystemExpressions = make(map[string][]*model.Expression)
	orgExpressions, err := readExpressionFile(p.Cfg.SystemOrgExpression)
	if err != nil {
		return err
	}
	SystemExpressions["org"] = orgExpressions
	microExpressions, err := readExpressionFile(p.Cfg.SystemMicroExpression)
	if err != nil {
		return err
	}
	SystemExpressions["micro_service"] = microExpressions
	if p.Register != nil {
		pb.RegisterExpressionServiceImp(p.Register, p.expressionService, apis.Options())
	}
	SystemTemplate = make([]*model.Template, 0)
	templates, err := readTemplateFile(p.Cfg.SystemTemplate)
	if err != nil {
		return err
	}
	SystemTemplate = templates
	return nil
}

func readExpressionFile(root string) ([]*model.Expression, error) {
	f, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	expressions := make([]*model.Expression, 0)
	for _, pkg := range f {
		err := filepath.Walk(filepath.Join(root, pkg.Name()), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				dir, err := os.ReadDir(path)
				if err != nil {
					return err
				}
				f = append(f, dir...)
				return nil
			}
			f, err := ioutil.ReadFile(path)
			var expressionModel model.Expression
			err = json.Unmarshal(f, &expressionModel)
			if err != nil {
				return err
			}
			expressionModel.AlertIndex = strings.TrimSuffix(info.Name(), ".json")
			allRoute := strings.Split(path, "/")
			expressionModel.AlertType = allRoute[len(allRoute)-2]
			expressions = append(expressions, &expressionModel)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return expressions, nil
}

func readTemplateFile(root string) ([]*model.Template, error) {
	f, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	templates := make([]*model.Template, 0)
	for _, pkg := range f {
		err := filepath.Walk(filepath.Join(root, pkg.Name()), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				dir, err := os.ReadDir(path)
				if err != nil {
					return err
				}
				f = append(f, dir...)
				return nil
			}
			f, err := ioutil.ReadFile(path)
			var templateModel []*model.Template
			err = yaml.Unmarshal(f, &templateModel)
			if err != nil {
				return err
			}
			for _, temp := range templateModel {
				var t *model.Template
				t = temp
				t.AlertIndex = strings.TrimSuffix(info.Name(), ".yml")
				allRoute := strings.Split(path, "/")
				t.AlertType = allRoute[len(allRoute)-2]
				templates = append(templates, t)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return templates, nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.monitor.expression.ExpressionService" || ctx.Type() == pb.ExpressionServiceServerType() || ctx.Type() == pb.ExpressionServiceHandlerType():
		return p.expressionService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.monitor.expression", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		OptionalDependencies: []string{"service-register"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Types: pb.Types(),
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
