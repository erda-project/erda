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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gogap/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/terminal/color_str"
	"github.com/erda-project/erda/tools/cli/status"
	"github.com/erda-project/erda/tools/cli/utils"
)

var (
	host         string // erda host, format: http[s]://<domain> eg: https://erda.cloud
	Remote       string // git remote name for erda repo
	OrgName      string // preferred org name override
	ProjectName  string // preferred project name override
	OrgID        uint64 // preferred org id override
	ProjectID    uint64 // preferred project id override
	username     string
	password     string
	debugMode    bool
	Interactive  bool
	IsCompletion bool
)

var (
	getEnv            = os.Getenv
	getGlobalConfig   = GetGlobalConfig
	getSessionInfos   = status.GetSessionInfos
	storeSessionInfo  = status.StoreSessionInfo
	deleteSessionInfo = status.DeleteSessionInfo
	parseContext      = parseCtx
	inputNormal       = utils.InputNormal
	inputPassword     = utils.InputPWD
	statPath          = os.Stat
	getWorkspaceInfo  = utils.GetWorkspaceInfo
)

var loginAndStoreSession = loginAndStoreSessionWithPassword

var preferWorkspaceHost bool
const PreferWorkspaceHostAnnotationKey = "erda.prefer_workspace_host"

var (
	commandErrorOutput  io.Writer = os.Stdout
	commandErrorPrinted bool
	cursorStateMu       sync.Mutex
	cursorHidden        bool
	exitWithCode        = os.Exit
)

// Default portal origin (same host family as git clone URLs, e.g. https://erda.cloud/org/dop/...).
// parseCtx then calls FetchOpenapi(), which reads /metadata.json here and switches API calls to openapi_public_url.
const defaultErdaHost = "https://erda.cloud"

// Cmds which not require login
var (
	loginWhiteList = []string{
		"completion",
		"ext retag",
		"gorm gen",
		"migrate",
		"migrate lint",
		"migrate mkpy",
		"migrate mkpypkg",
		"migrate record",
		"gw debug-auth",
		"login",
		"logout",
		"pipeline",
		"update",
		"update check",
		"update list",
		"update set-default",
		"update set-default [channel]",
		"version",
		"whoami",
		"help",
		"help [command]",
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
	SilenceUsage:      true,
	PersistentPreRunE: PrepareCtx,
}

type BaseResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Err     interface{}     `json:"err"`
}

type OrgsResponseData struct {
	List  []OrgInfo `json:"list"`
	Total int       `json:"total"`
}

type loginResponseUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Nick  string `json:"nick"`
}

type loginResponseToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type loginResponse struct {
	User  *loginResponseUser  `json:"user"`
	Token *loginResponseToken `json:"token"`
}

const credentialTypeBasic = "basic"

