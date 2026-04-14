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

import (
	"fmt"
	"time"

	"github.com/gogap/errors"

	"github.com/erda-project/erda/tools/cli/command"
)

var WHOAMI = command.Command{
	Name:      "whoami",
	ShortHelp: "show current erda authentication info",
	Example: `
$ erda-cli whoami
$ ERDA_HOST=https://erda.cloud erda-cli whoami
`,
	Run: RunWhoami,
}

func RunWhoami(ctx *command.Context) error {
	if err := command.LoadAuthState(); err != nil {
		return err
	}

	authInfo, ok := command.GetContext().CurrentAuthInfo()
	if !ok {
		return errors.New("not login yet, please login first")
	}
	if authInfo.ExpiredAt != nil && time.Now().After(*authInfo.ExpiredAt) {
		return fmt.Errorf("session expired at %s", authInfo.ExpiredAt.String())
	}

	fmt.Printf("Host: %s\n", command.GetContext().CurrentHost)
	fmt.Printf("UserID: %s\n", authInfo.ID)
	fmt.Printf("Email: %s\n", authInfo.Email)
	fmt.Printf("Nickname: %s\n", authInfo.NickName)
	if authInfo.ExpiredAt != nil {
		fmt.Printf("ExpiredAt: %s\n", authInfo.ExpiredAt.Format(time.RFC3339))
	} else {
		fmt.Println("ExpiredAt: never")
	}
	return nil
}
