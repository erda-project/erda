// Package org 封装企业资源相关操作
package org

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/conf"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/cmdb/services/nexussvc"
	"github.com/erda-project/erda/modules/cmdb/services/publisher"
	"github.com/erda-project/erda/modules/cmdb/types"
	"github.com/erda-project/erda/modules/cmdb/utils"
	"github.com/erda-project/erda/pkg/numeral"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/uuid"
)

// Org 资源对象操作封装
type Org struct {
	db        *dao.DBClient
	uc        *utils.UCClient
	bdl       *bundle.Bundle
	publisher *publisher.Publisher
	nexusSvc  *nexussvc.NexusSvc
	redisCli  *redis.Client
}

// Option 定义 Org 对象的配置选项
type Option func(*Org)

// New 新建 Org 实例，通过 Org 实例操作企业资源
func New(options ...Option) *Org {
	o := &Org{}
	for _, op := range options {
		op(o)
	}
	return o
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(o *Org) {
		o.db = db
	}
}

// WithUCClient 配置 uc client
func WithUCClient(uc *utils.UCClient) Option {
	return func(o *Org) {
		o.uc = uc
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(o *Org) {
		o.bdl = bdl
	}
}

// WithPublisher 配置 publisher
func WithPublisher(publisher *publisher.Publisher) Option {
	return func(o *Org) {
		o.publisher = publisher
	}
}

// WithNexusSvc 配置 nexus service
func WithNexusSvc(svc *nexussvc.NexusSvc) Option {
	return func(o *Org) {
		o.nexusSvc = svc
	}
}

// WithRedisClient 配置 redis client
func WithRedisClient(cli *redis.Client) Option {
	return func(o *Org) {
		o.redisCli = cli
	}
}

// CreateWithEvent 创建企业 & 发送创建事件
func (o *Org) CreateWithEvent(userID string, createReq apistructs.OrgCreateRequest) (*model.Org, error) {
	// 创建企业
	org, err := o.Create(createReq)
	if err != nil {
		return nil, err
	}

	// 创建企业发布商，这里只根据发布商名字创建
	if createReq.PublisherName != "" {
		pub := &apistructs.PublisherCreateRequest{
			Name:          createReq.PublisherName,
			PublisherType: "ORG",
			OrgID:         uint64(org.ID),
		}

		// 将企业管理员作为 publisher 管理员
		// 若无企业管理员，则将创建企业者作为管理员
		var managerID = userID
		if len(createReq.Admins) > 0 {
			managerID = createReq.Admins[0]
		}

		_, err := o.publisher.Create(managerID, pub)
		if err != nil {
			return org, err
		}
	}

	// 异步保证企业级别 nexus group repo
	go func() {
		if err := o.EnsureNexusOrgGroupRepos(org); err != nil {
			logrus.Errorf("[alert] org nexus: failed to ensure org group repo when create org, orgID: %d, err: %v", org.ID, err)
			return
		}
	}()

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.OrgEvent,
			Action:        bundle.CreateAction,
			OrgID:         strconv.FormatInt(org.ID, 10),
			ProjectID:     "-1",
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCMDB,
		Content: *org,
	}
	// 发送企业创建事件
	if err = o.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send org create event, (%v)", err)
	}
	stageName := []string{
		"设计", "开发", "测试", "实施", "部署", "运维",
		"需求设计", "架构设计", "代码研发",
	}
	stage := []string{
		"design", "dev", "test", "implement", "deploy", "operator",
		"demandDesign", "architectureDesign", "codeDevelopment",
	}
	// stage
	var stages []dao.IssueStage
	for i := 0; i < 9; i++ {
		if i < 6 {
			stages = append(stages, dao.IssueStage{
				OrgID:     org.ID,
				IssueType: apistructs.IssueTypeTask,
				Name:      stageName[i],
				Value:     stage[i],
			})
		} else {
			stages = append(stages, dao.IssueStage{
				OrgID:     org.ID,
				IssueType: apistructs.IssueTypeBug,
				Name:      stageName[i],
				Value:     stage[i],
			})
		}
	}
	err = o.db.CreateIssueStage(stages)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// Create 创建企业处理逻辑
