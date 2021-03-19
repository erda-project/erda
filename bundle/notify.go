package bundle

import (
	"fmt"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
	"github.com/sirupsen/logrus"
	"strconv"
)

var (
	RenderType = "memberAvatarGroup"
	roleMap    = map[string]string{
		"Owner": "应用所有者",
		"Lead":  "应用主管",
		"Ops":   "运维",
		"Dev":   "开发工程师",
		"QA":    "测试工程师",
	}
	TypeOperation = map[string][]apistructs.Option{
		"role": {
			{
				Name:  "email",
				Value: "email",
			},
			{
				Name:  "message",
				Value: "mbox",
			},
		},
		"user": {
			{
				Name:  "邮箱",
				Value: "email",
			},
			{
				Name:  "站内信",
				Value: "zhanneixin",
			},
		},
		"dingding": {
			{
				Name:  "DingTalk",
				Value: "dingding",
			},
		},
		"webhook": {
			{
				Name:  "webhook",
				Value: "webhook",
			},
		},
		"external_user": {
			{
				Name:  "邮箱",
				Value: "email",
			},
		},
	}
)

func (b *Bundle) NotifyList(req apistructs.NotifyPageRequest) ([]apistructs.NotifyTableList, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var resp apistructs.NotifyListResponse
	path := fmt.Sprintf("/api/notify/records?scope=%v&scopeId=%v", req.Scope, req.ScopeId)
	httpResp, err := hc.Get(host).Path(path).Header(httputil.UserHeader, req.UserId).Do().JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return nil, toAPIError(httpResp.StatusCode(), resp.Error)
	}
	var list []apistructs.NotifyTableList
	for _, v := range resp.Data.List {
		value := apistructs.Value{
			Type:   v.NotifyTarget[0].Type,
			Values: v.NotifyTarget[0].Values,
		}
		targets := apistructs.TableTarget{
			RoleMap:    roleMap,
			RenderType: "listTargets",
			Value:      []apistructs.Value{value},
		}
		switchText := "开启"
		if v.Enable {
			switchText = "关闭"
		}
		operate := apistructs.Operate{
			RenderType: "tableOperation",
			Operations: map[string]apistructs.Operations{
				"edit": {
					Key:    "edit",
					Text:   "编辑",
					Reload: true,
					Meta: apistructs.Meta{
						Id: v.Id,
					},
				},
				"delete": {
					Key:     "delete",
					Text:    "删除",
					Confirm: "确认删除该条通知？",
					Meta: apistructs.Meta{
						Id: v.Id,
					},
					Reload: true,
				},
				"switch": {
					Key:  "switch",
					Text: switchText,
					Meta: apistructs.Meta{
						Id: v.Id,
					},
					Reload: true,
				},
			},
		}
		creatTime := v.CreatedAt.Format("2006-01-02 15:04:05")
		listMember := apistructs.NotifyTableList{
			Id:        v.Id,
			Name:      v.NotifyName,
			Targets:   targets,
			CreatedAt: creatTime,
			Operate:   operate,
		}
		list = append(list, listMember)
	}
	return list, nil
}

func (b *Bundle) DeleteNotifyRecord(scope, scopeId string, id uint64, userId string) error {
	host, err := b.urls.Monitor()
	if err != nil {
		return err
	}
	hc := b.hc
	var resp apistructs.ProNotifyResponse
	path := fmt.Sprintf("/api/notify/records/%d?scope=%v&scopeId=%v", id, scope, scopeId)
	httpResp, err := hc.Delete(host).Path(path).Header(httputil.UserHeader, userId).Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Error)
	}
	return nil
}

func (b *Bundle) SwitchNotifyRecord(scope, scopeId, userId string, operation *apistructs.SwitchOperationData) error {
	host, err := b.urls.Monitor()
	if err != nil {
		return err
	}
	hc := b.hc
	var resp apistructs.ProNotifyResponse
	path := fmt.Sprintf("/api/notify/%d/switch?scope=%v&scopeId=%v", operation.Id, scope, scopeId)
	httpResp, err := hc.Put(host).Path(path).Header(httputil.UserHeader, userId).Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Error)
	}
	return nil
}

