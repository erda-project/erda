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

package cmd

import "github.com/erda-project/erda/tools/cli/command"

var LOGIN = command.Command{
	Name:      "login",
	ShortHelp: "login and persist the default erda host",
	Example: `
$ erda-cli login -u yourname
$ erda-cli login --host https://erda.cloud -u yourname
$ ERDA_HOST=https://erda.cloud erda-cli login
$ erda-cli login --host https://openapi.erda.cloud -u yourname
`,
	Run: RunLogin,
}

func RunLogin(ctx *command.Context) error {
	if err := command.Login(); err != nil {
		return err
	}

	if authInfo, ok := command.GetContext().CurrentAuthInfo(); ok {
		command.GetContext().Succ("logged in to %s as %s", command.GetContext().CurrentHost, authInfo.NickName)
	}
	return nil
}
