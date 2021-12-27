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
	AlertRules  string `file:"alert_rules"`
	MetricRules string `file:"metric_rules"`
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
	alertNotifyDB        *alertdb.AlertNotifyDB
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
	p.alertNotifyDB = &alertdb.AlertNotifyDB{
		DB: p.DB,
	}
	p.expressionService = &expressionService{
		p: p,
	}
	p.t = ctx.Service("i18n").(i18n.I18n).Translator("alert")
	err := readAlertRule(p.Cfg.AlertRules)
	if err != nil {
		return err
	}
	err = readMetricRule(p.Cfg.MetricRules)
	if err != nil {
		return err
	}
	getAlertType()
	if p.Register != nil {
		pb.RegisterExpressionServiceImp(p.Register, p.expressionService, apis.Options())
	}
	return nil
}

func getAlertType() {
	OrgAlertType = make([]string, 0)
	MicroServiceAlertType = make([]string, 0)
	OrgAlertTypeMap := make(map[string]bool)
	MicroServiceAlertTypeMap := make(map[string]bool)
	for _, v := range ExpressionConfig {
		if v.AlertScope == Org {
			OrgAlertTypeMap[v.AlertType] = true
		} else {
			MicroServiceAlertTypeMap[v.AlertType] = true
		}
	}
	for k := range OrgAlertTypeMap {
		OrgAlertType = append(OrgAlertType, k)
	}
	for k := range MicroServiceAlertTypeMap {
		MicroServiceAlertType = append(MicroServiceAlertType, k)
	}
}

func readMetricRule(root string) error {
	MetricExpression = make([]*pb.Expression, 0)
	f, err := os.ReadDir(root)
	if err != nil {
		return err
	}
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
			expression := &pb.Expression{}
			err = json.Unmarshal(f, expression)
			if err != nil {
				return err
			}
			MetricExpression = append(MetricExpression, expression)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func readAlertRule(root string) error {
	Expressions = make([]*model.Expression, 0)
	TemplateIndex = make(map[string][]*model.Template)
	ExpressionIndex = make(map[string]*model.Expression)
	ExpressionConfig = make(map[string]*model.ExpressionConfig)
	Templates = make([]*model.Template, 0)
	f, err := os.ReadDir(root)
	if err != nil {
		return err
	}
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
			allRoute := strings.Split(path, "/")
			alertIndex := allRoute[len(allRoute)-2]
			alertType := allRoute[len(allRoute)-3]
			if info.Name() == model.NOTIFY_TEMPLATE {
				templates := make([]*model.Template, 0)
				err = yaml.Unmarshal(f, &templates)
				if err != nil {
					return err
				}
				for _, v := range templates {
					v.AlertIndex = alertIndex
					v.AlertType = alertType
					//(*v).AlertIndex = alertIndex
					//(*v).AlertType = alertType
				}
				TemplateIndex[alertIndex] = templates
				Templates = append(Templates, templates...)
			}
			if info.Name() == model.ALERT_RULE {
				expressionConfig := &model.ExpressionConfig{}
				err = yaml.Unmarshal(f, expressionConfig)
				if err != nil {
					return err
				}
				expressionConfig.AlertType = alertType
				ExpressionConfig[alertIndex] = expressionConfig
			}
			if info.Name() == model.ANALYZER_EXPRESSION {
				expression := &model.Expression{}
				err = json.Unmarshal(f, expression)
				if err != nil {
					return err
				}
				Expressions = append(Expressions, expression)
				ExpressionIndex[alertIndex] = expression
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
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
