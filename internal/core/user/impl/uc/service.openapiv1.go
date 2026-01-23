package uc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"

	useroauthpb "github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apierrors"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/pointer"
	"github.com/erda-project/erda/pkg/strutil"
)

var ucLoginMethodI18nMap = map[string]map[string]string{
	"username": {"en-US": "username", "zh-CN": "账密登录", "marks": "external"},
	"sso":      {"en-US": "sso", "zh-CN": "单点登录", "marks": "internal"},
	"email":    {"en-US": "default", "zh-CN": "默认登录方式", "marks": ""},
	"mobile":   {"en-US": "default", "zh-CN": "默认登录方式", "marks": ""},
}

func (p *provider) newAuthedClient(refresh *bool) (*httpclient.HTTPClient, error) {
	oauthToken, err := p.UserOAuthSvc.ExchangeClientCredentials(
		context.Background(), &useroauthpb.ExchangeClientCredentialsRequest{
			Refresh: pointer.BoolDeref(refresh, false),
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to exchange client credentials token")
	}

	return p.client.BearerTokenAuth(oauthToken.AccessToken), nil
}

func (p *provider) UserListLoginMethod(ctx context.Context, req *pb.UserListLoginMethodRequest) (*pb.UserListLoginMethodResponse, error) {
	res, err := p.handleListLoginMethod()
	if err != nil {
		return nil, err
	}

	locale := apis.GetLang(ctx)
	if locale == "" {
		locale = "zh-CN"
	}

	deDup := make(map[string]struct{})
	var methods []*pb.UserLoginMethod
	for _, v := range res.RegistryType {
		tmp := getLoginTypeByUC(v)
		if tmp == nil {
			continue
		}
		if _, ok := deDup[tmp["marks"]]; ok {
			continue
		}
		methods = append(methods, &pb.UserLoginMethod{
			DisplayName: tmp[locale],
			Value:       tmp["marks"],
		})
		deDup[tmp["marks"]] = struct{}{}
	}

	return &pb.UserListLoginMethodResponse{Data: methods}, nil
}

func (p *provider) PwdSecurityConfigGet(ctx context.Context, req *pb.PwdSecurityConfigGetRequest) (*pb.PwdSecurityConfigGetResponse, error) {
	config, err := p.handleGetPwdSecurityConfig()
	if err != nil {
		return nil, err
	}

	return &pb.PwdSecurityConfigGetResponse{
		Data: &pb.PwdSecurityConfig{
			CaptchaChallengeNumber:   int64(config.CaptchaChallengeNumber),
			ContinuousPwdErrorNumber: int64(config.ContinuousPwdErrorNumber),
			MaxPwdErrorNumber:        int64(config.MaxPwdErrorNumber),
			ResetPassWordPeriod:      int64(config.ResetPassWordPeriod),
		},
	}, nil
}

func (p *provider) PwdSecurityConfigUpdate(ctx context.Context, req *pb.PwdSecurityConfigUpdateRequest) (*pb.PwdSecurityConfigUpdateResponse, error) {
	config := apistructs.PwdSecurityConfig{
		CaptchaChallengeNumber:   int(req.CaptchaChallengeNumber),
		ContinuousPwdErrorNumber: int(req.ContinuousPwdErrorNumber),
		MaxPwdErrorNumber:        int(req.MaxPwdErrorNumber),
		ResetPassWordPeriod:      int(req.ResetPassWordPeriod),
	}

	if err := p.handleUpdatePwdSecurityConfig(&config); err != nil {
		return nil, err
	}

	return &pb.PwdSecurityConfigUpdateResponse{}, nil
}

func (p *provider) UserBatchFreeze(ctx context.Context, req *pb.UserBatchFreezeRequest) (*pb.UserBatchFreezeResponse, error) {
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		if err := p.handleFreezeUser(id, operatorID); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchFreezeResponse{}, nil
}

func (p *provider) UserBatchUnfreeze(ctx context.Context, req *pb.UserBatchUnFreezeRequest) (*pb.UserBatchUnFreezeResponse, error) {
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		if err := p.handleUnfreezeUser(id, operatorID); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchUnFreezeResponse{}, nil
}

func (p *provider) UserBatchUpdateLoginMethod(ctx context.Context, req *pb.UserBatchUpdateLoginMethodRequest) (*pb.UserBatchUpdateLoginMethodResponse, error) {
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		updateReq := apistructs.UserUpdateLoginMethodRequest{
			ID:     id,
			Source: req.Source,
		}
		if err := p.handleUpdateLoginMethod(updateReq, operatorID); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchUpdateLoginMethodResponse{}, nil
}

func (p *provider) UserCreate(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserCreateResponse, error) {
	operatorID := apis.GetUserID(ctx)
	if err := p.handleCreateUsers(req, operatorID); err != nil {
		return nil, err
	}
	return &pb.UserCreateResponse{}, nil
}

func (p *provider) UserExport(ctx context.Context, req *pb.UserPagingRequest) (*pb.UserExportResponse, error) {
	var (
		users  []*pb.ManagedUser
		total  int64
		pageNo = req.PageNo
	)

	if pageNo == 0 {
		pageNo = 100
	}

	for {
		data, err := p.UserPaging(ctx, req)
		if err != nil {
			return nil, err
		}

		if total == 0 {
			total = data.Total
		}

		users = append(users, data.List...)
		if int64(len(users)) >= total {
			break
		}
		req.PageNo++
	}

	locale := apis.GetLang(ctx)
	if locale == "" {
		locale = "zh-CN"
	}

	loginMethodMap, err := p.getLoginMethodMap(locale)
	if err != nil {
		return nil, err
	}

	return &pb.UserExportResponse{
		Total:        total,
		List:         users,
		LoginMethods: loginMethodMap,
	}, nil
}

func (p *provider) UserFreeze(ctx context.Context, req *pb.UserFreezeRequest) (*pb.UserFreezeResponse, error) {
	if err := p.handleFreezeUser(req.UserID, apis.GetUserID(ctx)); err != nil {
		return nil, err
	}
	return &pb.UserFreezeResponse{}, nil
}

func (p *provider) UserUnfreeze(ctx context.Context, req *pb.UserUnfreezeRequest) (*pb.UserUnfreezeResponse, error) {
	if err := p.handleUnfreezeUser(req.UserID, apis.GetUserID(ctx)); err != nil {
		return nil, err
	}
	return &pb.UserUnfreezeResponse{}, nil
}

func (p *provider) UserUpdateLoginMethod(ctx context.Context, req *pb.UserUpdateLoginMethodRequest) (*pb.UserUpdateLoginMethodResponse, error) {
	id := req.ID
	if id == "" {
		id = req.UserID
	}
	updateReq := apistructs.UserUpdateLoginMethodRequest{
		ID:     id,
		Source: req.Source,
	}
	if err := p.handleUpdateLoginMethod(updateReq, apis.GetUserID(ctx)); err != nil {
		return nil, err
	}
	return &pb.UserUpdateLoginMethodResponse{}, nil
}

func (p *provider) UserUpdateUserinfo(ctx context.Context, req *pb.UserUpdateInfoRequest) (*pb.UserUpdateInfoResponse, error) {
	token, err := p.UserOAuthSvc.ExchangeClientCredentials(context.Background(), &useroauthpb.ExchangeClientCredentialsRequest{
		Refresh: false,
	})
	if err != nil {
		return nil, err
	}
	if err := p.handleUpdateUserInfo(req, apis.GetUserID(ctx), token); err != nil {
		return nil, err
	}
	return &pb.UserUpdateInfoResponse{}, nil
}

func getLoginTypeByUC(key string) map[string]string {
	if v, ok := ucLoginMethodI18nMap[key]; ok {
		return v
	}
	return nil
}

func (p *provider) handleListLoginMethod() (*ListLoginTypeResult, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/api/home/admin/login/style"
		body bytes.Buffer
	)

	r, err := client.Get(p.Cfg.Host).Path(path).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to freeze user")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[*ListLoginTypeResult]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

func (p *provider) handleGetPwdSecurityConfig() (*apistructs.PwdSecurityConfig, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/api/user/admin/pwd-security-config"
		body bytes.Buffer
	)

	r, err := client.Get(p.Cfg.Host).Path(path).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get password security config")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[*apistructs.PwdSecurityConfig]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

func (p *provider) handleUpdatePwdSecurityConfig(config *apistructs.PwdSecurityConfig) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/pwd-security-config"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		JSONBody(config).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to update pwd security config")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp common.UCResponseMeta
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if resp.Success != nil && !*resp.Success {
		return errors.New("failed to update pwd security config")
	}

	return nil
}

