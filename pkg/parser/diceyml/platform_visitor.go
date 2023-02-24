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

package diceyml

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/strutil"
)

type PlatformVisitor struct {
	DefaultVisitor
	platformInfo map[string]string
	errs         []error
}

func NewPlatformVisitor(platformInfo map[string]string) DiceYmlVisitor {
	return &PlatformVisitor{
		platformInfo: platformInfo,
	}
}

func (p *PlatformVisitor) VisitObject(v DiceYmlVisitor, obj *Object) {
	objByte, err := yaml.Marshal(obj)
	if err != nil {
		p.errs = append(p.errs,
			errors.Wrap(err, "marshal object to yaml failed"))
		return
	}

	replaced := p.renderPlatformInfo(objByte)
	if err = yaml.Unmarshal(replaced, obj); err != nil {
		p.errs = append(p.errs, errors.Wrap(err, "unmarshal object failed"))
	}

	return
}

func (p *PlatformVisitor) renderPlatformInfo(input []byte) []byte {
	var (
		left, right, platformTag = "${", "}", "platform."
		rePlaceholder            = regexp.MustCompile("\\$\\{(.+?)\\}")
		fullPlaceholderMap       = make(map[string]string)
	)

	replaced := string(input)
	if !rePlaceholder.MatchString(replaced) {
		return input
	}

	for _, r := range rePlaceholder.FindAllString(replaced, -1) {
		placeholder, start, end, err := strutil.FirstCustomPlaceholder(r, left, right)
		if err != nil || start == end || !strings.HasPrefix(placeholder, platformTag) {
			formatErr := errors.Errorf("placeholder %s format error, %v", placeholder, err)
			p.errs = append(p.errs, formatErr)
			return input
		}
		fmt.Println(placeholder)

		if val, ok := p.platformInfo[strings.Trim(placeholder, platformTag)]; !ok {
			notSupportErr := errors.Errorf("placeholder %s doesn't support", placeholder)
			p.errs = append(p.errs, notSupportErr)
			return input
		} else if _, ok := fullPlaceholderMap[placeholder]; !ok {
			fullPlaceholderMap[r] = val
		}
	}

	for fullPlaceholder, val := range fullPlaceholderMap {
		replaced = strings.ReplaceAll(replaced, fullPlaceholder, val)
	}

	return []byte(replaced)
}

func (p *PlatformVisitor) CollectErrors() []error {
	return p.errs
}

func RenderPlatformInfo(obj *Object, platformInfo map[string]string) error {
	if obj == nil {
		return errors.New("dice obj is nil")
	}

	if platformInfo != nil {
		platformVisitor := NewPlatformVisitor(platformInfo)
		obj.Accept(platformVisitor)
		if len(platformVisitor.CollectErrors()) != 0 {
			return errors.Errorf("platform visitor error: %v", platformVisitor.CollectErrors())
		}
	}
	return nil
}
