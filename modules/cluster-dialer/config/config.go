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

package config

import "time"

type Config struct {
	Debug           bool          `default:"false" desc:"enable debug logging"`
	NeedClusterInfo bool          `default:"true" desc:"need agent register cluster info"`
	Timeout         time.Duration `default:"60s" desc:"default timeout"`
	Listen          string        `default:":80" desc:"listen address"`
}
