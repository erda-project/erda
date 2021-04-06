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

package gitflowutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsMaster(t *testing.T) {
	require.True(t, IsMaster("master"))
	require.False(t, IsMaster("master1"))
	require.False(t, IsMaster("master/1"))
	require.False(t, IsMaster("1master"))
	require.False(t, IsMaster("1master1"))
}

func TestIsSupport(t *testing.T) {
	require.True(t, IsSupport("support/1"))
	require.True(t, IsSupport("support/2"))
	require.False(t, IsSupport("support/"))
	require.False(t, IsSupport("support"))
	require.False(t, IsSupport("support1"))
	require.False(t, IsSupport("1support"))
	require.False(t, IsSupport("1support1"))
}

func TestIsHotfix(t *testing.T) {
	require.True(t, IsHotfix("hotfix/1"))
	require.True(t, IsHotfix("hotfix/2"))
	require.False(t, IsHotfix("hotfix/"))
	require.False(t, IsHotfix("hotfix"))
	require.False(t, IsHotfix("hotfix1"))
	require.False(t, IsHotfix("1hotfix"))
	require.False(t, IsHotfix("1hotfix1"))
}

func TestIsRelease(t *testing.T) {
	require.True(t, IsRelease("release/1"))
	require.True(t, IsRelease("release/2"))
	require.False(t, IsRelease("release/"))
	require.False(t, IsRelease("release"))
	require.False(t, IsRelease("release1"))
	require.False(t, IsRelease("1release"))
	require.False(t, IsRelease("1release1"))
}

func TestIsDevelop(t *testing.T) {
	require.True(t, IsDevelop("develop"))
	require.False(t, IsDevelop("develop1"))
	require.False(t, IsDevelop("develop/1"))
	require.False(t, IsDevelop("1develop"))
	require.False(t, IsDevelop("1develop1"))
}

func TestIsFeature(t *testing.T) {
	require.True(t, IsFeature("feature/1"))
	require.True(t, IsFeature("feature/2"))
	require.False(t, IsFeature("feature/"))
	require.False(t, IsFeature("feature"))
	require.False(t, IsFeature("feature1"))
	require.False(t, IsFeature("1feature"))
	require.False(t, IsFeature("1feature1"))
}

func TestIsXXXSlash(t *testing.T) {
	require.True(t, isXXXSlash("feature/1", "feature/"))
	require.False(t, isXXXSlash("feature/", "feature/"))
	require.False(t, isXXXSlash("1feature/", "feature/"))
}
