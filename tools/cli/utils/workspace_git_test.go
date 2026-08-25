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
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsWorkspaceGitRepository(t *testing.T) {
	repo := t.TempDir()
	gitCommand(t, repo, "init", "-q", "-b", "main")
	gitCommand(t, repo, "config", "user.email", "test@example.invalid")
	gitCommand(t, repo, "config", "user.name", "test")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "README.md"), []byte("test\n"), 0o600))
	gitCommand(t, repo, "add", "README.md")
	gitCommand(t, repo, "commit", "-qm", "initial commit")

	got, err := IsWorkspaceGitRepository(repo)
	require.NoError(t, err)
	require.True(t, got)

	worktree := filepath.Join(repo, "worktree")
	gitCommand(t, repo, "worktree", "add", "-q", worktree, "-b", "feature")
	info, err := os.Stat(filepath.Join(worktree, ".git"))
	require.NoError(t, err)
	require.False(t, info.IsDir())

	got, err = IsWorkspaceGitRepository(worktree)
	require.NoError(t, err)
	require.True(t, got)
}

func TestIsWorkspaceGitRepositoryRejectsNonRepository(t *testing.T) {
	got, err := IsWorkspaceGitRepository(t.TempDir())
	require.Error(t, err)
	require.False(t, got)
}

func TestIsWorkspaceGitRepositoryRejectsFakeGitFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: nowhere\n"), 0o600))

	got, err := IsWorkspaceGitRepository(dir)
	require.Error(t, err)
	require.False(t, got)
}

func gitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
}
