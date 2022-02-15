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

package eventbox

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/erda-project/erda-proto-go/core/messenger/notify/pb"
	"github.com/erda-project/erda/modules/core-services/services/dingtalk/api/interfaces"
	"github.com/erda-project/erda/modules/eventbox/dispatcher"
)

func Initialize(dingtalk interfaces.DingTalkApiClientFactory, messenger pb.NotifyServiceServer) error {
	dp, err := dispatcher.New(dingtalk, messenger)
	if err != nil {
		panic(err)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for range sig {
			dp.Stop()
			os.Exit(0)
		}
	}()

	dp.Start()
	return nil
}