func (o *Org) Create(createReq apistructs.OrgCreateRequest) (*model.Org, error) {
	if len(createReq.Admins) == 0 {
		return nil, errors.Errorf("failed to create org(param admins is empty)")
	}
	if createReq.DisplayName == "" {
		createReq.DisplayName = createReq.Name
	}

	// name 校验
	if err := strutil.Validate(createReq.Name,
		strutil.NoChineseValidator,
		strutil.MaxLenValidator(50),
		strutil.AlphaNumericDashUnderscoreValidator,
	); err != nil {
		return nil, err
	}

	// 校验企业名唯一
	existOrg, err := o.db.GetOrgByName(createReq.Name)
	if err != nil && err != dao.ErrNotFoundOrg {
		return nil, err
	}
	if existOrg != nil {
		return nil, errors.Errorf("org name already exist")
	}

	// 添加企业至DB
	userID := createReq.Admins[0]
	org := &model.Org{
		Name:        createReq.Name,
		DisplayName: createReq.DisplayName,
		Desc:        createReq.Desc,
		Logo:        createReq.Logo,
		Locale:      createReq.Locale,
		UserID:      userID,
		Type:        "ENTERPRISE",
		Status:      "ACTIVE",
		IsPublic:    false,
	}
	if err := o.db.CreateOrg(org); err != nil {
		logrus.Warnf("failed to insert info to db, (%v)", err)
		return nil, errors.Errorf("failed to insert org info to db")
	}

	// 新增企业权限记录
	users, err := o.uc.FindUsers([]string{userID})
	if err != nil {
		logrus.Warnf("failed to query user info, (%v)", err)
	}
	if len(users) > 0 {
		member := model.Member{
			ScopeType:  apistructs.OrgScope,
			ScopeID:    org.ID,
			ScopeName:  org.Name,
			UserID:     userID,
			Email:      users[0].Email,
			Mobile:     users[0].Phone,
			Name:       users[0].Name,
			Nick:       users[0].Nick,
			Avatar:     users[0].AvatarURL,
			UserSyncAt: time.Now(),
			OrgID:      org.ID,
			Token:      uuid.UUID(),
		}
		if err = o.db.CreateMember(&member); err != nil {
			logrus.Warnf("failed to insert member info to db, (%v)", err)
		}
		if err = o.db.CreateMemberExtra(&model.MemberExtra{
			UserID:        userID,
			ScopeType:     apistructs.OrgScope,
			ScopeID:       org.ID,
			ResourceKey:   apistructs.RoleResourceKey,
			ResourceValue: types.RoleOrgManager,
		}); err != nil {
			logrus.Warnf("failed to insert member extra to db, (%v)", err)
		}
	}

	return org, nil
}

// UpdateWithEvent 更新企业 & 发送更新事件
func (o *Org) UpdateWithEvent(orgID int64, updateReq apistructs.OrgUpdateRequestBody) (*model.Org, error) {
	if updateReq.DisplayName == "" {
		updateReq.DisplayName = updateReq.Name
	}
	org, err := o.Update(orgID, updateReq)
	if err != nil {
		return nil, err
	}

	// 异步保证企业级别 nexus group repo
	go func() {
		if err := o.EnsureNexusOrgGroupRepos(org); err != nil {
			logrus.Errorf("[alert] org nexus: failed to ensure org group repo when update org, orgID: %d, err: %v", org.ID, err)
			return
		}
		// TODO 写入 etcd 记录
	}()

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.OrgEvent,
			Action:        bundle.UpdateAction,
			OrgID:         strconv.FormatInt(org.ID, 10),
			ProjectID:     "-1",
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCMDB,
		Content: *org,
	}
	// 发送企业更新事件
	if err = o.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send org update event, (%v)", err)
	}

	return org, nil
}

