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

package auto_test_plan_list

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-plan-list/i18n"
)

// GenCreateFormModalProps 生成创建测试计划表单的props
func GenCreateFormModalProps(ctxBundle protocol.ContextBundle, testSpace, iteration []byte) interface{} {
	i18nLocale := ctxBundle.Bdl.GetLocale(ctxBundle.Locale)
	props := fmt.Sprintf(`{
          "name": "%s",
          "fields": [
            {
              "component": "input",
              "key": "name",
              "label": "%s",
              "required": true,
              "rule": [
                {
                  "pattern": "/^[a-z\u4e00-\u9fa5A-Z0-9_-]*$/",
                  "msg": "%s"
                }
              ],
              "componentProps": {
                "maxLength": 50
              }
            },
            {
              "component": "select",
              "key": "spaceId",
              "label": "%s",
							"disabled": false,
              "required": true,
              "componentProps": {
                "options": `+string(testSpace)+
		`}
            },
			{
              "component": "select",
              "key": "iterationId",
              "label": "%s",
							"disabled": false,
              "required": true,
              "componentProps": {
                "options": `+string(iteration)+
		`}
            },
            {
              "key": "owners",
              "label": "%s",
              "required": true,
              "component": "memberSelector",
              "componentProps": {
                "mode": "multiple",
                "scopeType": "project"
              }
            }
          ]
        }`, i18nLocale.Get(i18n.I18nKeyPlan), i18nLocale.Get(i18n.I18nKeyPlanName),
		i18nLocale.Get(i18n.I18nKeyPlanNameRegex), i18nLocale.Get(i18n.I18nKeyTestSpace),
		i18nLocale.Get(i18n.I18nKeyIteration), i18nLocale.Get(i18n.I18nKeyPrincipal))

	var propsI interface{}
	if err := json.Unmarshal([]byte(props), &propsI); err != nil {
		logrus.Errorf("init props name=testplan component=formModal propsType=CreateTestPlan err: errMsg: %v", err)
	}

	return propsI
}

// GenUpdateFormModalProps 生成更新测试计划表单的props
func GenUpdateFormModalProps(ctxBundle protocol.ContextBundle, testSpace, iteration []byte) interface{} {
	i18nLocale := ctxBundle.Bdl.GetLocale(ctxBundle.Locale)
	props := fmt.Sprintf(`{
          "name": "%s",
          "fields": [
            {
              "component": "input",
              "key": "name",
              "label": "%s",
              "required": true,
              "rule": [
                {
                  "pattern": "/^[a-z\u4e00-\u9fa5A-Z0-9_-]*$/",
                  "msg": "%s"
                }
              ],
              "componentProps": {
                "maxLength": 50
              }
            },
            {
              "component": "select",
              "key": "spaceId",
              "label": "%s",
              "disabled": true,
							"componentProps": {
                "options": `+string(testSpace)+
		`}
            },
			{
              "component": "select",
              "key": "iterationId",
              "label": "%s",
              "required": true,
              "disabled": false,
							"componentProps": {
                "options": `+string(iteration)+
		`}
            },
            {
              "key": "owners",
              "label": "%s",
              "required": true,
              "component": "memberSelector",
              "componentProps": {
                "mode": "multiple",
                "scopeType": "project"
              }
            }
          ]
        }`, i18nLocale.Get(i18n.I18nKeyPlan), i18nLocale.Get(i18n.I18nKeyPlanName),
		i18nLocale.Get(i18n.I18nKeyPlanNameRegex), i18nLocale.Get(i18n.I18nKeyTestSpace),
		i18nLocale.Get(i18n.I18nKeyIteration), i18nLocale.Get(i18n.I18nKeyPrincipal))

	var propsI interface{}
	if err := json.Unmarshal([]byte(props), &propsI); err != nil {
		logrus.Errorf("init props name=testplan component=formModal propsType=UpdateTestPlan err: errMsg: %v", err)
	}

	return propsI
}
