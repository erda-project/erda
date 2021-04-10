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

// Keep the specified decimal places.
func (d Duration) FormatDuration(precision int) (string, error) {
	if precision < 0 {
		return "", fmt.Errorf("invalid precision: %v", precision)
	}
	val, base := int64(d), int64(time.Nanosecond)
	switch {
	case val <= int64(time.Microsecond):
		return d.Duration().String(), nil
	case val <= int64(time.Millisecond):
		base = int64(time.Microsecond)
	case val <= int64(time.Second):
		base = int64(time.Millisecond)
	default:
		base = int64(time.Second)
	}
	for i := 0; i < precision; i++ {
		base /= 10
	}
	if base > 1 {
		if (val % base) >= (base / 2) {
			return time.Duration((val/base + 1) * base).String(), nil
		}
		return time.Duration(val / base * base).String(), nil
	}
	return d.Duration().String(), nil
}
