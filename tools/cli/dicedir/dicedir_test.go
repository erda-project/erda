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

package dicedir

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestFindGlobalDiceDir(t *testing.T) {
// 	u, _ := user.Current()
// 	assert.Nil(t, os.Mkdir(u.HomeDir+"/.dice.d", 0666))
// 	defer func() { os.Remove(u.HomeDir + "/.dice.d") }()

// 	_, err := FindGlobalDiceDir()
// 	assert.Nil(t, err)
// }

func TestFindProjectDiceDir(t *testing.T) {
	u, _ := user.Current()
	assert.Nil(t, os.Mkdir(u.HomeDir+"/.dice", 0666))
	defer func() { os.Remove(u.HomeDir + "/.dice") }()

	curr, err := os.Getwd()
	upper := filepath.Dir(curr)
	assert.Nil(t, err)
	assert.Nil(t, os.Mkdir(upper+"/.dice", 0666))
	defer func() { os.Remove(upper + "/.dice") }()

	s, err := FindProjectDiceDir()
	assert.Nil(t, err)
	fmt.Printf("%+v\n", s) // debug print

}