func (p *provider) handleFreezeUser(userID, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/freeze/" + userID
		body bytes.Buffer
	)

	r, err := client.Put(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to freeze user")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Result {
		return errors.New("failed to freeze user")
	}

	return nil
}

func (p *provider) handleUnfreezeUser(userID, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/unfreeze/" + userID
		body bytes.Buffer
	)

	r, err := client.Put(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to unfreeze user")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Result {
		return errors.New("failed to unfreeze user")
	}

	return nil
}

func (p *provider) handleUpdateLoginMethod(req apistructs.UserUpdateLoginMethodRequest, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/api/user/admin/change-full-info"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		JSONBody(&req).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to invoke change user login method")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Result {
		return errors.New("failed to invoke change user login method")
	}

	return nil
}

func (p *provider) handleCreateUsers(req *pb.UserCreateRequest, operatorID string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	users := make([]createUserItem, len(req.Users))
	for i, u := range req.Users {
		users[i] = convertCreateUserItem(u)
	}
	reqBody := createUser{Users: users}

	var (
		path = "/api/user/admin/batch-create-user"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		Param("operatorId", operatorID).
		JSONBody(&reqBody).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to invoke create user")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp Response[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Result {
		return errors.New("failed to invoke create user")
	}
	return nil
}

type createUser struct {
	Users []createUserItem `json:"users"`
}

