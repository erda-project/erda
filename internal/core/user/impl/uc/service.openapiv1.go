package uc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	perrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apierrors"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/internal/core/user/util"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

var ucLoginMethodI18nMap = map[string]map[string]string{
	"username": {"en-US": "username", "zh-CN": "账密登录", "marks": "external"},
	"sso":      {"en-US": "sso", "zh-CN": "单点登录", "marks": "internal"},
	"email":    {"en-US": "default", "zh-CN": "默认登录方式", "marks": ""},
	"mobile":   {"en-US": "default", "zh-CN": "默认登录方式", "marks": ""},
}

func (p *provider) UserListLoginMethod(ctx context.Context, req *pb.UserListLoginMethodRequest) (*pb.UserListLoginMethodResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}

	res, err := p.handleListLoginMethod(token)
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
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}

	config, err := p.handleGetPwdSecurityConfig(token)
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
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}
	if req.PwdSecurityConfig == nil {
		return nil, apierrors.ErrUpdatePwdSecurityConfig.InvalidParameter("pwdSecurityConfig is empty")
	}
	config := apistructs.PwdSecurityConfig{
		CaptchaChallengeNumber:   int(req.PwdSecurityConfig.CaptchaChallengeNumber),
		ContinuousPwdErrorNumber: int(req.PwdSecurityConfig.ContinuousPwdErrorNumber),
		MaxPwdErrorNumber:        int(req.PwdSecurityConfig.MaxPwdErrorNumber),
		ResetPassWordPeriod:      int(req.PwdSecurityConfig.ResetPassWordPeriod),
	}

	if err := p.handleUpdatePwdSecurityConfig(&config, token); err != nil {
		return nil, err
	}

	return &pb.PwdSecurityConfigUpdateResponse{}, nil
}

func (p *provider) UserBatchFreeze(ctx context.Context, req *pb.UserBatchFreezeRequest) (*pb.UserBatchFreezeResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		if err := p.handleFreezeUser(id, operatorID, token); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchFreezeResponse{}, nil
}

func (p *provider) UserBatchUnfreeze(ctx context.Context, req *pb.UserBatchUnFreezeRequest) (*pb.UserBatchUnFreezeResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		if err := p.handleUnfreezeUser(id, operatorID, token); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchUnFreezeResponse{}, nil
}

func (p *provider) UserBatchUpdateLoginMethod(ctx context.Context, req *pb.UserBatchUpdateLoginMethodRequest) (*pb.UserBatchUpdateLoginMethodResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		updateReq := apistructs.UserUpdateLoginMethodRequest{
			ID:     id,
			Source: req.Source,
		}
		if err := p.handleUpdateLoginMethod(updateReq, operatorID, token); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchUpdateLoginMethodResponse{}, nil
}

func (p *provider) UserCreate(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserCreateResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}
	operatorID := apis.GetUserID(ctx)
	if err := p.handleCreateUsers(req, operatorID, token); err != nil {
		return nil, err
	}
	return &pb.UserCreateResponse{}, nil
}

func (p *provider) UserExport(ctx context.Context, req *pb.UserPagingRequest) (*emptypb.Empty, error) {
	rw, ok := ctx.Value(httpserver.ResponseWriter).(http.ResponseWriter)
	if !ok || rw == nil {
		return nil, perrors.New("response writer missing in context")
	}

	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}

	pagingReq := convertPagingReq(req)
	if pagingReq.PageSize == 0 {
		pagingReq.PageSize = 1024
	}
	if pagingReq.PageNo == 0 {
		pagingReq.PageNo = 1
	}

	var users []apistructs.UserInfoExt
	for i := 0; i < 100; i++ {
		data, err := HandlePagingUsers(pagingReq, token)
		if err != nil {
			return nil, err
		}
		pageData := util.ConvertToUserInfoExt(data)
		users = append(users, pageData.List...)
		if len(pageData.List) < pagingReq.PageSize {
			break
		}
		pagingReq.PageNo++
	}

	locale := apis.GetLang(ctx)
	if locale == "" {
		locale = "zh-CN"
	}
	loginMethodMap, err := p.getLoginMethodMap(token, locale)
	if err != nil {
		return nil, err
	}

	reader, name, err := exportExcel(users, loginMethodMap)
	if err != nil {
		return nil, err
	}

	rw.Header().Add("Content-Disposition", "attachment;fileName="+name+".xlsx")
	rw.Header().Add("Content-Type", "application/vnd.ms-excel")
	if _, err = io.Copy(rw, reader); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (p *provider) UserFreeze(ctx context.Context, req *pb.UserFreezeRequest) (*pb.UserFreezeResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}
	if err := p.handleFreezeUser(req.UserID, apis.GetUserID(ctx), token); err != nil {
		return nil, err
	}
	return &pb.UserFreezeResponse{}, nil
}

func (p *provider) UserUnfreeze(ctx context.Context, req *pb.UserUnfreezeRequest) (*pb.UserUnfreezeResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}
	if err := p.handleUnfreezeUser(req.UserID, apis.GetUserID(ctx), token); err != nil {
		return nil, err
	}
	return &pb.UserUnfreezeResponse{}, nil
}

