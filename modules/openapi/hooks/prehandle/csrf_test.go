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

package prehandle

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestValidateToken(t *testing.T) {
	token, err := generateCSRFToken()
	assert.Nil(t, err)
	fmt.Printf("%x\n", token) // debug print
	_, err = validateCSRFToken(string(token))
	assert.Nil(t, err)
}