func (b *Bundle) GetNotifyDetail(id uint64) (*apistructs.NotifyDetailResponse, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var detailResp apistructs.NotifyDetailResponse
	path := fmt.Sprintf("/api/notify/%d/detail", id)
	httpResp, err := hc.Get(host).Path(path).Do().JSON(&detailResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !detailResp.Success {
		return nil, toAPIError(httpResp.StatusCode(), detailResp.Error)
	}
	return &detailResp, nil
}

func (b *Bundle) GetFieldData(Id uint64, scope, scopeId, userId, orgId string) (*apistructs.NotifyDetailResponse, []apistructs.Field, error) {
	var detailResp *apistructs.NotifyDetailResponse
	var err error
	fileDataResp := make([]apistructs.Field, 0)
	var fileData apistructs.Field
	fileData = apistructs.Field{
		Key:       "name",
		Label:     "通知名称",
		Component: "input",
		Required:  true,
		ComponentProps: apistructs.ComponentProps{
			MaxLength: 50,
		},
	}
	fileData.Disabled = false
	if Id != 0 {
		fileData.Disabled = true
		detailResp, err = b.GetNotifyDetail(Id)
		if err != nil {
			logrus.Errorf("get notify detail is failed,err is %v", err)
			return nil, nil, err
		}
	}
	fileDataResp = append(fileDataResp, fileData)

	//获取所有的可用模版
	allTemplates, err := b.GetAllTemplates(scope, scopeId, userId)
	if err != nil {
		logrus.Errorf("get all templates is failed err is %v", err)
		return nil, nil, apierrors.ErrInvoke.InternalError(err)
	}
	options := make([]apistructs.Option, 0)
	for k, v := range allTemplates {
		option := apistructs.Option{
			Name:  v,
			Value: k,
		}
		options = append(options, option)
	}
	fileData = apistructs.Field{
		Key:       "items",
		Label:     "触发时机",
		Component: "select",
		Required:  true,
		ComponentProps: apistructs.ComponentProps{
			Mode:        "multiple",
			PlaceHolder: "请选择触发时机",
			Options:     options,
		},
	}
	//fileData.ComponentProps.Options = options
	fileDataResp = append(fileDataResp, fileData)
	fileData = apistructs.Field{
		Key:       "target",
		Label:     "选择群组",
		Component: "select",
		Required:  true,
		ComponentProps: apistructs.ComponentProps{
			PlaceHolder: "请选择群组",
		},
	}
	//获取所有的通知组信息
	allGroups, err := b.GetAllGroups(scope, scopeId, orgId, userId)
	if err != nil {
		logrus.Errorf("get all groups is failed err is %v", err)
		return nil, nil, apierrors.ErrInvoke.InternalError(err)
	}
	groupOptions := make([]apistructs.Option, 0)
	for _, v := range allGroups {
		groupOption := apistructs.Option{
			Name:  v.Name,
			Value: strconv.Itoa(int(v.Value)),
		}
		groupOptions = append(groupOptions, groupOption)
	}
	fileData.ComponentProps.Options = groupOptions
	fileDataResp = append(fileDataResp, fileData)
	//用于判断是否有配置通知组
	var flag bool
	for _, v := range allGroups {
		flag = true
		fileData = apistructs.Field{
			Key:       "channels-" + strconv.Itoa(int(v.Value)),
			Label:     "通知方式",
			Component: "select",
			Required:  true,
			ComponentProps: apistructs.ComponentProps{
				Mode:        "multiple",
				PlaceHolder: "请选择通知方式",
				Options:     TypeOperation[v.Type],
			},
			RemoveWhen: [][]apistructs.RemoveWhen{
				{
					{
						Field:    "target",
						Operator: "!=",
						Value:    strconv.Itoa(int(v.Value)),
					},
				},
			},
		}
		//判断是否需要添加电话和短信
		enableMS, err := b.GetNotifyConfigMS(userId, orgId)
		if err != nil {
			return nil, nil, apierrors.ErrInvoke.InternalError(err)
		}
		if enableMS {
			msOption := []apistructs.Option{
				{
					Name:  "SMS",
					Value: "sms",
				},
				{
					Name:  "phone",
					Value: "vms",
				},
			}
			fileData.ComponentProps.Options = append(fileData.ComponentProps.Options, msOption...)
		}
		fileDataResp = append(fileDataResp, fileData)
	}
	if !flag {
		nullOption := make([]apistructs.Option, 0)
		fileData = apistructs.Field{
			Key:       "channels",
			Label:     "通知方式",
			Component: "select",
			Required:  true,
			ComponentProps: apistructs.ComponentProps{
				Mode:        "multiple",
				PlaceHolder: "请选择通知方式",
				Options:     nullOption,
			},
			RemoveWhen: [][]apistructs.RemoveWhen{
				{
					{
						Field:    "target",
						Operator: "!=",
					},
				},
			},
		}
		fileDataResp = append(fileDataResp, fileData)
	}
	logrus.Errorf("fileDataResp is %+v", fileDataResp)
	return detailResp, fileDataResp, nil
}

func (b *Bundle) GetAllTemplates(scope, scopeId, userId string) (map[string]string, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var resp apistructs.AllTemplatesResponse
	path := fmt.Sprintf("/api/notify/templates?scope=%v&scopeId=%v", scope, scopeId)
	httpResp, err := hc.Get(host).Path(path).Header(httputil.UserHeader, userId).Do().JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return nil, toAPIError(httpResp.StatusCode(), resp.Error)
	}
	templateMap := make(map[string]string)
	for _, v := range resp.Data {
		templateMap[v.ID] = v.Name
	}
	return templateMap, nil
}

func (b *Bundle) GetAllGroups(scope, scopeId, orgId, userId string) ([]apistructs.AllGroups, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var resp apistructs.AllGroupResponse
	path := "/api/notify/all-group"
	httpResp, err := hc.Get(host).Path(path).Param("scope", scope).Param("scopeId", scopeId).
		Header(httputil.OrgHeader, orgId).Header(httputil.UserHeader, userId).Do().JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return nil, toAPIError(httpResp.StatusCode(), resp.Error)
	}
	return resp.Data, nil
}

func (b *Bundle) GetNotifyConfigMS(userId, orgId string) (bool, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return false, err
	}
	hc := b.hc
	var resp apistructs.NotifyConfigGetResponse
	path := fmt.Sprintf("/api/orgs/%v/actions/get-notify-config", orgId)
	httpResp, err := hc.Get(host).Path(path).Header(httputil.UserHeader, userId).Do().JSON(&resp)
	if err != nil {
		return false, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return false, toAPIError(httpResp.StatusCode(), resp.Error)
	}
	return resp.Data.Config.EnableMS, nil
}

func (b *Bundle) CollectNotifyMetrics(metrics *apistructs.Metric) error {
	host, err := b.urls.Collector()
	if err != nil {
		return err
	}
	hc := b.hc
	resp, err := hc.Post(host).Path("/collect/notify-metrics").Header("Content-Type", "application/json").
		JSONBody(&metrics).Do().DiscardBody()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("failed to call monitor status %d", resp.StatusCode()))
	}
	return nil
}
