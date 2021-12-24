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
	"fmt"
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
	SystemExpression       string `file:"system_expression"`
	SystemTemplate         string `file:"system_template"`
	SystemExpressionConfig string `file:"system_expression_config"`
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
	Expressions = make([]*model.Expression, 0)
	orgExpressions, err := readExpressionFile(p.Cfg.SystemExpression)
	if err != nil {
		return err
	}
	Expressions = orgExpressions
	ExpressionConfig = make(map[string]*model.ExpressionConfig)
	expressionConfig, err := redaExpressionConfig(p.Cfg.SystemExpressionConfig)
	if err != nil {
		return err
	}
	ExpressionConfig = expressionConfig
	TemplateIndex = make(map[string][]*model.Template, 0)
	templates, err := readTemplateFile(p.Cfg.SystemTemplate)
	if err != nil {
		return err
	}
	TemplateIndex = templates
	if p.Register != nil {
		pb.RegisterExpressionServiceImp(p.Register, p.expressionService, apis.Options())
	}
	return nil
}

func redaExpressionConfig(root string) (map[string]*model.ExpressionConfig, error) {
	expressionConfig := make([]*model.ExpressionConfig, 0)
	expressionConfigMap := make(map[string]*model.ExpressionConfig)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		f, err := ioutil.ReadFile(path)
		fmt.Println(string(f))
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(f, &expressionConfig)
		if err != nil {
			return err
		}
		for _, v := range expressionConfig {
			expressionConfigMap[v.Id] = v
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return expressionConfigMap, nil
}

func readExpressionFile(root string) ([]*model.Expression, error) {
	expressions := make([]*model.Expression, 0)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		f, err := ioutil.ReadFile(path)
		var expressionModel model.Expression
		err = json.Unmarshal(f, &expressionModel)
		if err != nil {
			return err
		}
		expressions = append(expressions, &expressionModel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return expressions, nil
}

func readTemplateFile(root string) (map[string][]*model.Template, error) {
	f, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	templateMap := make(map[string][]*model.Template)
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
			alertIndex := strings.TrimSuffix(info.Name(), ".yml")
			templateMap[alertIndex] = templateModel
			Templates = append(Templates, templateModel...)
			for i := range templateModel {
				templateModel[i].AlertIndex = strings.TrimSuffix(info.Name(), ".yml")
				allRoute := strings.Split(path, "/")
				templateModel[i].AlertType = allRoute[len(allRoute)-2]
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return templateMap, nil
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
