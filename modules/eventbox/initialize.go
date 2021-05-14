// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package eventbox

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/erda-project/erda/modules/eventbox/dispatcher"
)

func Initialize() error {
	dp, err := dispatcher.New()
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