// Update 更新企业
func (o *Org) Update(orgID int64, updateReq apistructs.OrgUpdateRequestBody) (*model.Org, error) {
	// 检查待更新的org是否存在
	org, err := o.db.GetOrg(orgID)
	if err != nil {
		logrus.Warnf("failed to find org when update org, (%v)", err)
		return nil, errors.Errorf("failed to find org when update org")
	}

	// 更新企业元信息，企业名称暂不可更改
	org.Desc = updateReq.Desc
	org.Logo = updateReq.Logo
	org.DisplayName = updateReq.DisplayName
	org.Locale = updateReq.Locale
	org.IsPublic = updateReq.IsPublic
	if updateReq.Config != nil {
		org.Config.SMTPHost = updateReq.Config.SMTPHost
		org.Config.SMTPPort = updateReq.Config.SMTPPort
		org.Config.SMTPUser = updateReq.Config.SMTPUser
		org.Config.SMTPIsSSL = updateReq.Config.SMTPIsSSL
		org.Config.SMSKeyID = updateReq.Config.SMSKeyID
		org.Config.SMSSignName = updateReq.Config.SMSSignName
		if updateReq.Config.SMSKeySecret != "" && updateReq.Config.SMSKeySecret != apistructs.SECRECT_PLACEHOLDER {
			org.Config.SMSKeySecret = updateReq.Config.SMSKeySecret
		}
		if updateReq.Config.SMTPPassword != "" && updateReq.Config.SMTPPassword != apistructs.SECRECT_PLACEHOLDER {
			org.Config.SMTPPassword = updateReq.Config.SMTPPassword
		}
	}
	if updateReq.BlockoutConfig != nil {
		org.BlockoutConfig.BlockDEV = updateReq.BlockoutConfig.BlockDEV
		org.BlockoutConfig.BlockTEST = updateReq.BlockoutConfig.BlockTEST
		org.BlockoutConfig.BlockStage = updateReq.BlockoutConfig.BlockStage
		org.BlockoutConfig.BlockProd = updateReq.BlockoutConfig.BlockProd
	}

	// 更新企业信息至DB
	if err = o.db.UpdateOrg(&org); err != nil {
		logrus.Warnf("failed to update org, (%v)", err)
		return nil, errors.Errorf("failed to update org")
	}

	return &org, nil
}

// Get 获取企业
func (o *Org) Get(orgID int64) (*model.Org, error) {
	// 检查org是否存在
	org, err := o.db.GetOrg(orgID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch org")
	}
	if org.DisplayName == "" {
		org.DisplayName = org.Name
	}
	return &org, nil
}

// Delete 删除企业
func (o *Org) Delete(orgID int64) error {
	// 检查企业下是否有项目
	projects, err := o.db.ListProjectByOrgID(uint64(orgID))
	if err != nil {
		return err
	}
	if len(projects) > 0 {
		return errors.Errorf("project is not empty")
	}

	relations, err := o.db.GetOrgClusterRelationsByOrg(orgID)
	if err != nil {
		return err
	}
	if len(relations) > 0 {
		return errors.Errorf("cluster relation exist")
	}

	// 删除企业下的成员及权限
	if err = o.db.DeleteMembersByScope(apistructs.OrgScope, orgID); err != nil {
		return errors.Errorf("failed to delete members, (%v)", err)
	}
	if err = o.db.DeleteMemberExtraByScope(apistructs.OrgScope, orgID); err != nil {
		return errors.Errorf("failed to delete members extra, (%v)", err)
	}

	return o.db.DeleteOrg(orgID)
}

// GetByName 获取企业
func (o *Org) GetByName(orgName string) (*model.Org, error) {
	// 检查org是否存在
	org, err := o.db.GetOrgByName(orgName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch org")
	}
	return org, nil
}

// SearchByName 按企业名称过滤
func (o *Org) SearchByName(name string, pageNo, pageSize int) (int, []model.Org, error) {
	total, orgs, err := o.db.GetOrgsByParam(name, pageNo, pageSize)
	if err != nil {
		logrus.Warnf("failed to get orgs, (%v)", err)
		return 0, nil, err
	}
	return total, orgs, nil
}

// Search public orgs
func (o *Org) SearchPublicOrgsByName(name string, pageNo, pageSize int) (int, []model.Org, error) {
	total, orgs, err := o.db.GetPublicOrgsByParam(name, pageNo, pageSize)
	if err != nil {
		return 0, nil, err
	}
	return total, orgs, nil
}