func (p *provider) UserUpdateLoginMethod(ctx context.Context, req *pb.UserUpdateLoginMethodRequest) (*pb.UserUpdateLoginMethodResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
	if err != nil {
		return nil, err
	}
	id := req.ID
	if id == "" {
		id = req.UserID
	}
	updateReq := apistructs.UserUpdateLoginMethodRequest{
		ID:     id,
		Source: req.Source,
	}
	if err := p.handleUpdateLoginMethod(updateReq, apis.GetUserID(ctx), token); err != nil {
		return nil, err
	}
	return &pb.UserUpdateLoginMethodResponse{}, nil
}

func (p *provider) UserUpdateUserinfo(ctx context.Context, req *pb.UserUpdateInfoRequset) (*pb.UserUpdateInfoResponse, error) {
	token, err := p.OAuthTokenProvider.ExchangeClientCredentials(context.Background(), false, nil)
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

func (p *provider) handleListLoginMethod(token *domain.OAuthToken) (*ListLoginTypeResult, error) {
	var resp Response[*ListLoginTypeResult]
	r, err := httpclient.New().Get(discover.UC()).
		Path("/api/home/admin/login/style").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Do().JSON(&resp)
	if err != nil {
		logrus.Errorf("failed to invoke list user login method, (%v)", err)
		return nil, apierrors.ErrListLoginMethod.InternalError(err)
	}
	if !r.IsOK() {
		logrus.Debugf("failed to list user login method, statusCode: %d, %v", r.StatusCode(), string(r.Body()))
		return nil, apierrors.ErrListLoginMethod.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		logrus.Debugf("failed to list user login method: %+v", resp.Error)
		return nil, apierrors.ErrListLoginMethod.InternalError(perrors.New(resp.Error))
	}
	return resp.Result, nil
}

func (p *provider) handleGetPwdSecurityConfig(token *domain.OAuthToken) (*apistructs.PwdSecurityConfig, error) {
	var resp Response[*apistructs.PwdSecurityConfig]
	r, err := httpclient.New().Get(discover.UC()).
		Path("/api/user/admin/pwd-security-config").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Do().JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrGetPwdSecurityConfig.InternalError(err)
	}
	if !r.IsOK() {
		return nil, apierrors.ErrGetPwdSecurityConfig.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		return nil, apierrors.ErrGetPwdSecurityConfig.InternalError(perrors.New(resp.Error))
	}
	return resp.Result, nil
}

func (p *provider) handleUpdatePwdSecurityConfig(config *apistructs.PwdSecurityConfig, token *domain.OAuthToken) error {
	var resp Response[bool]
	r, err := httpclient.New().Post(discover.UC()).
		Path("/api/user/admin/pwd-security-config").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		JSONBody(config).
		Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrUpdatePwdSecurityConfig.InternalError(err)
	}
	if !r.IsOK() {
		return apierrors.ErrUpdatePwdSecurityConfig.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		return apierrors.ErrUpdatePwdSecurityConfig.InternalError(perrors.New(resp.Error))
	}
	return nil
}

func (p *provider) handleFreezeUser(userID, operatorID string, token *domain.OAuthToken) error {
	var resp Response[bool]
	request := httpclient.New().Put(discover.UC()).
		Path("/api/user/admin/freeze/"+userID).
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken))
	if operatorID != "" {
		request = request.Header("operatorId", operatorID)
	}
	r, err := request.Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrFreezeUser.InternalError(err)
	}
	if !r.IsOK() {
		return apierrors.ErrFreezeUser.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		return apierrors.ErrFreezeUser.InternalError(perrors.New(resp.Error))
	}
	return nil
}

func (p *provider) handleUnfreezeUser(userID, operatorID string, token *domain.OAuthToken) error {
	var resp Response[bool]
	request := httpclient.New().Put(discover.UC()).
		Path("/api/user/admin/unfreeze/"+userID).
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken))
	if operatorID != "" {
		request = request.Header("operatorId", operatorID)
	}
	r, err := request.Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrUnfreezeUser.InternalError(err)
	}
	if !r.IsOK() {
		return apierrors.ErrUnfreezeUser.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		return apierrors.ErrUnfreezeUser.InternalError(perrors.New(resp.Error))
	}
	return nil
}

func (p *provider) handleUpdateLoginMethod(req apistructs.UserUpdateLoginMethodRequest, operatorID string, token *domain.OAuthToken) error {
	var resp Response[bool]
	request := httpclient.New().Post(discover.UC()).
		Path("/api/user/admin/change-full-info").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken))
	if operatorID != "" {
		request = request.Header("operatorID", operatorID)
	}
	r, err := request.JSONBody(&req).Do().JSON(&resp)
	if err != nil {
		logrus.Errorf("failed to invoke change user login method, (%v)", err)
		return apierrors.ErrUpdateLoginMethod.InternalError(err)
	}
	if !r.IsOK() {
		logrus.Debugf("failed to change user login method, statusCode: %d, %v", r.StatusCode(), string(r.Body()))
		return apierrors.ErrUpdateLoginMethod.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		logrus.Debugf("failed to change user login method: %+v", resp.Error)
		return apierrors.ErrUpdateLoginMethod.InternalError(perrors.New(resp.Error))
	}
	return nil
}

