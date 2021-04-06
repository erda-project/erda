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

package color_str

import (
	"strings"
)

//          foreground background
// black        30         40
// red          31         41
// green        32         42
// yellow       33         43
// blue         34         44
// magenta      35         45
// cyan         36         46
// white        37         47

// reset             0  (everything back to normal)
// bold/bright       1  (often a brighter shade of the same colour)
// underline         4
// inverse           7  (swap foreground and background colours)

type Option string

func join(op []Option, i string) string {
	ops := []string{}
	for _, o := range op {
		ops = append(ops, string(o))
	}
	return strings.Join(ops, i)
}

var (
	BlackBg   Option = "40"
	RedBg     Option = "41"
	GreenBg   Option = "42"
	YellowBg  Option = "43"
	BlueBg    Option = "44"
	MagentaBg Option = "45"
	CyanBg    Option = "46"
	WhiteBg   Option = "47"

	Reset     Option = "0"
	Bold      Option = "1"
	Underline Option = "4"
	Inverse   Option = "7"
)

func Black(s string, op ...Option) string {
	return "\033[" + join(append(op, Option("30")), ";") + "m" + s + "\033[0m"
}
func Red(s string, op ...Option) string {
	return "\033[" + join(append(op, Option("31")), ";") + "m" + s + "\033[0m"
}
func Green(s string, op ...Option) string {
	return "\033[" + join(append(op, Option("32")), ";") + "m" + s + "\033[0m"
}
func Yellow(s string, op ...Option) string {
	return "\033[" + join(append(op, Option("33")), ";") + "m" + s + "\033[0m"
}
func Blue(s string, op ...Option) string {
	return "\033[" + join(append(op, Option("34")), ";") + "m" + s + "\033[0m"
}
func Magenta(s string, op ...Option) string {
	return "\033[" + join(append(op, Option("35")), ";") + "m" + s + "\033[0m"
}
func Cyan(s string, op ...Option) string {
	return "\033[" + join(append(op, Option("36")), ";") + "m" + s + "\033[0m"
}
func White(s string, op ...Option) string {
	return "\033[" + join(append(op, Option("37")), ";") + "m" + s + "\033[0m"
}