func PrepareCtx(cmd *cobra.Command, args []string) error {
	logrus.SetOutput(os.Stdout)
	hideCursorIfInteractive(tput)
	var err error
	defer func() {
		cmd.SilenceErrors = true
	}()
	defer func() {
		if !IsCompletion && err != nil {
			MarkCommandErrorPrinted()
			fmt.Println(err)
		}
	}()

	ctx.Debug = debugMode
	httpOption := []httpclient.OpOption{httpclient.WithCompleteRedirect()}
	if debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		httpOption = append(httpOption, httpclient.WithDebug(os.Stdout))
	} else if Interactive {
		httpOption = append(httpOption, httpclient.WithLoadingPrint(""))
	}
	if strings.HasPrefix(host, "https") {
		httpOption = append(httpOption, httpclient.WithHTTPS())
	}
	ctx.HttpClient = httpclient.New(httpOption...)

	u, err := getFullUse(cmd)
	if err != nil {
		err = fmt.Errorf("%s", color_str.Red("✗ ")+err.Error())
		return err
	}
	preferWorkspaceHost = lookupPreferWorkspaceHost(cmd)

	// For completion zsh etc.
	if strings.HasPrefix(u, "completion ") || strings.HasPrefix(u, "__complete") {
		return nil
	}
	for _, w := range loginWhiteList {
		if w == u {
			return nil
		}
	}

	if strings.HasPrefix(u, "clone") || strings.HasPrefix(u, "push") {
		u, err := url.Parse(args[0])
		if err != nil {
			return err
		}
		host = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	}

	// parse and use context according to host param or config file
	if err := parseContext(); err != nil {
		err = fmt.Errorf("%s", color_str.Red("✗ ")+err.Error())
		return err
	}

	sessionInfos, err := ensureSessionInfos()
	if err != nil {
		err = fmt.Errorf("%s", color_str.Red("✗ ")+err.Error())
		return err
	}
	ctx.Sessions = sessionInfos

	return nil
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
	sessionInfos, err := getSessionInfos()
	if err != nil && err != utils.NotExist {
		return nil, err
	}
	// file ~/.erda.d/sessions exist & and session for host also exist; otherwise need login fisrt
	currentSession, ok := sessionInfos[ctx.CurrentHost]
	if ok && hasUsableAuth(currentSession) {
		// check session if expired
		if currentSession.ExpiredAt == nil || time.Now().Before(*currentSession.ExpiredAt) {
			return sessionInfos, nil
		}
	}
	if ok && currentSession.Credential != nil {
		credentialUsername, credentialPassword, err := decodeBasicAuth(currentSession.Credential.Auth)
		if err != nil {
			return nil, err
		}
		if currentSession.Credential.Type == credentialTypeBasic &&
			credentialUsername != "" &&
			credentialPassword != "" {
			if err = loginAndStoreSession(ctx.CurrentHost, credentialUsername, credentialPassword); err != nil {
				return nil, err
			}
			sessionInfos, err = getSessionInfos()
			if err != nil {
				return nil, err
			}
			return sessionInfos, nil
		}
	}

	if Interactive {
		printLoginPrompt(ctx.CurrentHost, ok && currentSession.ExpiredAt != nil)
		if username == "" {
			username = inputNormal("Username: ")
		}
		if password == "" {
			password = inputPassword("Password: ")
		}
	}

	if Interactive {
		if username == "" {
			return nil, errors.Errorf("username is required to login to %s", ctx.CurrentHost)
		}
		if password == "" {
			return nil, errors.Errorf("password is required to login to %s", ctx.CurrentHost)
		}
	}

	if username != "" && password != "" {
		// fetch session & user info according to host, username & password
		if err = loginAndStoreSession(ctx.CurrentHost, username, password); err != nil {
			return nil, err
		}

		// fetch sessions again
		sessionInfos, err = getSessionInfos()
		if err != nil {
			return nil, err
		}
	} else {
		if ok && currentSession.ExpiredAt != nil {
			return nil, errors.Errorf("session expired at %s", currentSession.ExpiredAt.String())
		}
		return nil, errors.New("not login yet, please login first")
	}

	return sessionInfos, nil
}

func printLoginPrompt(host string, sessionExpired bool) {
	if sessionExpired {
		fmt.Fprintf(contextOutput(), "Session expired for %s, please log in again.\n", host)
		return
	}
	fmt.Fprintf(contextOutput(), "Login required for %s\n", host)
}

func LoadAuthState() error {
	if err := parseContext(); err != nil {
		return err
	}
	sessionInfos, err := ensureSessionInfos()
	if err != nil {
		if err == utils.NotExist {
			ctx.Sessions = map[string]status.StatusInfo{}
			return nil
		}
		return err
	}
	ctx.Sessions = sessionInfos
	return nil
}

func Login() error {
	if err := parseContext(); err != nil {
		return err
	}
	sessionInfos, err := ensureSessionInfos()
	if err != nil {
		return err
	}
	ctx.Sessions = sessionInfos
	return SaveDefaultHost(ctx.CurrentHost)
}

func Logout() error {
	if err := parseContext(); err != nil {
		return err
	}
	sessionInfos, err := getSessionInfos()
	if err != nil {
		if err == utils.NotExist {
			ctx.Sessions = map[string]status.StatusInfo{}
			return errors.New("not login yet, please login first")
		}
		return err
	}
	ctx.Sessions = sessionInfos
	if _, ok := ctx.CurrentAuthInfo(); !ok {
		return errors.New("not login yet, please login first")
	}
	if err := deleteSessionInfo(ctx.CurrentHost); err != nil {
		return err
	}
	ctx.Sessions = map[string]status.StatusInfo{}
	return nil
}