type createUserItem struct {
	Username    string      `json:"username,omitempty"`
	Nickname    string      `json:"nickname,omitempty"`
	Mobile      string      `json:"mobile,omitempty"`
	Email       string      `json:"email,omitempty"`
	Password    string      `json:"password"`
	Avatar      string      `json:"avatar,omitempty"`
	Channel     string      `json:"channel,omitempty"`
	ChannelType string      `json:"channelType,omitempty"`
	Extra       interface{} `json:"extra,omitempty"`
	Source      string      `json:"source,omitempty"`
	SourceType  string      `json:"sourceType,omitempty"`
	Tag         string      `json:"tag,omitempty"`
	UserDetail  interface{} `json:"userDetail,omitempty"`
}

func convertCreateUserItem(item *pb.UserCreateItem) createUserItem {
	return createUserItem{
		Username: item.Name,
		Nickname: item.Nick,
		Mobile:   item.Phone,
		Email:    item.Email,
		Password: item.Password,
	}
}

type ucUpdateUserInfoReq struct {
	ID       string `json:"id,omitempty"`
	UserName string `json:"username,omitempty"`
	Nick     string `json:"nickname,omitempty"`
	Mobile   string `json:"mobile,omitempty"`
	Email    string `json:"email,omitempty"`
}

func (p *provider) handleUpdateUserInfo(req *pb.UserUpdateInfoRequest, operatorID string, token *useroauthpb.OAuthToken) error {
	body, err := getUCUpdateUserInfoReq(req)
	if err != nil {
		return apierrors.ErrUpdateUserInfo.InternalError(err)
	}

	var resp Response[bool]
	request := httpclient.New().Post(discover.UC()).
		Path("/api/user/admin/change-full-info").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken))
	if operatorID != "" {
		request = request.Header("operatorId", operatorID)
	}
	r, err := request.JSONBody(body).Do().JSON(&resp)
	if err != nil {
		logrus.Errorf("failed to invoke update userinfo, (%v)", err)
		return apierrors.ErrUpdateUserInfo.InternalError(err)
	}
	if !r.IsOK() {
		logrus.Debugf("failed to update userinfo, statusCode: %d, %v", r.StatusCode(), string(r.Body()))
		return apierrors.ErrUpdateUserInfo.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		logrus.Debugf("failed to update userinfo: %+v", resp.Error)
		return apierrors.ErrUpdateUserInfo.InternalError(errors.New(resp.Error))
	}

	return nil
}