func (p *provider) handleCreateUsers(req *pb.UserCreateRequest, operatorID string, token *domain.OAuthToken) error {
	var resp Response[bool]
	users := make([]createUserItem, len(req.Users))
	for i, u := range req.Users {
		users[i] = convertCreateUserItem(u)
	}
	reqBody := createUser{Users: users}
	request := httpclient.New().Post(discover.UC()).
		Path("/api/user/admin/batch-create-user").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken))
	if operatorID != "" {
		request = request.Header("operatorId", operatorID)
	}
	r, err := request.JSONBody(&reqBody).Do().JSON(&resp)
	if err != nil {
		logrus.Errorf("failed to invoke create user, (%v)", err)
		return apierrors.ErrCreateUser.InternalError(err)
	}
	if !r.IsOK() {
		logrus.Debugf("failed to create user, statusCode: %d, %v", r.StatusCode(), string(r.Body()))
		return apierrors.ErrCreateUser.InternalError(fmt.Errorf("internal status code: %v", r.StatusCode()))
	}
	if !resp.Success {
		logrus.Debugf("failed to create user: %+v", resp.Error)
		return apierrors.ErrCreateUser.InternalError(perrors.New(resp.Error))
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

func (p *provider) handleUpdateUserInfo(req *pb.UserUpdateInfoRequset, operatorID string, token *domain.OAuthToken) error {
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
		return apierrors.ErrUpdateUserInfo.InternalError(perrors.New(resp.Error))
	}

	return nil
}

func getUCUpdateUserInfoReq(req *pb.UserUpdateInfoRequset) (*ucUpdateUserInfoReq, error) {
	if req.UserID == "" {
		return nil, perrors.New("user id is empty")
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

func convertPagingReq(req *pb.UserPagingRequest) *apistructs.UserPagingRequest {
	var locked *int
	if req.Locked {
		val := 1
		locked = &val
	}
	return &apistructs.UserPagingRequest{
		Name:     req.Name,
		Nick:     req.Nick,
		Phone:    req.Phone,
		Email:    req.Email,
		Locked:   locked,
		Source:   req.Source,
		PageNo:   int(req.PageNo),
		PageSize: int(req.PageSize),
	}
}

func (p *provider) getLoginMethodMap(token *domain.OAuthToken, locale string) (map[string]string, error) {
	res, err := p.handleListLoginMethod(token)
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

func exportExcel(users []apistructs.UserInfoExt, loginMethodMap map[string]string) (io.Reader, string, error) {
	var (
		table     [][]string
		tableName = "users"
		buf       = bytes.NewBuffer([]byte{})
	)

	table = convertUserToExcelList(users, loginMethodMap)

	if err := excel.ExportExcel(buf, table, tableName); err != nil {
		return nil, "", err
	}

	return buf, tableName, nil
}

func convertUserToExcelList(users []apistructs.UserInfoExt, loginMethodMap map[string]string) [][]string {
	r := [][]string{{"用户名", "昵称", "邮箱", "手机号", "登录方式", "上次登录时间", "密码过期时间", "状态"}}
	for _, user := range users {
		state := "未冻结"
		if user.Locked {
			state = "冻结"
		}
		r = append(r, []string{user.Name, user.Nick, user.Email, user.Phone, loginMethodMap[user.Source], user.LastLoginAt, user.PwdExpireAt, state})
	}
	return r
}

func (p *provider) UserPaging(ctx context.Context, req *pb.UserPagingRequest) (*pb.UserPagingResponse, error) {
	return nil, errors.New("need impl")
}

func HandlePagingUsers(req *apistructs.UserPagingRequest, token *domain.OAuthToken) (*common.UserPaging, error) {
	v := httpclient.New().Get(discover.UC()).Path("/api/user/admin/paging").
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken))
	if req.Name != "" {
		v.Param("username", req.Name)
	}
	if req.Nick != "" {
		v.Param("nickname", req.Nick)
	}
	if req.Phone != "" {
		v.Param("mobile", req.Phone)
	}
	if req.Email != "" {
		v.Param("email", req.Email)
	}
	if req.Locked != nil {
		v.Param("locked", strconv.Itoa(*req.Locked))
	}
	if req.Source != "" {
		v.Param("source", req.Source)
	}
	if req.PageNo > 0 {
		v.Param("pageNo", strconv.Itoa(req.PageNo))
	}
	if req.PageSize > 0 {
		v.Param("pageSize", strconv.Itoa(req.PageSize))
	}
	// 批量查询用户
	var resp struct {
		Success bool               `json:"success"`
		Result  *common.UserPaging `json:"result"`
		Error   string             `json:"error"`
	}
	r, err := v.Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("internal status code: %v", r.StatusCode())
	}
	if !resp.Success {
		return nil, errors.New(resp.Error)
	}
	return resp.Result, nil
}
