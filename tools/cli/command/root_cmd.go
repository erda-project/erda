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
	host      string // erda host, format: http[s]://<domain> eg: https://erda.cloud
	Remote    string // git remote name for erda repo
	username  string
	password  string
	debugMode bool
)

// Cmds which not require login
var (
	loginWhiteList = []string{
		"config <ops>",
		"config-set <write-ops> <name>",
		"erda init",
		"ext retag",
		"migrate",
		"migrate lint",
		"migrate mkpy",
		"migrate mkpypkg",
		"migrate record",
		"pipeline",
		"pipeline init",
		"pipeline check",
		"parse",
		"check",
		"version",
		"help",
	}
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "erda-cli",
	Short: "Erda commandline client",
	Long: `
    _/_/_/_/       _/_/_/        _/_/_/          _/_/    
   _/             _/    _/      _/    _/      _/    _/   
  _/_/_/         _/_/_/        _/    _/      _/_/_/_/    
 _/             _/    _/      _/    _/      _/    _/     
_/_/_/_/       _/    _/      _/_/_/        _/    _/      
`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logrus.SetOutput(os.Stdout)
		defer func() {
			cmd.SilenceErrors = true
		}()

		ctx.Debug = debugMode
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

		// TODO handle error
		u, err := getFullUse(cmd)
		if err != nil {
			err = fmt.Errorf(color_str.Red("✗ ") + err.Error())
			fmt.Println(err)
			return err
		}

		for _, w := range loginWhiteList {
			if w == u {
				return nil
			}
		}

		// parse and use context according to host param or config file
		if err := parseCtx(); err != nil {
			err = fmt.Errorf(color_str.Red("✗ ") + err.Error())
			fmt.Println(err)
			return err
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

func getFullUse(cmd *cobra.Command) (string, error) {
	if cmd.HasParent() {
		pUse, err := getFullUse(cmd.Parent())
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(strings.Join([]string{
			strings.TrimSpace(pUse), strings.TrimSpace(cmd.Use),
		}, " ")), nil
	}

	return "", nil
}

func ensureSessionInfos() (map[string]status.StatusInfo, error) {
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

	// TODO need git info ?
	//if username == "" || password == "" {
	//	gitCredentialStorage := fetchGitCredentialStorage()
	//	switch gitCredentialStorage {
	//	case "osxkeychain", "store":
	//		// fetch username & password from osxkeychain
	//		username, password = fetchGitUserInfo(gitCredentialStorage)
	//	}
	//}

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

func parseCtx() error {
	if host == "" {
		c, err := GetCurContext()
		if err != nil && err != dicedir.NotExist && !os.IsNotExist(err) {
			return err
		}

		host = c.Platform.Server
		ctx.CurrentOrg = *c.Platform.OrgInfo

		if _, err := os.Stat(".git"); err == nil {
			// fetch host from git remote url
			cmd := exec.Command("git", "remote", "get-url", Remote)
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

			if host == "" {
				host = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			} else {
				if !strings.Contains(host, u.Host) {
					fmt.Println(color_str.Yellow(
						fmt.Sprintf("current git repo remote %s: %s, different from config: %s", Remote, newStr, host)))
				}
			}
		}

		if host == "" {
			// fetch host from stdin
			fmt.Print("Enter your dice host: ")
			fmt.Scanln(&host)
		}
	}
	slashIndex := strings.Index(host, "://")
	if slashIndex < 0 {
		return errors.Errorf("invalid host format, it should be http[s]://<domain>")
	}
	hostHasOpenApi := strings.Index(host, "openapi.") != -1
	openAPIAddr := host
	if strings.HasPrefix(host, "https") {
		if !hostHasOpenApi {
			openAPIAddr = "https://openapi." + host[slashIndex+3:]
		}
	} else {
		if !hostHasOpenApi {
			openAPIAddr = "http://openapi." + host[slashIndex+3:]
		}
	}

	logrus.Debugf("openapi addr: %s", openAPIAddr)
	ctx.CurrentOpenApiHost = openAPIAddr

	return nil
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

	RootCmd.PersistentFlags().StringVar(&host, "host", "", "erda host to visit, eg: https://erda.cloud")
	RootCmd.PersistentFlags().StringVarP(&Remote, "remote", "r", "origin", "the remote for erda git repo")
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