// ListByIDsAndName 根据IDs列表 & name 获取企业列表
func (o *Org) ListByIDsAndName(orgIDs []int64, name string, pageNo, pageSize int) (int, []model.Org, error) {
	return o.db.GetOrgsByIDsAndName(orgIDs, name, pageNo, pageSize)
}

// ChangeCurrentOrg 切换用户当前所属企业
func (o *Org) ChangeCurrentOrg(userID string, req *apistructs.OrgChangeRequest) error {
	// 检查用户是否匹配
	if req.UserID != userID {
		return errors.Errorf("user id doesn't match")
	}
	// 检查企业是否存在
	if _, err := o.db.GetOrg(int64(req.OrgID)); err != nil {
		return err
	}

	orgID, err := o.db.GetCurrentOrgByUser(req.UserID)
	if err != nil || orgID == 0 { // 若当前登录用户currentOrg记录不存在
		currentOrg := &model.CurrentOrg{
			UserID: req.UserID,
			OrgID:  int64(req.OrgID),
		}
		return o.db.CreateCurrentOrg(currentOrg)
	}

	return o.db.UpdateCurrentOrg(req.UserID, int64(req.OrgID))
}

// GetCurrentOrgByUser 根据userID获取用户当前关联企业
func (o *Org) GetCurrentOrgByUser(userID string) (int64, error) {
	return o.db.GetCurrentOrgByUser(userID)
}

// List 获取所有企业列表
func (o *Org) List() ([]model.Org, error) {
	return o.db.GetOrgList()
}

