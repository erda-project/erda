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

package pattern_test

import (
	"os"
	"testing"

	"github.com/erda-project/erda/pkg/database/pyorm/pattern"
)

func TestGenSettings(t *testing.T) {
	var settings = pattern.Settings{
		Engine:        pattern.DjangoMySQLEngine,
		User:          "root",
		Password:      "12345678",
		Host:          "3306",
		Port:          3306,
		Name:          "erda",
		TimeZone:      pattern.TimeZoneAsiaShanghai,
		InstalledApps: "myfeature",
	}
	if err := pattern.GenSettings(os.Stdout, settings); err != nil {
		t.Fatal(err)
	}
}
