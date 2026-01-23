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

package helper

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRemoveEndMarkerFromHeader(t *testing.T) {
	tt := []struct {
		header string
		want   string
	}{
		{"00aa0000000000000000000000000000000000000000 5cf4abb264b4a3a0d99e2b967ef8980a1cc41e77 refs/heads/test1 report-status-v2 side-band-64k object-format=sha1 agent=git/2.33.10000",
			"00aa0000000000000000000000000000000000000000 5cf4abb264b4a3a0d99e2b967ef8980a1cc41e77 refs/heads/test1 report-status-v2 side-band-64k object-format=sha1 agent=git/2.33.1",
		},
		{
			"00aa0000000000000000000000000000000000000000 5cf4abb264b4a3a0d99e2b967ef8980a1cc41e77 refs/heads/test2 report-status-v2 side-band-64k object-format=sha1 agent=git/2.33.100660000000000000000000000000000000000000000 5cf4abb264b4a3a0d99e2b967ef8980a1cc41e77 refs/heads/test30000",
			"00aa0000000000000000000000000000000000000000 5cf4abb264b4a3a0d99e2b967ef8980a1cc41e77 refs/heads/test2 report-status-v2 side-band-64k object-format=sha1 agent=git/2.33.100660000000000000000000000000000000000000000 5cf4abb264b4a3a0d99e2b967ef8980a1cc41e77 refs/heads/test3",
		},
	}
	for _, v := range tt {
		assert.Equal(t, v.want, string(removeEndMarkerFromHeader([]byte(v.header))))
	}
}

func TestGitCommand_ContextCancellation(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	// 1. Arrange
	// Use explicit cancel instead of timeout for better control
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use git alias to sleep, simulating a long running command
	cmd := gitCommand(ctx, "", "-c", "alias.wait=!sleep 10", "wait")

	// Start the command but don't wait yet
	err := cmd.Start()
	assert.NoError(t, err)

	// Verify process is running
	if cmd.Process == nil {
		t.Fatal("process is nil")
	}
	pid := cmd.Process.Pid

	// Check if process exists
	p, err := os.FindProcess(pid)
	assert.NoError(t, err)
	// On Unix, Verify process is alive
	err = p.Signal(syscall.Signal(0))
	assert.NoError(t, err, "process should be running")

	// 2. Act
	// Trigger cancellation manually
	cancel()

	// Wait a bit to ensure signal propagation and process termination
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// 3. Assert
		// The command should return an error because it was killed
		assert.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("command did not exit within 2 seconds after context cancellation")
	}
}

// TestExecCommand_WithoutContext proves that plain exec.Command does NOT respond to context cancellation.
// This is the contrast case to demonstrate why CommandContext is necessary.
func TestExecCommand_WithoutContext(t *testing.T) {
	if _, err := exec.LookPath("sleep"); err != nil {
		t.Skip("sleep not found in PATH")
	}

	// 1. Arrange
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use plain exec.Command (NOT CommandContext) - this simulates old behavior
	cmd := exec.Command("sleep", "10")

	// Start the command
	err := cmd.Start()
	assert.NoError(t, err)

	// Verify process is running
	if cmd.Process == nil {
		t.Fatal("process is nil")
	}
	pid := cmd.Process.Pid

	p, err := os.FindProcess(pid)
	assert.NoError(t, err)
	err = p.Signal(syscall.Signal(0))
	assert.NoError(t, err, "process should be running")

	// 2. Act
	// Cancel the context - but since we used exec.Command (not CommandContext),
	// this should have NO EFFECT on the process
	cancel()

	// Wait a short period
	time.Sleep(500 * time.Millisecond)

	// 3. Assert
	// Process should STILL be running because exec.Command ignores context
	err = p.Signal(syscall.Signal(0))
	assert.NoError(t, err, "process should STILL be running after context cancel (exec.Command ignores context)")

	// Clean up: manually kill the process
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}
