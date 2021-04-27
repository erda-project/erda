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

package main

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/cmd/migrator/internal/config"
	"github.com/erda-project/erda/pkg/sqlparser/migrator"
)

func main() {
	go startSandbox()

	logrus.Infoln("Erda MySQL Migration 4 start working")
	mig, err := migrator.New(config.Config())
	if err != nil {
		logrus.Fatalf("failed to start Erda MySQL Migration 4: %v", err)
	}

	if err := mig.Run(); err != nil {
		logrus.Fatalf("failed to do migrating: %v", err)
	}

	logrus.Infoln("migrate complete.")
	os.Exit(0)
}

// startSandbox start a MySQL server in the container
func startSandbox() {
	logrus.Infoln("create sandbox")
	var command = "/usr/bin/run-mysqld"
	sandbox := exec.Command(command)
	if err := sandbox.Run(); err != nil {
		logrus.Fatalf("failed to run %s: %v", command, err)
	}
}
