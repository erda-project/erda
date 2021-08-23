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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gogap/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/terminal/color_str"
	"github.com/erda-project/erda/tools/cli/dicedir"
	"github.com/erda-project/erda/tools/cli/format"
	"github.com/erda-project/erda/tools/cli/status"
)

var (
	host      string // dice host, format: http[s]://<org>.<wildcard-domain> eg: https://terminus-org.app.terminus.io
	username  string
	password  string
	debugMode bool
)

// Cmds which not require login
var (
	loginWhiteList = []string{
		"init",
		"parse",
		"version",
		"migrate",
		"lint",
		"mkpy",
		"mkpypkg",
		"record",
		"help",
	}
	loginWhiteListCmds = strings.Join(loginWhiteList, ",")
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dice",
	Short: "Dice commandline client",
	Long: `
    _/_/_/_/       _/_/_/        _/_/_/          _/_/    
   _/             _/    _/      _/    _/      _/    _/   
  _/_/_/         _/_/_/        _/    _/      _/_/_/_/    
 _/             _/    _/      _/    _/      _/    _/     
_/_/_/_/       _/    _/      _/_/_/        _/    _/      
`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx.Debug = debugMode
		logrus.SetOutput(os.Stdout)
		defer func() {
			cmd.SilenceErrors = true
		}()

		httpOption := []httpclient.OpOption{httpclient.WithCompleteRedirect()}
		if debugMode {
			logrus.SetLevel(logrus.DebugLevel)
			httpOption = append(httpOption, httpclient.WithDebug(os.Stdout))
		} else {
			httpOption = append(httpOption, httpclient.WithLoadingPrint(""))
		}
		if strings.HasPrefix(host, "https") {
			httpOption = append(httpOption, httpclient.WithHTTPS())
		}
		ctx.HttpClient = httpclient.New(httpOption...)

		if strings.Contains(loginWhiteListCmds, strings.Split(cmd.Use, " ")[0]) {
			return nil
		}

		sessionInfos, err := ensureSessionInfos()
		if err != nil {
			err = fmt.Errorf(color_str.Red("✗ ") + err.Error())
			fmt.Println(err)
			return err
		}
		ctx.Sessions = sessionInfos

		return nil
	},
}

func setHost() error {
	if host == "" {
		if _, err := os.Stat(".git"); err != nil {
			// fetch host from stdin
			if os.IsNotExist(err) {
				fmt.Print("Enter your dice host: ")
				fmt.Scanln(&host)
			}
			return err
		} else {
			// fetch host from git remote url
			cmd := exec.Command("git", "remote", "get-url", "origin")
			out, err := cmd.CombinedOutput()
			if err != nil {
				return err
			}
			// remove crlf, otherwise parse error
			re := regexp.MustCompile(`\r?\n`)
			newStr := re.ReplaceAllString(string(out), "")
			u, err := url.Parse(newStr)
			if err != nil {
				return err
			}
			host = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		}
	}
	dotIndex := strings.Index(host, ".")
	if dotIndex < 0 {
		return errors.Errorf("invalid host format, it should be <org>.<wildcard-domain>")
	}
	hostHasOpenApi := strings.Index(host, "openapi-") != -1
	var openAPIAddr string
	if strings.HasPrefix(host, "https") {
		if hostHasOpenApi {
			openAPIAddr = "https://" + host
		} else {
			openAPIAddr = "https://openapi" + host[dotIndex:]
		}
	} else {
		if hostHasOpenApi {
			openAPIAddr = "http://" + host
		} else {
			openAPIAddr = "http://openapi" + host[dotIndex:]
		}
	}
	logrus.Debugf("openapi addr: %s", openAPIAddr)
	ctx.CurrentOpenApiHost = openAPIAddr

	return nil
}