// GetOrgByDomain 通过域名获取企业
func (o *Org) GetOrgByDomain(domain string) (*model.Org, error) {
	// TODO: 返回 apistructs.OrgDTO instead of model.Org
	if domain != "" && conf.OrgWiteList[domain] {
		return nil, nil
	}
	suf := strutil.Concat(".", conf.RootDomain())
	domain_and_port := strutil.Split(domain, ":", true)
	domain = domain_and_port[0]
	if !strutil.HasSuffixes(domain, suf) {
		return nil, apierrors.ErrGetOrg.NotFound()
	}
	orgName := strutil.TrimSuffixes(domain, suf)
	// TODO: after 3.9, we check suffix must have "-org"
	if strutil.HasSuffixes(orgName, "-org") {
		orgName = strutil.TrimSuffixes(orgName, "-org")
	}
	org, err := o.GetByName(orgName)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// RelateCluster 关联集群，创建企业集群关联关系
func (o *Org) RelateCluster(userID string, req *apistructs.OrgClusterRelationCreateRequest) error {
	org, err := o.db.GetOrg(int64(req.OrgID))
	if err != nil {
		return err
	}
	if org.ID == 0 {
		return errors.Errorf("org not found")
	}
	if req.OrgName == "" {
		req.OrgName = org.Name
	} else if org.Name != req.OrgName {
		return errors.Errorf("org info doesn't match")
	}
	cluster, err := o.db.GetClusterByName(req.ClusterName)
	if err != nil {
		return err
	}
	if cluster == nil {
		return errors.Errorf("cluster not found")
	}
	// 若企业集群关系已存在，则返回
	relation, err := o.db.GetOrgClusterRelationByOrgAndCluster(org.ID, cluster.ID)
	if err != nil {
		return err
	}
	if relation != nil {
		return nil
	}

	relation = &model.OrgClusterRelation{
		OrgID:       req.OrgID,
		OrgName:     req.OrgName,
		ClusterID:   uint64(cluster.ID),
		ClusterName: req.ClusterName,
		Creator:     userID,
	}
	return o.db.CreateOrgClusterRelation(relation)
}

// FetchOrgResources 获取企业资源情况
func (o *Org) FetchOrgResources(orgID uint64) (*apistructs.OrgResourceInfo, error) {
	relations, err := o.db.GetOrgClusterRelationsByOrg(int64(orgID))
	if err != nil {
		return nil, err
	}

	var (
		totalCpu float64
		totalMem float64
		usedCpu  float64
		usedMem  float64
	)
	for _, v := range relations {
		// 集群维度获取集群资源总量
		clusterRes, err := o.bdl.ResourceInfo(v.ClusterName, true)
		if err != nil {
			return nil, err
		}
		for _, node := range clusterRes.Nodes {
			cpuovercommit := 1.0
			memovercommit := 1.0

			if !node.IgnoreLabels { // dcos & edas 情况下忽略 label 过滤
				if !strutil.Exist(node.Labels, fmt.Sprintf("%s/org-%s", apistructs.PlatformLabelPrefix, v.OrgName)) {
					continue
				}
				if !strutil.Exist(node.Labels, fmt.Sprintf("%s/%s", apistructs.PlatformLabelPrefix, apistructs.StatelessLabel)) &&
					!strutil.Exist(node.Labels, fmt.Sprintf("%s/%s", apistructs.PlatformLabelPrefix, apistructs.StatefulLabel)) {
					continue
				}
				if strutil.Contains(strutil.Concat(node.Labels...), fmt.Sprintf("%s/workspace-dev", apistructs.PlatformLabelPrefix)) {
					if clusterRes.DevCPUOverCommit > cpuovercommit {
						cpuovercommit = clusterRes.DevCPUOverCommit
					}
					if clusterRes.DevMEMOverCommit > memovercommit {
						memovercommit = clusterRes.DevMEMOverCommit
					}
				}
				if strutil.Contains(strutil.Concat(node.Labels...), fmt.Sprintf("%s/workspace-test", apistructs.PlatformLabelPrefix)) {
					if clusterRes.TestCPUOverCommit > cpuovercommit {
						cpuovercommit = clusterRes.TestCPUOverCommit
					}
					if clusterRes.TestMEMOverCommit > memovercommit {
						memovercommit = clusterRes.TestMEMOverCommit
					}
				}
				if strutil.Contains(strutil.Concat(node.Labels...), fmt.Sprintf("%s/workspace-staging", apistructs.PlatformLabelPrefix)) {
					if clusterRes.StagingCPUOverCommit > cpuovercommit {
						cpuovercommit = clusterRes.StagingCPUOverCommit
					}
					if clusterRes.StagingMEMOverCommit > memovercommit {
						memovercommit = clusterRes.StagingMEMOverCommit
					}
				}
				if strutil.Contains(strutil.Concat(node.Labels...), fmt.Sprintf("%s/workspace-prod", apistructs.PlatformLabelPrefix)) {
					if clusterRes.ProdCPUOverCommit > cpuovercommit {
						cpuovercommit = clusterRes.ProdCPUOverCommit
					}
					if clusterRes.ProdMEMOverCommit > memovercommit {
						memovercommit = clusterRes.ProdMEMOverCommit
					}
				}
			}
			totalCpu += node.CPUAllocatable * cpuovercommit
			totalMem += float64(node.MemAllocatable) * memovercommit / apistructs.GB
		}
	}

	// 获取当前企业已分配资源
	projects, err := o.db.ListProjectByOrgID(orgID)
	if err != nil {
		return nil, err
	}
	for _, v := range projects {
		usedCpu += v.CpuQuota
		usedMem += v.MemQuota
	}

	return &apistructs.OrgResourceInfo{
		TotalCpu:     numeral.Round(totalCpu, 2),
		TotalMem:     numeral.Round(totalMem, 2),
		AvailableCpu: numeral.Round(totalCpu-usedCpu, 2),
		AvailableMem: numeral.Round(totalMem-usedMem, 2),
	}, nil
}

// ListOrgRunningTasks 获取指定集群正在运行的job或者deployment列表
func (o *Org) ListOrgRunningTasks(param *apistructs.OrgRunningTasksListRequest,
	orgID int64) (int64, []apistructs.OrgRunningTasks, error) {
	var (
		total       int64
		resultTasks []apistructs.OrgRunningTasks
	)

	if param.Type == "deployment" {
		totalCount, deployments, err := o.db.ListDeploymentsByOrgID(param, uint64(orgID))
		if err != nil {
			return 0, nil, err
		}

		for _, dep := range *deployments {
			taskData := apistructs.OrgRunningTasks{
				OrgID:           dep.OrgID,
				ProjectID:       dep.ProjectID,
				ApplicationID:   dep.ApplicationID,
				PipelineID:      dep.PipelineID,
				TaskID:          dep.TaskID,
				QueueTimeSec:    dep.QueueTimeSec,
				CostTimeSec:     dep.CostTimeSec,
				Status:          dep.Status,
				Env:             dep.Env,
				ClusterName:     dep.ClusterName,
				CreatedAt:       dep.CreatedAt,
				ProjectName:     dep.ProjectName,
				ApplicationName: dep.ApplicationName,
				TaskName:        dep.TaskName,
				RuntimeID:       dep.RuntimeID,
				ReleaseID:       dep.ReleaseID,
				UserID:          dep.UserID,
			}

			resultTasks = append(resultTasks, taskData)
		}

		total = totalCount
	} else if param.Type == "job" {
		totalCount, jobs, err := o.db.ListJobsByOrgID(param, uint64(orgID))
		if err != nil {
			return 0, nil, err
		}

		for _, job := range *jobs {
			taskData := apistructs.OrgRunningTasks{
				OrgID:           job.OrgID,
				ProjectID:       job.ProjectID,
				ApplicationID:   job.ApplicationID,
				PipelineID:      job.PipelineID,
				TaskID:          job.TaskID,
				QueueTimeSec:    job.QueueTimeSec,
				CostTimeSec:     job.CostTimeSec,
				Status:          job.Status,
				Env:             job.Env,
				ClusterName:     job.ClusterName,
				CreatedAt:       job.CreatedAt,
				ProjectName:     job.ProjectName,
				ApplicationName: job.ApplicationName,
				TaskName:        job.TaskName,
				TaskType:        job.TaskType,
				UserID:          job.UserID,
			}

			resultTasks = append(resultTasks, taskData)
		}

		total = totalCount
	}

	return total, resultTasks, nil
}

// DealReceiveTaskEvent 处理接收到 pipeline 的 task 事件
func (o *Org) DealReceiveTaskEvent(req *apistructs.PipelineTaskEvent) (int64, error) {
	// 参数校验
	if err := o.checkReceiveTaskEventParam(req); err != nil {
		// 若订阅的事件参数校验失败，则无需处理
		return 0, nil
	}

	// task如果是dice，则属于deployment
	// 如果是其它，则属于job
	if req.Content.ActionType == "dice" {
		return o.dealReceiveDeploymentEvent(req)
	} else {
		return o.dealReceiveJobEvent(req)
	}

	return 0, nil
}

// ListAllOrgClusterRelation 获取所有企业对应集群关系
func (o *Org) ListAllOrgClusterRelation() ([]model.OrgClusterRelation, error) {
	return o.db.ListAllOrgClusterRelations()
}

func (o *Org) dealReceiveDeploymentEvent(req *apistructs.PipelineTaskEvent) (int64, error) {
	// 判断记录是否已存在，存在只是更新状态
	deployments := o.db.GetDeployment(req.OrgID, req.Content.PipelineTaskID)
	if deployments != nil && len(deployments) > 0 {
		deployment := &deployments[0]
		// delete other running task record
		for i, task := range deployments {
			if i == 0 {
				continue
			}

			o.db.DeleteDeployment(req.OrgID, task.TaskID)
		}

		if deployment.Status == req.Content.Status && deployment.RuntimeID == req.Content.RuntimeID &&
			deployment.ReleaseID == req.Content.ReleaseID {
			return deployment.ID, nil
		} else {
			deployment.Status = req.Content.Status
			deployment.RuntimeID = req.Content.RuntimeID
			deployment.ReleaseID = req.Content.ReleaseID
		}

		return deployment.ID, o.db.UpdateDeploymentStatus(deployment)
	}

	// 正在运行的任务信息入库
	orgID, err := strconv.ParseInt(req.OrgID, 10, 64)
	if err != nil {
		return 0, err
	}

	projectID, err := strconv.ParseInt(req.ProjectID, 10, 64)
	if err != nil {
		return 0, err
	}

	appID, err := strconv.ParseInt(req.ApplicationID, 10, 64)
	if err != nil {
		return 0, err
	}

	task := &model.Deployments{
		OrgID:           uint64(orgID),
		ProjectID:       uint64(projectID),
		ApplicationID:   uint64(appID),
		PipelineID:      req.Content.PipelineID,
		TaskID:          req.Content.PipelineTaskID,
		QueueTimeSec:    req.Content.QueueTimeSec,
		CostTimeSec:     req.Content.CostTimeSec,
		Status:          req.Content.Status,
		Env:             req.Env,
		ClusterName:     req.Content.ClusterName,
		UserID:          req.Content.UserID,
		CreatedAt:       req.Content.CreatedAt,
		ProjectName:     req.Content.ProjectName,
		ApplicationName: req.Content.ApplicationName,
		TaskName:        req.Content.TaskName,
		RuntimeID:       req.Content.RuntimeID,
		ReleaseID:       req.Content.ReleaseID,
	}

	if err := o.db.CreateDeployment(task); err != nil {
		return 0, err
	}

	return task.ID, nil
}

func (o *Org) dealReceiveJobEvent(req *apistructs.PipelineTaskEvent) (int64, error) {
	// 判断记录是否已存在，存在只是更新状态
	jobs := o.db.GetJob(req.OrgID, req.Content.PipelineTaskID)
	if jobs != nil && len(jobs) > 0 {
		job := &jobs[0]
		// delete other running task record
		for i, task := range jobs {
			if i == 0 {
				continue
			}

			o.db.DeleteJob(req.OrgID, task.TaskID)
		}

		if job.Status == req.Content.Status {
			return job.ID, nil
		} else {
			job.Status = req.Content.Status
		}

		return job.ID, o.db.UpdateJobStatus(job)
	}

	// 正在运行的任务信息入库
	orgID, err := strconv.ParseInt(req.OrgID, 10, 64)
	if err != nil {
		return 0, err
	}

	projectID, _ := strconv.ParseInt(req.ProjectID, 10, 64)

	appID, _ := strconv.ParseInt(req.ApplicationID, 10, 64)

	task := &model.Jobs{
		OrgID:           uint64(orgID),
		ProjectID:       uint64(projectID),
		ApplicationID:   uint64(appID),
		PipelineID:      req.Content.PipelineID,
		TaskID:          req.Content.PipelineTaskID,
		QueueTimeSec:    req.Content.QueueTimeSec,
		CostTimeSec:     req.Content.CostTimeSec,
		Status:          req.Content.Status,
		Env:             req.Env,
		ClusterName:     req.Content.ClusterName,
		UserID:          req.Content.UserID,
		CreatedAt:       req.Content.CreatedAt,
		ProjectName:     req.Content.ProjectName,
		ApplicationName: req.Content.ApplicationName,
		TaskName:        req.Content.TaskName,
	}

	if err := o.db.CreateJob(task); err != nil {
		return 0, err
	}

	return task.ID, nil
}

func (o *Org) checkReceiveTaskEventParam(req *apistructs.PipelineTaskEvent) error {
	if req.OrgID == "" {
		return errors.Errorf("OrgID is empty")
	}

	if req.Content.Status == "" {
		return errors.Errorf("status is empty")
	}

	if req.Content.ActionType == "" {
		return errors.Errorf("actionType is empty")
	}

	if req.Content.PipelineTaskID == 0 {
		return errors.Errorf("pipelineTaskID is empty")
	}

	return nil
}

// DealReceiveTaskRuntimeEvent 处理接收到 pipeline 的 runtimeID 事件
func (o *Org) DealReceiveTaskRuntimeEvent(req *apistructs.PipelineTaskEvent) (int64, error) {
	// 参数校验
	if err := o.checkReceiveTaskRuntimeEventParam(req); err != nil {
		// 若订阅的事件参数校验失败，则无需处理
		return 0, nil
	}

	// 更新状态 runtimeID, releaseID
	jobs := o.db.GetDeployment(req.OrgID, req.Content.PipelineTaskID)
	if jobs != nil && len(jobs) > 0 {
		job := &jobs[0]
		// delete other running job record
		for i, task := range jobs {
			if i == 0 {
				continue
			}

			o.db.DeleteDeployment(req.OrgID, task.TaskID)
		}

		if job.RuntimeID == req.Content.RuntimeID &&
			job.ReleaseID == req.Content.ReleaseID {
			return job.ID, nil
		} else {
			job.RuntimeID = req.Content.RuntimeID
			job.ReleaseID = req.Content.ReleaseID
		}

		return job.ID, o.db.UpdateDeploymentStatus(job)
	}

	return 0, nil
}

func (o *Org) checkReceiveTaskRuntimeEventParam(req *apistructs.PipelineTaskEvent) error {
	if req.OrgID == "" {
		return errors.Errorf("OrgID is empty")
	}

	if req.Content.RuntimeID == "" {
		return errors.Errorf("RuntimeID is empty")
	}

	if req.Content.PipelineTaskID == 0 {
		return errors.Errorf("pipelineTaskID is empty")
	}

	return nil
}

func (o *Org) GetPublisherID(orgID int64) int64 {
	pub, err := o.db.GetPublisherByOrgID(orgID)
	if err != nil && err != dao.ErrNotFoundPublisher {
		logrus.Warning(err)
		return 0
	}
	if pub == nil {
		return 0
	}
	return pub.ID
}

// SetReleaseCrossCluster 设置企业是否允许跨集群部署开关
func (o *Org) SetReleaseCrossCluster(orgID uint64, enable bool) error {
	// 检查待更新的org是否存在
	org, err := o.db.GetOrg(int64(orgID))
	if err != nil {
		return apierrors.ErrSetReleaseCrossCluster.InvalidParameter(err)
	}
	org.Config.EnableReleaseCrossCluster = enable
	return o.db.DB.Model(&model.Org{}).Update(org).Error
}

// GenVerifiCode 生成邀请成员加入企业的验证码
func (o *Org) GenVerifiCode(identityInfo apistructs.IdentityInfo, orgID uint64) (string, error) {
	now := time.Now()
	key := apistructs.OrgInviteCodeRedisKey.GetKey(now.Day(), identityInfo.UserID, strconv.FormatUint(orgID, 10))
	code, err := o.redisCli.Get(key).Result()
	if err == redis.Nil {
		newCode := strutil.RandStr(5) + apistructs.CodeUserID(identityInfo.UserID)
		tommory := now.AddDate(0, 0, 1).Format("2006-01-02") + " 01:00:00"
		tommoryTime, _ := time.ParseInLocation("2006-01-02 15:04:05", tommory, time.Local)
		_, err := o.redisCli.Set(key, newCode, tommoryTime.Sub(now)).Result()
		if err != nil {
			return "", err
		}

		return newCode, nil
	} else if err != nil {
		return "", err
	}

	return code, nil
}

// SetNotifyConfig 设置通知配置
func (o *Org) SetNotifyConfig(orgID int64, notifyConfig apistructs.NotifyConfigUpdateRequestBody) error {
	org, err := o.db.GetOrg(orgID)
	if err != nil {
		return err
	}

	// 目前只开放配置语音短信通知的开关，之后会开发语音短信通知的其他配置
	org.Config.EnableMS = notifyConfig.Config.EnableMS

	return o.db.UpdateOrg(&org)
}

// GetNotifyConfig 获取通知配置
func (o *Org) GetNotifyConfig(orgID int64) (*apistructs.OrgConfig, error) {
	org, err := o.db.GetOrg(orgID)
	if err != nil {
		return nil, err
	}

	return &apistructs.OrgConfig{
		EnableMS: org.Config.EnableMS,
		// SMTPHost:                   org.Config.SMTPHost,
		// SMTPUser:                   org.Config.SMTPUser,
		// SMTPPassword:               org.Config.SMTPPassword,
		// SMTPPort:                   org.Config.SMTPPort,
		// SMTPIsSSL:                  org.Config.SMTPIsSSL,
		// SMSKeyID:                   org.Config.SMSKeyID,
		// SMSKeySecret:               org.Config.SMSKeySecret,
		// SMSSignName:                org.Config.SMSSignName,
		// SMSMonitorTemplateCode:     org.Config.SMSMonitorTemplateCode, // 监控单独的短信模版
		// VMSKeyID:                   org.Config.VMSKeyID,
		// VMSKeySecret:               org.Config.VMSKeySecret,
		// VMSMonitorTtsCode:          org.Config.VMSMonitorTtsCode, // 监控单独的语音模版
		// VMSMonitorCalledShowNumber: org.Config.VMSMonitorCalledShowNumber,
	}, nil
}
