// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"fmt"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var text string
	err := unmarshal(&text)
	if err != nil {
		return err
	}
	duration, err := time.ParseDuration(text)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return fmt.Sprint(time.Duration(d)), nil
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
