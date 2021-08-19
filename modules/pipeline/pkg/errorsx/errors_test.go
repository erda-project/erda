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

package errorsx

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsPlatformError(t *testing.T) {
	pErr := PlatformErrorf("failed to do")
	err := Errorf("failed to do")
	assert.Equal(t, true, IsPlatformError(pErr))
	assert.Equal(t, false, IsPlatformError(err))
}

func TestIsUserError(t *testing.T) {
	uErr := UserErrorf("failed to do")
	err := Errorf("failed to do")
	assert.Equal(t, true, IsUserError(uErr))
	assert.Equal(t, false, IsUserError(err))
}

func TestIsContainUserError(t *testing.T) {
	pErr := PlatformErrorf("failed to do")
	uErr := UserErrorf("failed to do")
	assert.Equal(t, false, IsContainUserError(pErr))
	assert.Equal(t, true, IsContainUserError(uErr))
}

func TestIsNetworkError(t *testing.T) {
	sessionErr := errors.New("failed to find Session for client xxx")
	timeoutErr := errors.New("Get http://xxx.com: net/http TLS handshake timeout")
	normalErr := errors.New("failed to do")
	assert.Equal(t, true, IsNetworkError(sessionErr))
	assert.Equal(t, true, IsNetworkError(timeoutErr))
	assert.Equal(t, false, IsNetworkError(normalErr))
}