func SaveDefaultHost(defaultHost string) error {
	if defaultHost == "" {
		return nil
	}
	configFile, globalConfig, err := GetGlobalConfig()
	if err != nil && err != utils.NotExist {
		return err
	}
	if err == utils.NotExist {
		configFile, err = utils.FindGlobalConfig()
		if err != nil && err != utils.NotExist {
			return err
		}
		if err == utils.NotExist {
			globalConfig = &GlobalConfig{Version: ConfigVersion}
		}
	}
	if globalConfig == nil {
		globalConfig = &GlobalConfig{Version: ConfigVersion}
	}
	globalConfig.Host = defaultHost
	globalConfig.Server = ""
	return SetGlobalConfig(configFile, globalConfig)
}

func resolveHostFromFlagOrEnv() (string, error) {
	if host != "" {
		return host, nil
	}

	if envHost := strings.TrimSpace(getEnv("ERDA_HOST")); envHost != "" {
		return envHost, nil
	}

	return "", nil
}

func resolveGlobalHost() (string, error) {
	_, globalConfig, err := getGlobalConfig()
	if err != nil && err != utils.NotExist {
		return "", err
	}
	if err == nil {
		if resolved := globalConfig.ResolvedHost(); resolved != "" {
			return resolved, nil
		}
	}
	return "", nil
}

func resolveBaseHost() (string, error) {
	resolved, err := resolveHostFromFlagOrEnv()
	if err != nil || resolved != "" {
		return resolved, err
	}

	return resolveGlobalHost()
}

// lookupPreferWorkspaceHost returns true when the active command or an ancestor
// was generated with PreferWorkspaceHost (workspace git remote drives API host).
func lookupPreferWorkspaceHost(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Annotations[PreferWorkspaceHostAnnotationKey] == "true" {
			return true
		}
	}
	return false
}

func loadWorkspaceInfo() (utils.GitterURLInfo, bool, error) {
	if _, err := statPath(".git"); err != nil {
		if os.IsNotExist(err) {
			return utils.GitterURLInfo{}, false, nil
		}
		return utils.GitterURLInfo{}, false, err
	}

	info, err := getWorkspaceInfo(".", Remote)
	if err != nil {
		if err == utils.InvalidErdaRepo {
			return utils.GitterURLInfo{}, false, nil
		}
		return utils.GitterURLInfo{}, false, err
	}
	return info, true, nil
}

