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

package auth

import (
	"errors"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar/conf"
	"github.com/erda-project/erda/modules/gittar/models"
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/modules/gittar/uc"
	"github.com/erda-project/erda/modules/gittar/webcontext"
	"github.com/erda-project/erda/pkg/ucauth"
)

var (
	authCache = cache.New(3*time.Minute, 1*time.Minute)
)

var (
	NO_AUTH_ERROR = errors.New("no auth info")
)

func renderTemplate(template string, params map[string]string) string {
	paramPattern, _ := regexp.Compile("{{.+?}}")
	result := paramPattern.ReplaceAllStringFunc(template, func(s string) string {
		key := s[2:(len(s) - 2)]
		value, ok := params[key]
		if ok {
			return value
		} else {
			return s
		}
	})
	return result
}

func Authenticate(c *webcontext.Context) {
	// Repository content
	orgProject := c.Param("org")
	appName := strings.TrimSuffix(c.Param("repo"), ".git")

	repoPath := filepath.Join(orgProject, appName)

	echoReqPath := c.EchoContext.Path()
	repo, err := c.Service.GetRepoByPath(repoPath)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound, errors.New("repo not found"))
		return
	}

	//没有子path尝试重定向到UI
	if c.EchoContext.Request().Method == "GET" && (echoReqPath == "/:org/:repo" || echoReqPath == "/:org/:repo/*") && c.EchoContext.QueryString() == "" {
		params := map[string]string{
			"projectId": strconv.FormatInt(repo.ProjectID, 10),
			"appId":     strconv.FormatInt(repo.AppID, 10),
			"orgId":     strconv.FormatInt(repo.OrgID, 10),
		}
		orgDto, err := c.Bundle.GetOrg(repo.OrgID)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound, errors.New("org not found"))
			return
		}
		redirectUrlPrefix := "http://" + orgDto.Domain
		c.EchoContext.Redirect(301, redirectUrlPrefix+renderTemplate(conf.RepoPathTemplate(), params))
		return
	}

	doAuth(c, repo, repoPath)
}

func AuthenticateByApp(c *webcontext.Context) {
	appIDStr := c.Param("appId")
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound, errors.New("invalid appId"))
		return
	}
	repo, err := c.Service.GetRepoByApp(appID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatus(http.StatusNotFound, errors.New("repo not found"))
			return
		}
		c.EchoContext.String(500, err.Error())
		return
	}

	repoName := path.Join(repo.OrgName, repo.ProjectName, repo.AppName)
	doAuth(c, repo, repoName)
}

func AuthenticateV2(c *webcontext.Context) {
	host := c.EchoContext.Request().Host
	// 优先域名 第二优先读取ORG-ID头
	dto, err := c.Bundle.GetDopOrgByDomain(host, "x")
	var orgID int64
	if err == nil {
		orgID = int64(dto.ID)
	} else {
		orgIdStr := c.HttpRequest().Header.Get("Org-ID")
		orgID, err = strconv.ParseInt(orgIdStr, 10, 64)
		if err != nil {
			c.EchoContext.String(404, "org not found")
			return
		}
	}
	project := c.Param("project")
	app := strings.TrimSuffix(c.Param("app"), ".git")
	repo, err := c.Service.GetRepoByNames(orgID, project, app)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatus(http.StatusNotFound, errors.New("repo not found"))
			return
		}
		c.EchoContext.String(500, err.Error())
		return
	}

	//todo delete redirect
	echoReqPath := c.EchoContext.Path()
	//没有子path尝试重定向到UI
	if c.EchoContext.Request().Method == "GET" &&
		(echoReqPath == "/wb/:project/:app" || echoReqPath == "/wb/:project/:app/*") && c.EchoContext.QueryString() == "" {
		params := map[string]string{
			"projectId": strconv.FormatInt(repo.ProjectID, 10),
			"appId":     strconv.FormatInt(repo.AppID, 10),
			"orgId":     strconv.FormatInt(repo.OrgID, 10),
		}
		redirectUrlPrefix := c.EchoContext.Scheme() + "://" + c.Host()
		c.EchoContext.Redirect(301, redirectUrlPrefix+renderTemplate(conf.RepoPathTemplate(), params))
		return
	}

	repoName := path.Join(strconv.FormatInt(orgID, 10), project, app)
	doAuth(c, repo, repoName)
}

func AuthenticateV3(c *webcontext.Context) {
	org := c.Param("org")
	project := c.Param("project")
	app := strings.TrimSuffix(c.Param("app"), ".git")

	orgID, err := getOrgIDV3(c, org)
	if err != nil {
		c.EchoContext.String(404, "org not found")
		return
	}
	repo, err := c.Service.GetRepoByNames(orgID, project, app)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatus(http.StatusNotFound, errors.New("repo not found"))
			return
		}
		c.EchoContext.String(500, err.Error())
		return
	}

	repoName := path.Join(strconv.FormatInt(orgID, 10), project, app)
	doAuth(c, repo, repoName)
}

