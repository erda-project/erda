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

package command

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/tools/cli/status"
	"github.com/erda-project/erda/tools/cli/utils"
)

func TestDecodeLoginStatusWithTokenResponse(t *testing.T) {
	now := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	body := []byte(`{
		"user": {
			"id": "1001",
			"email": "ash@example.com",
			"nick": "ash"
		},
		"token": {
			"access_token": "access-token",
			"expires_in": 3600,
			"token_type": "Bearer"
		}
	}`)

	info, err := decodeLoginStatus(body, func() time.Time { return now })
	require.NoError(t, err)
	require.Equal(t, "Bearer access-token", info.Token)
	require.Equal(t, "1001", info.ID)
	require.Equal(t, "ash@example.com", info.Email)
	require.Equal(t, "ash", info.NickName)
	require.NotNil(t, info.ExpiredAt)
	require.Equal(t, now.Add(time.Hour), *info.ExpiredAt)
}

func TestDecodeLoginStatusWithLegacySessionResponse(t *testing.T) {
	body := []byte(`{
		"sessionid": "legacy-session",
		"id": "1002",
		"email": "legacy@example.com",
		"nickName": "legacy"
	}`)

	info, err := decodeLoginStatus(body, time.Now)
	require.NoError(t, err)
	require.Equal(t, "legacy-session", info.SessionID)
	require.Equal(t, "1002", info.ID)
	require.Equal(t, "legacy@example.com", info.Email)
	require.Equal(t, "legacy", info.NickName)
	require.Empty(t, info.Token)
}

func TestContextCurrentAuthInfoFallsBackToOpenapiHost(t *testing.T) {
	ctx := Context{
		CurrentHost: "https://erda.example.com",
		Sessions: map[string]status.StatusInfo{
			"https://openapi.erda.example.com": {Token: "Bearer access-token"},
		},
	}

	var err error
	ctx.Domain, err = parseURLForTest("https://erda.example.com")
	require.NoError(t, err)
	ctx.Openapi, err = parseURLForTest("https://openapi.erda.example.com")
	require.NoError(t, err)

	info, ok := ctx.CurrentAuthInfo()
	require.True(t, ok)
	require.Equal(t, "Bearer access-token", info.Token)
}