func parseCtx() error {
	resolvedHost, err := resolveHostFromFlagOrEnv()
	if err != nil {
		return err
	}
	host = resolvedHost

	_, config, err := GetProjectConfig()
	if err != nil && err != utils.NotExist {
		return err
	}
	if err == nil {
		ctx.CurrentOrg.ID = config.OrgID
		ctx.CurrentOrg.Name = config.Org
		ctx.CurrentProject.ProjectID = config.ProjectID
		ctx.CurrentProject.Project = config.Project
		ctx.Applications = nil
		for _, a := range config.Applications {
			a2 := ApplicationInfo{
				a.Application,
				a.ApplicationID,
				a.Mode,
				a.Desc,
				a.Sonarhost,
				a.Sonartoken,
				a.Sonarproject,
			}
			ctx.Applications = append(ctx.Applications, a2)
		}
	}
	if OrgID > 0 {
		ctx.CurrentOrg.ID = OrgID
	}
	if ProjectID > 0 {
		ctx.CurrentProject.ProjectID = ProjectID
	}
	if strings.TrimSpace(OrgName) != "" {
		ctx.CurrentOrg.Name = strings.TrimSpace(OrgName)
	}
	if strings.TrimSpace(ProjectName) != "" {
		ctx.CurrentProject.Project = strings.TrimSpace(ProjectName)
	}

	workspaceInfo, hasWorkspaceInfo, err := loadWorkspaceInfo()
	if err != nil {
		return err
	}
	if hasWorkspaceInfo {
		for _, a := range ctx.Applications {
			if a.Application == workspaceInfo.Application {
				ctx.CurrentApplication.Application = a.Application
				ctx.CurrentApplication.ApplicationID = a.ApplicationID
			}
		}

		if host == "" && preferWorkspaceHost {
			host = fmt.Sprintf("%s://%s", workspaceInfo.Scheme, workspaceInfo.Host)

			if username == "" || password == "" {
				gitCredentialStorage := fetchGitCredentialStorage()
				switch gitCredentialStorage {
				case "osxkeychain", "store":
					username, password, _ = fetchGitUserInfo(workspaceInfo.Host, gitCredentialStorage)
				}
			}
		}
	}

	if host == "" {
		resolvedHost, err = resolveGlobalHost()
		if err != nil {
			return err
		}
		host = resolvedHost
	}

	if host == "" && hasWorkspaceInfo {
		host = fmt.Sprintf("%s://%s", workspaceInfo.Scheme, workspaceInfo.Host)
	}

	if host == "" {
		host = defaultErdaHost
	}

	slashIndex := strings.Index(host, "://")
	if slashIndex < 0 {
		return errors.Errorf("invalid host format, it should be http[s]://<domain>")
	}

	ctx.CurrentHost = host
	if err := (&ctx).FetchOpenapi(); err != nil {
		return err
	}

	return nil
}

func fetchGitCredentialStorage() string {
	c, err := exec.Command("git", "config", "credential.helper").Output()
	if err != nil {
		fmt.Printf("fetch git credential err: %v", err)
		return ""
	}

	return strings.TrimSuffix(string(c), "\n")
}

func fetchGitUserInfo(host, credentialStorage string) (string, string, error) {
	c1 := exec.Command("echo", fmt.Sprintf("host=%s", host))
	c2 := exec.Command("git", fmt.Sprintf("credential-%s", credentialStorage), "get")

	rs, err := utils.PipeCmds(c1, c2)
	if err != nil {
		return "", "", err
	}

	sl := strings.Split(strings.TrimSpace(rs), "\n")
	if len(sl) < 2 {
		return "", "", errors.New("Get user info from git failed")
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

	return username, password, nil
}

func loginAndStoreSessionWithPassword(host, username, password string) error {
	form := make(url.Values)
	form.Set("username", username)
	form.Set("password", password)

	logrus.Debugf("current ctx: %+v", ctx)
	var body bytes.Buffer
	request := ctx.UseOpenapi().Post().Path("/login").FormBody(form)
	res, err := request.Do().Body(&body)
	if err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg("login", "error: "+err.Error(), false))
	}
	if !res.IsOK() {
		ctx.Error("login not ok, url: %s, response body: %s", request.GetUrl(), body.String())
		return fmt.Errorf("%s", utils.FormatErrMsg("login",
			"failed to login, status code: "+strconv.Itoa(res.StatusCode())+string(res.Body()), false))

	}
	s, err := decodeLoginStatus(body.Bytes(), time.Now)
	if err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg(
			"login", "failed to  decode login response ("+err.Error()+")", false))

	}
	s.Credential = &status.Credential{
		Type: credentialTypeBasic,
		Auth: encodeBasicAuth(username, password),
	}

	if err := storeSessionInfo(host, s); err != nil {
		return err
	}
	return nil
}

func encodeBasicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func decodeBasicAuth(auth string) (string, string, error) {
	decodedAuth, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", err
	}
	username, password, ok := strings.Cut(string(decodedAuth), ":")
	if !ok {
		return "", "", errors.New("invalid basic auth credential")
	}
	return username, password, nil
}

