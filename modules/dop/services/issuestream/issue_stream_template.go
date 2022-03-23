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

package issuestream

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
)

// getIssueStreamTemplate get issue stream template
func getIssueStreamTemplate(locale string, ist apistructs.IssueStreamType) (string, error) {
	if locale != "zh" && locale != "en" {
		return "", errors.Errorf("invalid locale %v", locale)
	}

	v, ok := apistructs.IssueTemplate[locale][ist]
	if !ok {
		return "", errors.Errorf("issue stream template not found")
	}

	return v, nil
}

func getIssueStreamReasonTemplate(locale, key string) (string, error) {
	if locale != "zh" && locale != "en" {
		return "", errors.Errorf("invalid locale %v", locale)
	}

	v, ok := apistructs.ReasonTemplate[locale][key]
	if !ok {
		return "", errors.Errorf("issue stream reason template not found")
	}

	return v, nil
}

// getIssueStreamTemplateForMsgSending get issue stream template for msg sending
func getIssueStreamTemplateForMsgSending(locale string, ist apistructs.IssueStreamType) (string, error) {
	if locale == "" || locale == "zh-CN" {
		locale = "zh"
	}
	if locale == "en-US" {
		locale = "en"
	}
	templateContent, err := getIssueStreamTemplate(locale, ist)
	if err != nil {
		return "", err
	}

	// override template if have for msg sending
	vv, ok := apistructs.IssueTemplateOverrideForMsgSending[locale][ist]
	if ok {
		templateContent = vv
	}

	return templateContent, nil
}

// getDefaultContent get rendered msg
func getDefaultContent(ist apistructs.IssueStreamType, param apistructs.ISTParam, locale string) (string, error) {
	if strings.Contains(locale, "zh") {
		locale = "zh"
	}
	if strings.Contains(locale, "en") {
		locale = "en"
	}
	ct, err := getIssueStreamTemplate(locale, ist)
	if err != nil {
		return "", err
	}
	content, err := renderTemplate(locale, ct, param)
	if err != nil {
		return "", err
	}
	if param.ReasonDetail != "" {
		reason, err := getIssueStreamReasonTemplate(locale, param.ReasonDetail)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v %v", content, reason), nil
	}
	return content, nil
}

// getDefaultContentForMsgSending get rendered msg for sending
func getDefaultContentForMsgSending(ist apistructs.IssueStreamType, param apistructs.ISTParam, locale string) (string, error) {
	ct, err := getIssueStreamTemplateForMsgSending(locale, ist)
	if err != nil {
		return "", err
	}

	return renderTemplate(locale, ct, param)
}

// renderTemplate render template
func renderTemplate(locale, templateContent string, param apistructs.ISTParam) (string, error) {
	tpl, err := template.New("c").Parse(templateContent)
	if err != nil {
		return "", err
	}

	var content bytes.Buffer
	if err := tpl.Execute(&content, param.Localize(locale)); err != nil {
		return "", err
	}

	return content.String(), nil
}