func parseURLForTest(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

func TestResolveBaseHostPrefersExplicitHost(t *testing.T) {
	origHost, origGetEnv, origGetGlobalConfig := host, getEnv, getGlobalConfig
	t.Cleanup(func() {
		host = origHost
		getEnv = origGetEnv
		getGlobalConfig = origGetGlobalConfig
	})

	host = "https://flag.example.com"
	getEnv = func(string) string { return "https://env.example.com" }
	getGlobalConfig = func() (string, *GlobalConfig, error) {
		return "", &GlobalConfig{Host: "https://global.example.com"}, nil
	}

	resolved, err := resolveBaseHost()
	require.NoError(t, err)
	require.Equal(t, "https://flag.example.com", resolved)
}

func TestResolveBaseHostPrefersEnvOverGlobalConfig(t *testing.T) {
	origHost, origGetEnv, origGetGlobalConfig := host, getEnv, getGlobalConfig
	t.Cleanup(func() {
		host = origHost
		getEnv = origGetEnv
		getGlobalConfig = origGetGlobalConfig
	})

	host = ""
	getEnv = func(string) string { return "https://env.example.com" }
	getGlobalConfig = func() (string, *GlobalConfig, error) {
		return "", &GlobalConfig{Host: "https://global.example.com"}, nil
	}

	resolved, err := resolveBaseHost()
	require.NoError(t, err)
	require.Equal(t, "https://env.example.com", resolved)
}

func TestResolveBaseHostFallsBackToGlobalConfig(t *testing.T) {
	origHost, origGetEnv, origGetGlobalConfig := host, getEnv, getGlobalConfig
	t.Cleanup(func() {
		host = origHost
		getEnv = origGetEnv
		getGlobalConfig = origGetGlobalConfig
	})

	host = ""
	getEnv = func(string) string { return "" }
	getGlobalConfig = func() (string, *GlobalConfig, error) {
		return "", &GlobalConfig{Host: "https://global.example.com"}, nil
	}

	resolved, err := resolveBaseHost()
	require.NoError(t, err)
	require.Equal(t, "https://global.example.com", resolved)
}

func TestResolveBaseHostIgnoresMissingGlobalConfig(t *testing.T) {
	origHost, origGetEnv, origGetGlobalConfig := host, getEnv, getGlobalConfig
	t.Cleanup(func() {
		host = origHost
		getEnv = origGetEnv
		getGlobalConfig = origGetGlobalConfig
	})

	host = ""
	getEnv = func(string) string { return "" }
	getGlobalConfig = func() (string, *GlobalConfig, error) {
		return "", &GlobalConfig{}, utils.NotExist
	}

	resolved, err := resolveBaseHost()
	require.NoError(t, err)
	require.Empty(t, resolved)
}

func TestResolveBaseHostReturnsEmptyWithoutConfiguredSources(t *testing.T) {
	origHost, origGetEnv, origGetGlobalConfig := host, getEnv, getGlobalConfig
	t.Cleanup(func() {
		host = origHost
		getEnv = origGetEnv
		getGlobalConfig = origGetGlobalConfig
	})

	host = ""
	getEnv = func(string) string { return "" }
	getGlobalConfig = func() (string, *GlobalConfig, error) {
		return "", &GlobalConfig{}, nil
	}

	resolved, err := resolveBaseHost()
	require.NoError(t, err)
	require.Empty(t, resolved)
}

func TestExecuteRootCommandPrintsReturnedErrorWhenNotReported(t *testing.T) {
	origOutput := commandErrorOutput
	origPrinted := commandErrorPrinted
	t.Cleanup(func() {
		commandErrorOutput = origOutput
		commandErrorPrinted = origPrinted
	})

	var output bytes.Buffer
	commandErrorOutput = &output
	commandErrorPrinted = false

	root := &cobra.Command{Use: "erda-cli", SilenceUsage: true}
	logs := &cobra.Command{
		Use: "logs",
		RunE: func(_ *cobra.Command, _ []string) error {
			t.Fatal("RunE should not be called when required flag is missing")
			return nil
		},
	}
	logs.Flags().Uint64("pipelineID", 0, "")
	logs.MarkFlagRequired("pipelineID")
	root.AddCommand(logs)
	root.SetArgs([]string{"logs"})

	err := executeRootCommand(root)
	require.Error(t, err)
	require.Contains(t, output.String(), `required flag(s) "pipelineID" not set`)
}

func TestExecuteRootCommandDoesNotDuplicateReportedError(t *testing.T) {
	origOutput := commandErrorOutput
	origPrinted := commandErrorPrinted
	t.Cleanup(func() {
		commandErrorOutput = origOutput
		commandErrorPrinted = origPrinted
	})

	var output bytes.Buffer
	commandErrorOutput = &output
	commandErrorPrinted = false

	root := &cobra.Command{
		Use: "erda-cli",
		RunE: func(_ *cobra.Command, _ []string) error {
			MarkCommandErrorPrinted()
			_, _ = output.WriteString("already printed\n")
			return errors.New("boom")
		},
	}

	err := executeRootCommand(root)
	require.Error(t, err)
	require.Equal(t, 1, strings.Count(output.String(), "already printed"))
	require.NotContains(t, output.String(), "boom")
}

func TestHandleInterruptSignalsExits(t *testing.T) {
	sigCh := make(chan os.Signal, 1)
	exitCh := make(chan int, 1)
	restoreCh := make(chan struct{}, 1)

	handleInterruptSignals(sigCh, func(code int) {
		exitCh <- code
	}, func() {
		restoreCh <- struct{}{}
	})

	sigCh <- syscall.SIGINT

	require.Equal(t, 1, <-exitCh)
	require.Len(t, restoreCh, 1)
}

func TestCursorControlRespectsInteractiveFlag(t *testing.T) {
	origInteractive := Interactive
	origCursorHidden := cursorHidden
	t.Cleanup(func() {
		Interactive = origInteractive
		cursorHidden = origCursorHidden
	})

	var calls []string
	control := func(arg string) error {
		calls = append(calls, arg)
		return nil
	}

	Interactive = false
	cursorHidden = false
	hideCursorIfInteractive(control)
	require.Empty(t, calls)
	require.False(t, cursorHidden)

	Interactive = true
	hideCursorIfInteractive(control)
	require.Equal(t, []string{"civis"}, calls)
	require.True(t, cursorHidden)

	restoreCursorIfHidden(control)
	require.Equal(t, []string{"civis", "cnorm"}, calls)
	require.False(t, cursorHidden)
}