func doAuth(c *webcontext.Context, repo *models.Repo, repoName string) {
	// Git Protocol version v2
	version := c.GetHeader("Git-Protocol")
	c.Set("gitProtocol", version)

	var gitRepository = &gitmodule.Repository{}
	var err error
	var userInfo *ucauth.UserInfo
	//skip authentication
	host := c.Host()
	for _, skipUrl := range conf.SkipAuthUrls() {
		if skipUrl != "" && strings.HasSuffix(host, skipUrl) {
			logrus.Debugf("skip authenticate host: %s", host)
			gitRepository, err = openRepository(repo)
			if err != nil {
				c.AbortWithStatus(500, err)
				return
			}
			c.Set("repository", gitRepository)
			c.Set("user", models.NewInnerUser())

			//只有在skipAuth范围内,如果读取到了user-id,也触发校验
			userIdStr := c.GetHeader("User-Id")
			if userIdStr != "" {
				userInfoDto, err := uc.FindUserById(userIdStr)
				if err != nil {
					c.AbortWithStatus(500, err)
					return
				}
				logrus.Infof("repo: %s userId: %v, username: %s", repoName, userIdStr, userInfoDto.Username)
				//校验通过缓存5分钟结果
				//校验失败每次都会请求
				_, validateError := ValidaUserRepoWithCache(c, userIdStr, repo)
				if validateError != nil {
					logrus.Infof("openapi auth fail repo:%s user:%s", repoName, userInfoDto.Username)
					c.AbortWithStatus(403, validateError)
					return
				}
				c.Set("repository", gitRepository)
				//c.Set("lock", repoLock.Lock)
				c.Set("user", &models.User{
					Name:     userInfoDto.Username,
					NickName: userInfoDto.NickName,
					Email:    userInfoDto.Email,
					Id:       userIdStr,
				})
				c.Next()
				return
			}
			logrus.Warn("no user user info ")
			c.Next()
			return
		}
	}

	//如果是内置账户不做校验
	innerUser, err := GetInnerUser(c)
	if err == nil {
		gitRepository, err = openRepository(repo)
		if err != nil {
			c.AbortWithStatus(500, err)
			return
		}
		c.Set("repository", gitRepository)
		c.Set("user", innerUser)
		c.Next()
		return
	}

	userInfo, err = GetUserInfoByTokenOrBasicAuth(c, repo.ProjectID)
	if err == nil {
		_, validateError := ValidaUserRepo(c, string(userInfo.ID), repo)
		if validateError != nil {
			c.AbortWithString(403, validateError.Error()+" 403")
		} else {
			gitRepository, err = openRepository(repo)
			if err != nil {
				c.AbortWithStatus(500, err)
				return
			}
			c.Set("repository", gitRepository)
			c.Set("user", &models.User{
				Name:     userInfo.UserName,
				NickName: userInfo.NickName,
				Email:    userInfo.Email,
				Id:       string(userInfo.ID)})
			c.Next()
		}
	} else {
		c.Header("WWW-Authenticate", "Basic realm=Restricted")
		if err == NO_AUTH_ERROR {
			c.AbortWithStatus(401)
		} else {
			c.AbortWithString(401, err.Error()+" 401")
		}
	}
}

func GetInnerUser(c *webcontext.Context) (*models.User, error) {
	username, password, ok := c.BasicAuth()
	if ok {
		if username == conf.GitInnerUserName() && password == conf.GitInnerUserPassword() {
			return models.NewInnerUser(), nil
		}
	}
	return nil, errors.New("not inner user")
}

type AuthResp struct {
	Permission *apistructs.ScopeRole
	Repo       *models.Repo
}