func decodeLoginStatus(body []byte, now func() time.Time) (status.StatusInfo, error) {
	var loginResp loginResponse
	if err := json.Unmarshal(body, &loginResp); err == nil && loginResp.Token != nil && loginResp.Token.AccessToken != "" {
		loginStatus := status.StatusInfo{
			Token: formatAuthorization(loginResp.Token.TokenType, loginResp.Token.AccessToken),
		}
		if loginResp.User != nil {
			loginStatus.UserInfo = status.UserInfo{
				ID:       loginResp.User.ID,
				Email:    loginResp.User.Email,
				NickName: loginResp.User.Nick,
			}
		}
		if loginResp.Token.ExpiresIn > 0 {
			expiredAt := now().Add(time.Duration(loginResp.Token.ExpiresIn) * time.Second)
			loginStatus.ExpiredAt = &expiredAt
		}
		return loginStatus, nil
	}

	var legacyStatus status.StatusInfo
	if err := json.Unmarshal(body, &legacyStatus); err != nil {
		return status.StatusInfo{}, err
	}
	if !hasUsableAuth(legacyStatus) {
		return status.StatusInfo{}, fmt.Errorf("unrecognized login response")
	}
	return legacyStatus, nil
}

func formatAuthorization(tokenType, accessToken string) string {
	if accessToken == "" {
		return ""
	}
	if strings.Contains(accessToken, " ") {
		return accessToken
	}
	if tokenType == "" {
		return accessToken
	}
	return tokenType + " " + accessToken
}

func hasUsableAuth(s status.StatusInfo) bool {
	return s.Token != "" || s.SessionID != ""
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.PersistentFlags().StringVar(&host, "host", "", "Erda portal base URL (default: https://erda.cloud, same as git remote host; use your OpenAPI URL only if the deployment exposes API without portal; overrides ERDA_HOST and ~/.erda.d/config)")
	RootCmd.PersistentFlags().StringVarP(&Remote, "remote", "", "origin", "the remote for Erda repo")
	RootCmd.PersistentFlags().StringVar(&OrgName, "org", "", "organization name override for commands using workspace context")
	RootCmd.PersistentFlags().StringVar(&ProjectName, "project", "", "project name override for commands using workspace context")
	RootCmd.PersistentFlags().Uint64Var(&OrgID, "org-id", 0, "organization ID override for commands using workspace context")
	RootCmd.PersistentFlags().Uint64Var(&ProjectID, "project-id", 0, "project ID override for commands using workspace context")
	RootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Erda username to authenticate")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Erda password to authenticate")
	RootCmd.PersistentFlags().BoolVarP(&debugMode, "verbose", "V", false, "if true, enable verbose mode")
	RootCmd.PersistentFlags().BoolVarP(&Interactive, "interactive", "", true, "if true, interactive with user")

	defer restoreCursorIfHidden(tput)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	handleInterruptSignals(c, os.Exit, func() {
		restoreCursorIfHidden(tput)
	})

	if err := executeRootCommand(RootCmd); err != nil {
		exitWithCode(1)
	}
}

func MarkCommandErrorPrinted() {
	commandErrorPrinted = true
}

func executeRootCommand(root *cobra.Command) error {
	commandErrorPrinted = false
	err := root.Execute()
	if err != nil && !IsCompletion && !commandErrorPrinted {
		MarkCommandErrorPrinted()
		_, _ = fmt.Fprintln(commandErrorOutput, formatCommandError(err))
	}
	return err
}

func formatCommandError(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if strings.Contains(message, "accepts between 0 and 0 arg(s), received 1") {
		return fmt.Errorf("unknown subcommand or unexpected argument; run \"erda-cli --help\" or \"erda-cli <command> --help\"")
	}
	return err
}

func handleInterruptSignals(c <-chan os.Signal, exit func(int), restore func()) {
	go func() {
		for range c {
			restore()
			exit(1)
		}
	}()
}

func hideCursorIfInteractive(control func(string) error) {
	cursorStateMu.Lock()
	defer cursorStateMu.Unlock()

	if !Interactive || cursorHidden {
		return
	}
	_ = control("civis")
	cursorHidden = true
}

func restoreCursorIfHidden(control func(string) error) {
	cursorStateMu.Lock()
	defer cursorStateMu.Unlock()

	if !cursorHidden {
		return
	}
	_ = control("cnorm")
	cursorHidden = false
}

func tput(arg string) error {
	cmd := exec.Command("tput", arg)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
