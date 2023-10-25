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

package utils

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

const appNameLengthLimit = 36

func CheckRuntime(r *apistructs.ServiceGroup) error {
	group := CombineEDASAppGroup(r.Type, r.ID)
	length := appNameLengthLimit - len(group)

	var regexString = "^[A-Za-z_][A-Za-z0-9_]*$"

	for _, s := range r.Services {
		if len(s.Name) > length {
			return errors.Errorf("edas app name is longer than %d characters, name: %s",
				appNameLengthLimit, group+s.Name)
		}

		for k := range s.Env {
			match, err := regexp.MatchString(regexString, k)
			if err != nil {
				errMsg := fmt.Sprintf("regexp env key err %v", err)
				logrus.Errorf(errMsg)
				return errors.New(errMsg)
			}
			if !match {
				errMsg := fmt.Sprintf("key %s not match the regex express %s", k, regexString)
				logrus.Errorf(errMsg)
				return errors.New(errMsg)
			}
		}
	}

	return nil
}