func ensureSessionInfos() (map[string]status.StatusInfo, error) {
	if err := setHost(); err != nil {
		return nil, err
	}
	sessionInfos, err := status.GetSessionInfos()
	if err != nil && err != dicedir.NotExist {
		return nil, err
	}
	// file ~/.dice.d/sessions exist & and session for host also exist; otherwise need login fisrt
	if currentSession, ok := sessionInfos[ctx.CurrentOpenApiHost]; ok {
		// check session if expired
		if currentSession.ExpiredAt != nil && time.Now().Before(*currentSession.ExpiredAt) {
			return sessionInfos, nil
		}
	}

	if username == "" || password == "" {
		gitCredentialStorage := fetchGitCredentialStorage()
		switch gitCredentialStorage {
		case "osxkeychain", "store":
			// fetch username & password from osxkeychain
			username, password = fetchGitUserInfo(gitCredentialStorage)
		}
	}

	if username == "" {
		username = inputNormal("Enter your dice username: ")
	}
	if password == "" {
		password = inputPWD("Enter your dice password: ")
	}

	// fetch session & user info according to host, username & password
	if err = loginAndStoreSession(ctx.CurrentOpenApiHost, username, password); err != nil {
		return nil, err
	}

	// fetch sessions again
	sessionInfos, err = status.GetSessionInfos()
	if err != nil {
		return nil, err
	}

	return sessionInfos, nil
}

func fetchGitCredentialStorage() string {
	c, err := exec.Command("git", "config", "--global", "credential.helper").Output()
	if err != nil {
		fmt.Printf("fetch git credential err: %v", err)
		return ""
	}

	return strings.TrimSuffix(string(c), "\n")
}

func fetchGitUserInfo(credentialStorage string) (string, string) {
	u, err := url.Parse(host)
	if err != nil {
		fmt.Println(err)
		return "", ""
	}
	c1 := exec.Command("echo", fmt.Sprintf("host=%s", u.Hostname()))
	c2 := exec.Command("git", fmt.Sprintf("credential-%s", credentialStorage), "get")

	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var buf bytes.Buffer
	c2.Stdout = &buf

	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()

	sl := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(sl) < 2 {
		return "", ""
	}

	var (
		username string
		password string
	)
	for _, v := range sl {
		if strings.HasPrefix(v, "username=") {
			username = strings.TrimSpace(strings.SplitN(v, "=", 2)[1])
		} else if strings.HasPrefix(v, "password=") {
			password = strings.TrimSpace(strings.SplitN(v, "=", 2)[1])
		}
	}

	return username, password
}

func loginAndStoreSession(host, username, password string) error {
	form := make(url.Values)
	form.Set("username", username)
	form.Set("password", password)

	logrus.Debugf("current ctx: %+v", ctx)
	var body bytes.Buffer
	res, err := ctx.Post().Path("/login").FormBody(form).Do().Body(&body)
	if err != nil {
		return fmt.Errorf(format.FormatErrMsg("login", "error: "+err.Error(), false))
	}
	if !res.IsOK() {
		return fmt.Errorf(format.FormatErrMsg("login",
			"failed to login, status code: "+strconv.Itoa(res.StatusCode()), false))
	}
	var s status.StatusInfo
	d := json.NewDecoder(&body)
	if err := d.Decode(&s); err != nil {
		return fmt.Errorf(format.FormatErrMsg(
			"login", "failed to  decode login response ("+err.Error()+")", false))
	}
	// 从 openapi 获取的 session 无过期时间，暂设 12 小时，小于 openapi 的 24 小时
	expiredAt := time.Now().Add(time.Hour * 12)
	s.ExpiredAt = &expiredAt

	// TODO set orgID after login, get org info by org name
	if err := status.StoreSessionInfo(host, s); err != nil {
		return err
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// hind cursor
	tput("civis")
	// unhind cursor
	defer tput("cnorm")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		for range c {
			tput("cnorm")
			os.Exit(1)
		}
	}()

	RootCmd.PersistentFlags().StringVar(&host, "host", "", "dice host to visit, format: <org>.<wildcard domain>, eg: https://terminus-org.app.terminus.io")
	RootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "dice username to authenticate")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "dice password to authenticate")
	RootCmd.PersistentFlags().BoolVarP(&debugMode, "verbose", "V", false, "enable verbose mode")

	RootCmd.Execute()
}

func tput(arg string) error {
	cmd := exec.Command("tput", arg)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func inputPWD(prompt string) string {
	cmd := exec.Command("stty", "-echo")
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	defer func() {
		cmd := exec.Command("stty", "echo")
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		fmt.Println("")
	}()
	return inputNormal(prompt)
}

func inputNormal(prompt string) string {
	fmt.Printf(prompt)
	r := bufio.NewReader(os.Stdin)
	input, err := r.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return input[:len(input)-1]
}
