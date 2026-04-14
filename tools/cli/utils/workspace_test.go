// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGitRemoteURLWithHTTPS(t *testing.T) {
	u, err := parseGitRemoteURL("https://erda.cloud/erda/dop/erda-project/erda")
	require.NoError(t, err)
	require.Equal(t, "https", u.Scheme)
	require.Equal(t, "erda.cloud", u.Host)
	require.Equal(t, "/erda/dop/erda-project/erda", u.Path)
}

func TestParseGitRemoteURLWithSSHScpStyle(t *testing.T) {
	u, err := parseGitRemoteURL("git@github.com:iutx/erda.git")
	require.NoError(t, err)
	require.Equal(t, "ssh", u.Scheme)
	require.Equal(t, "github.com", u.Host)
	require.Equal(t, "/iutx/erda.git", u.Path)
}

func TestParseGitRemoteURLRejectsInvalidRemote(t *testing.T) {
	_, err := parseGitRemoteURL(":::")
	require.Error(t, err)
}