type ErrorData struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func GetUserInfoByTokenOrBasicAuth(c *webcontext.Context, projectID int64) (*ucauth.UserInfo, error) {
	var userInfo = &ucauth.UserInfo{}
	var err error

	org := c.Param("org")
	repo := strings.TrimSuffix(c.Param("repo"), ".git")
	repoName := filepath.Join(org, repo)

	username, password, ok := c.BasicAuth()
	if ok && username != "" && password != "" {
		// 触发token校验
		if username == conf.GitTokenUserName() {
			// dice token
			member, err := c.Bundle.GetMemberByToken(&apistructs.GetMemberByTokenRequest{
				Token: password,
			})
			if err == nil {
				userInfo, err := uc.FindUserById(member.UserID)
				if err == nil {
					return &ucauth.UserInfo{
						ID:        ucauth.USERID(member.UserID),
						Email:     userInfo.Email,
						Phone:     userInfo.Phone,
						AvatarUrl: userInfo.AvatarURL,
						UserName:  userInfo.Username,
						NickName:  userInfo.NickName,
					}, nil
				}
			} else {
				logrus.Errorf("GetMemberByToken err:%s", err)
				return nil, errors.New("invalid token")
			}
		} else {
			userInfo, err = GetUserByBasicAuth(c, username, password)
			if err != nil {
				logrus.Debugf("basic auth failed repo:%s login_name:%s err:%v", repoName, username, err)
			} else {
				logrus.Debugf("basic auth success repo:%s login_name:%s user_name:%s", repoName, username, userInfo.UserName)
			}
		}
		return userInfo, err
	} else {
		logrus.Debugf("no auth info repo:%s", repoName)
		return nil, NO_AUTH_ERROR
	}

}
func GetUserByBasicAuth(c *webcontext.Context, username string, passwd string) (*ucauth.UserInfo, error) {
	token, err := c.UCAuth.PwdAuth(username, passwd)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("login success username: %s", username)
	userInfo, err := c.UCAuth.GetUserInfo(token)
	if err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func ValidaUserRepoWithCache(c *webcontext.Context, userId string, repo *models.Repo) (*AuthResp, error) {
	key := userId + "-" + repo.Path
	authResultCache, found := authCache.Get(key)
	if found {
		autuResp := authResultCache.(*AuthResp)
		// 只有缓存存在并且是有权限的才使用缓存
		if autuResp != nil && autuResp.Permission.Access {
			return authResultCache.(*AuthResp), nil
		}
	}
	result, err := ValidaUserRepo(c, userId, repo)
	if err != nil {
		return nil, err
	}
	authCache.Set(key, result, cache.DefaultExpiration)
	return result, err
}

func ValidaUserRepo(c *webcontext.Context, userId string, repo *models.Repo) (*AuthResp, error) {
	permission, err := c.Bundle.ScopeRoleAccess(userId, &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.AppScope,
			ID:   strconv.FormatInt(repo.AppID, 10),
		},
	})
	if err != nil {
		return nil, err
	}
	if !permission.Access {
		return nil, errors.New("no permission to access")
	}
	return &AuthResp{
		Repo:       repo,
		Permission: permission,
	}, nil
}

// 用于外置仓库的锁定，同一时间只有一个线程在同步外置仓库
var lockMap = sync.Map{}

func openRepository(repo *models.Repo) (*gitmodule.Repository, error) {
	repo.RwMutex = &sync.RWMutex{}
	gitRepository, err := gitmodule.OpenRepositoryWithInit(conf.RepoRoot(), repo.Path)
	if err != nil {
		return nil, err
	}
	gitRepository.ID = repo.ID
	gitRepository.ProjectId = repo.ProjectID
	gitRepository.ApplicationId = repo.AppID
	gitRepository.OrgId = repo.OrgID
	gitRepository.Size = repo.Size
	gitRepository.Url = conf.GittarUrl() + "/" + repo.Path
	if repo.IsExternal {
		repoPath := path.Join(conf.RepoRoot(), repo.Path)

		// 判定是否锁定
		lock, ok := lockMap.Load(repoPath)
		if ok {
			if lock != nil && !lock.(bool) {
				// 假如没有锁定，就开始锁定
				lockMap.Store(repoPath, true)
				// 结束后取消锁定
				repo.RwMutex.Lock()
				go func() {
					defer func() {
						lockMap.Store(repoPath, false)
						repo.RwMutex.Unlock()
					}()
					err = gitmodule.SyncExternalRepository(repoPath)
					if err != nil {
						logrus.Errorf(" SyncExternalRepository error: %v ", err)
					}
				}()
			}
		} else {
			lockMap.Store(repoPath, true)
			repo.RwMutex.Lock()
			go func() {
				defer func() {
					lockMap.Store(repoPath, false)
					repo.RwMutex.Unlock()
				}()
				err = gitmodule.SyncExternalRepository(repoPath)
				if err != nil {
					logrus.Errorf(" SyncExternalRepository error: %v ", err)
				}
			}()
		}

	}
	gitRepository.IsExternal = repo.IsExternal
	gitRepository.RwLock = repo.RwMutex

	return gitRepository, nil
}

// getOrgIDV3 get orgID v3
func getOrgIDV3(c *webcontext.Context, orgName string) (int64, error) {
	orgDto, err := c.Bundle.GetOrg(orgName)
	if err == nil {
		return int64(orgDto.ID), nil
	}
	orgIdStr := c.HttpRequest().Header.Get("Org-ID")
	orgID, err := strconv.ParseInt(orgIdStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return orgID, nil
}