func getUCUpdateUserInfoReq(req *pb.UserUpdateInfoRequest) (*ucUpdateUserInfoReq, error) {
	if req.UserID == "" {
		return nil, errors.New("user id is empty")
	}
	ucReq := &ucUpdateUserInfoReq{
		ID: req.UserID,
	}
	if req.Nick != "" {
		ucReq.Nick = req.Nick
	}
	if req.Name != "" {
		ucReq.UserName = req.Name
	}
	if req.Mobile != "" {
		ucReq.Mobile = req.Mobile
	}
	if req.Email != "" {
		ucReq.Email = req.Email
	}
	return ucReq, nil
}

func (p *provider) getLoginMethodMap(locale string) (map[string]string, error) {
	res, err := p.handleListLoginMethod()
	if err != nil {
		return nil, err
	}

	valueDisplayNameMap := make(map[string]string)
	deDupMap := make(map[string]struct{})
	for _, v := range res.RegistryType {
		tmp := getLoginTypeByUC(v)
		if tmp == nil {
			continue
		}
		if _, ok := deDupMap[tmp["marks"]]; ok {
			continue
		}
		valueDisplayNameMap[tmp["marks"]] = tmp[locale]
		deDupMap[tmp["marks"]] = struct{}{}
	}

	return valueDisplayNameMap, nil
}

func (p *provider) UserPaging(ctx context.Context, req *pb.UserPagingRequest) (*pb.UserPagingResponse, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	conditions := url.Values{}
	if req.Name != "" {
		conditions.Add("username", req.Name)
	}
	if req.Nick != "" {
		conditions.Add("nickname", req.Nick)
	}
	if req.Phone != "" {
		conditions.Add("mobile", req.Phone)
	}
	if req.Email != "" {
		conditions.Add("email", req.Email)
	}
	if req.Locked {
		conditions.Add("locked", strconv.Itoa(1))
	}
	if req.Source != "" {
		conditions.Add("source", req.Source)
	}
	if req.PageNo > 0 {
		conditions.Add("pageNo", cast.ToString(req.PageNo))
	}
	if req.PageSize > 0 {
		conditions.Add("pageSize", cast.ToString(req.PageSize))
	}

	var (
		path = "/api/user/admin/paging"
		body bytes.Buffer
	)

	r, err := client.Get(p.Cfg.Host).Path(path).
		Params(conditions).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to paging user")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var users common.UCResponse[*common.UserPaging]
	if err := json.NewDecoder(&body).Decode(&users); err != nil {
		return nil, err
	}

	userList := make([]*pb.ManagedUser, 0, len(users.Result.Data))
	for _, datum := range users.Result.Data {
		pbUser, err := managedUserMapper(datum)
		if err != nil {
			return nil, err
		}
		userList = append(userList, pbUser)
	}

	return &pb.UserPagingResponse{
		Total: int64(users.Result.Total),
		List:  userList,
	}, nil
}

func managedUserMapper(u common.UserInPaging) (*pb.ManagedUser, error) {
	return &pb.ManagedUser{
		Id:          cast.ToString(u.Id),
		Name:        u.Username,
		Nick:        u.Nickname,
		Avatar:      u.Avatar,
		Phone:       u.Mobile,
		Email:       u.Email,
		LastLoginAt: timestamppb.New(time.Time(u.LastLoginAt)),
		PwdExpireAt: timestamppb.New(time.Time(u.PwdExpireAt)),
		Source:      u.Source,
		Locked:      u.Locked,
	}, nil
}
