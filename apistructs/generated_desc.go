package apistructs

var structDescMap = map[string]map[string]string{"APIAssetVersionInstanceCreateRequest": {"EndpointID": "关联 API Gateway EndpointID\n", "InstanceType": "实例类型，必填\n", "RuntimeID": "关联一个 Runtime Service\n", "URL": "URL 为用户直接输入\n"}, "APIOperation": {"Description": "Optional description. Should use CommonMark syntax.\n", "Headers": "Optional Headers\n", "OperationID": "Optional operation ID.\n", "Parameters": "Optional parameters.\n", "RequestBodyDescription": "Optional body parameter.\n", "Responses": "Responses.\n其中 key 为 http status code\n"}, "ActionCache": {"Key": "缓存生成的 key 或者是用户指定的 key\n用户指定的话 需要 {{basePath}}/路径/{{endPath}} 来自定义 key\n用户没有指定 key 有一定的生成规则, 具体生成规则看 prepare.go 的 setActionCacheStorageAndBinds 方法\n"}, "ActionCallback": {"MachineStat": "machine stat\n", "Metadata": "show in stdout\n", "PipelineID": "behind\n"}, "ActionCreateRequest": {"ImageSrc": "源action镜像地址\n", "Name": "action名\n", "SpecSrc": "spec yml 内容\n", "Version": "action版本\n"}, "ActivitiyListRequest": {"PageNo": "default 1\n", "PageSize": "default 20\n"}, "AddNodesRequest": {"DataDiskDevice": "optional\n", "Source": "optional\n"}, "AddonAvailableRequest": {"ProjectID": "项目Id\n", "Workspace": "环境, 可选值: DEV/TEST/STAGING/PROD\n"}, "AddonConfigRes": {"AttachCount": "addon被引用数\n", "Config": "环境变量信息\n", "CreateAt": "创建时间\n", "DocInfo": "文档信息\n", "Engine": "addon名称\n", "Label": "Label label信息\n", "LogoURL": "logo图片\n", "Name": "addon实例名称\n", "ReferenceInfo": "addon被引用信息\n", "Status": "addon状态\n", "Type": "addon类型\n", "UpdateAt": "更新时间\n"}, "AddonConfigUpdateRequest": {"Config": "更新配置信息，覆盖更新\n"}, "AddonCreateItem": {"Actions": "action\n", "Configs": "环境变量配置\n", "Name": "addon实例名称\n", "Options": "额外恶心\n", "Plan": "addon规格\n", "Type": "addon名称\n"}, "AddonCreateOptions": {"ApplicationID": "应用ID\n", "ApplicationName": "应用名称\n", "ClusterName": "集群名称\n", "DeploymentID": "发布ID\n", "Env": "所属环境\n", "LogSource": "日志类型\n", "OrgID": "企业ID\n", "OrgName": "企业名称\n", "ProjectID": "项目ID\n", "ProjectName": "项目名称\n", "RuntimeID": "分支名称\n", "RuntimeName": "分支名称\n", "Workspace": "所属环境\n"}, "AddonCreateRequest": {"ApplicationID": "应用ID\n", "ClusterName": "集群\n", "Operator": "操作人\n", "Options": "补充信息\n", "OrgID": "企业ID\n", "ProjectID": "项目ID\n", "RuntimeID": "runtimeId\n", "RuntimeName": "分支名称\n", "Workspace": "所属环境\n"}, "AddonDependsRelation": {"AddonName": "AddonName addon名称\n", "ChildDepends": "ChildDepends 子依赖\n", "InstanceName": "InstanceName 实例名称\n", "ParentDepends": "ParentDepends 父依赖\n", "Plan": "Plan addon规格\n", "Version": "Version addon版本\n"}, "AddonDirectCreateRequest": {"ApplicationID": "应用ID\n", "ClusterName": "集群\n", "Operator": "操作人\n", "OrgID": "企业ID\n", "ProjectID": "项目ID\n", "ShareScope": "CLUSTER | PROJECT\n", "Workspace": "所属环境\n"}, "AddonExtension": {"ConfigVars": "ConfigVars 返回内容配置约定，根据不同服务属性来返回对应的内容\n", "Domain": "Domain addon 服务地址 (仅针对服务部署类型,默认该服务为addon详情介绍页)\n", "Envs": "Envs 添加非第三方addon需要的环境变量\n", "Plan": "Plan addon 支持规格 (仅针对服务部署类型)，根据能力自身的标准来制定，规格名称可以自行指定，比如basic(基础版)、professional(专业版)、ultimate(旗舰版)\n", "Requires": "Requires addon 配置要求，目前支持以下三种属性，某项配置不允许，不传即可\n", "ShareScopes": "ShareScopes 共享级别，PROJECT、CLUSTER、DICE(未来会下掉)\n", "Similar": "Similar 同类addon，如mysql对应rds\n", "SubCategory": "主分类信息\n", "Version": "Version 版本信息\n"}, "AddonFetchResponseData": {"AddonDisplayName": "AddonDisplayName addon 显示名称\n", "AddonName": "AddonName addon 名称，eg: mysql, kafka\n", "AttachCount": "AttachCount 引用数量\n", "CanDel": "CanDel 是否可删除\n", "Category": "Category addon 类别: 微服务/数据库/配置中心，etc\n", "Cluster": "Cluster 集群名称\n", "Config": "Config addon 使用配置, eg: 地址/端口/账号\n", "ConsoleUrl": "ConsoleUrl addon跳转界面\n", "CreatedAt": "CreatedAt 创建时间\n", "CustomAddonType": "CustomAddonType cloud addon信息\n", "Desc": "Desc addon desc\n", "LogoURL": "LogoURL addon logo\n", "Name": "addon实例名称\n", "OrgID": "OrgID 企业 id\n", "Plan": "Plan addon 规格, basic/professional, etc\n", "Platform": "Platform 是否为微服务\n", "PlatformServiceType": "PlatformServiceType 平台服务类型，0：非平台服务，1：微服务，2：平台组件\n", "ProjectID": "ProjectID 项目 id\n", "ProjectName": "ProjectName 项目名称\n", "RealInstanceID": "RealInstanceID addon 真实实例Id\n", "RecordID": "RecordID cloud addon信息\n", "Reference": "Reference addon 引用计数\n", "ShareScope": "ShareScope 共享级别, eg: 项目共享/企业共享/集群共享/dice共享\n", "Status": "Status addon 状态\n", "Tag": "addon标签\n", "TenantOwner": "TenantOwner addon 租户owner的 instancerouting id\n", "TerminusKey": "Terminus Key 监控 addon 跳转使用\n", "UpdatedAt": "UpdatedAt 更新时间\n", "Version": "Version addon 版本\n", "Workspace": "Workspace， DEV/TEST/STAGING/PROD\n"}, "AddonHandlerCreateItem": {"AddonName": "AddonName addon名称\n", "ApplicationID": "ApplicationID应用ID\n", "ClusterName": "ClusterName 集群名称\n", "Config": "Config 环境变量配置\n", "InsideAddon": "InsideAddon 是否为内部依赖addon，N:否，Y:是\n", "InstanceName": "InstanceName addon实例名称\n", "OperatorID": "OperatorID 用户ID\n", "Options": "Options 额外信息配置\n", "OrgID": "OrgID 企业ID\n", "Plan": "Plan addon规格\n", "ProjectID": "ProjectID 项目ID\n", "RuntimeID": "RuntimeID runtimeID\n", "RuntimeName": "RuntimeName runtime名称\n", "ShareScope": "ShareScope 是否为内部依赖addon，N:否，Y:是\n", "Tag": "Tag 标签\n", "Workspace": "Workspace 所属环境\n"}, "AddonNameResponse": {"Data": "key 为 projectID\n"}, "AddonPlanItem": {"CPU": "CPU cpu大小\n", "InsideMoudle": "内部组件依赖信息，如果有，则用内部组件的信息\n", "Mem": "Mem 内存大小\n", "Nodes": "Nodes 节点数量\n", "Offerings": "Offerings 规格特征说明\n"}, "AddonPlanRes": {"Plan": "规格信息\n", "PlanCnName": "规格信息中文说明\n"}, "AddonProviderDataResp": {"Config": "Config 配置信息\n", "CreateAt": "CreateAt 创建时间\n", "Label": "Label 配置信息\n", "UUID": "UUID 唯一ID\n", "UpdateAt": "UpdateAt 更新时间\n"}, "AddonProviderRequest": {"Callback": "Callback 回调地址\n", "ClusterName": "ClusterName 集群名称\n", "Name": "Name 名称\n", "Options": "Options 额外信息\n", "Plan": "Plan 规格\n", "UUID": "Uuid 唯一ID\n"}, "AddonReferenceRes": {"ApplicationID": "应用ID\n", "AttachTime": "引用时间\n", "Name": "引用组成名称\n", "OrgID": "企业Id\n", "ProjectID": "项目ID\n", "RuntimeID": "runtimeID\n"}, "AddonRes": {"CategoryName": "addon分类\n", "Description": "addon描述信息\n", "DisplayName": "addon展示名称\n", "Envs": "ENVS信息\n", "ID": "addonId\n", "IconURL": "icon图片\n", "InstanceID": "addon实例Id\n", "InstanceName": "addon实例名称\n", "LogoURL": "logo图片\n", "Name": "addon名称\n", "NeedCreate": "是否需要创建\n", "Plan": "规格\n", "Plans": "规格信息列表\n", "ShareScope": "addon共享级别\n", "SubCategory": "addon所属类别\n", "Vars": "VARS信息\n", "Versions": "版本信息\n"}, "AddonScaleRequest": {"CPU": "CPU cpu大小\n", "Mem": "Mem 内存大小\n", "Nodes": "Nodes 节点数量\n"}, "AddonStrategy": {"CanRegister": "CanRegister 是否要注册，1：是，0：不是\n", "DiffEnv": "DiffEnv 是否区分环境，true：区分\n", "FrontDisplay": "FrontDisplay 是否前端展示。true：展示\n", "IsPlatform": "IsPlatform 是否微服务\n", "MenuDisplay": "MenuDisplay 是否展示菜单，true：展示\n", "SupportClusterType": "SupportClusterType 支持发布的集群(如：k8s,dcos,edas)\n"}, "AddonTenantCreateRequest": {"Configs": "对于 Mysql\ndatabases: db1,db2,...   ; 该tenant用户有权限操作的db, db若不存在则创建\n"}, "AppCertificateListRequest": {"Status": "对 AppCertificate 名进行like查询\n"}, "ApplicationCreateRequest": {"Config": "配置信息，eg: 钉钉通知地址\n", "IsExternalRepo": "是否外置仓库\n", "Mode": "模式 LIBRARY, SERVICE, BIGDATA, ABILITY\n", "RepoConfig": "仓库配置 isExternalRepo=true时设置\n"}, "ApplicationDTO": {"CreatedAt": "应用创建时间\n", "Creator": "创建者的userId\n", "IsExternalRepo": "是否外置仓库\n", "MemberRoles": "成员对应的角色\n", "Mode": "模式 LIBRARY, SERVICE, BIGDATA\n", "Stats": "统计信息\n", "UpdatedAt": "应用更新时间\n"}, "ApplicationFetchRequest": {"ApplicationIDOrName": "应用id/应用名\n", "ProjectID": "当path中传的是applicationName的时候，需要传入projectId\n"}, "ApplicationInitRequest": {"BundleID": "+optional ios bundle id, 移动应用时必传\n", "MobileAppName": "+optional 移动应用模板名称, 移动应用时必传\n", "MobileDisplayName": "+optional 移动应用显示名称\n", "PackageName": "+optional android package name, 移动应用时必传\n"}, "ApplicationListRequest": {"IsSimple": "是否只返回简单信息(应用级流水线打开列表使用)\n", "Query": "对项目名进行like查询\n"}, "ApplicationStats": {"CountMembers": "成员人数\n", "CountRuntimes": "runtime 数量\n"}, "ApplicationUpdateRequestBody": {"Config": "配置信息，eg: 钉钉通知地址\n", "Desc": "应用描述信息\n", "DisplayName": "展示名称\n", "Logo": "应用logo信息\n"}, "ApplicationWorkspace": {"Workspace": "工作空间 DEV,TEST,STAGING,PROD\n"}, "AttachDest": {"Namespace": "runtime.Service[x].Namespace\n", "Path": "容器中的路径\n"}, "Audit": {"AppID": "+optional 应用id\n", "AuditLevel": "+optional  事件等级\n", "ClientIP": "+optional 客户端地址\n", "Context": "+optional 事件上下文，前端用来渲染的键值对，如appName，projectName\n", "EndTime": "+required 事件结束时间\n", "ErrorMsg": "+optional 如果失败，可以记录失败原因\n", "FDPProjectID": "+optional fdp项目id\n", "OrgID": "+optional 企业id\n", "ProjectID": "+optional 项目id\n", "Result": "+required 操作结果\n", "ScopeID": "+required scope id\n", "ScopeType": "+required scope type\n", "StartTime": "+required 事件开始时间\n", "TemplateName": "+required 前端模版名，告诉前端应该用哪个模版来渲染\n", "UserAgent": "+optional 客户端类型\n", "UserID": "+required 用户id\n"}, "AuditListCleanCronRequest": {"OrgID": "+required 企业ID\n"}, "AuditSetCleanCronRequest": {"Interval": "+required 事件清理周期\n", "OrgID": "+required 企业ID\n"}, "AuditsListRequest": {"EndAt": "+required 事件结束事件\n", "FDPProjectID": "+optional fdp项目id\n", "OrgID": "+required 企业ID\n", "PageNo": "default 1\n", "PageSize": "default 20\n", "StartAt": "+required 事件开始时间\n", "Sys": "+required 是否是查看系统的事件\n", "UserID": "+optional 通过用户id过滤事件\n"}, "Authorize": {"Key": "权限key\n"}, "AutoTestSpace": {"CreatedAt": "CreatedAt 创建时间\n", "DeletedAt": "DeletedAt 删除时间\n", "SourceSpaceID": "被复制的源测试空间\n", "UpdatedAt": "UpdatedAt 更新时间\n"}, "BaseResource": {"Mem": "Mem 内存大小\n"}, "Bind": {"ContainerPath": "ContainerPath 指容器路径\n", "HostPath": "HostPath 指宿主机路径\n", "ReadOnly": "ReadOnly 是可选的，默认值是 false (read/write)\n"}, "Blame": {"Commit": "提交commit\n", "EndLineNo": "结束行号\n", "StartLineNo": "起始行号\n"}, "BranchRule": {"ArtifactWorkspace": "制品可部署的环境\n", "NeedApproval": "project级别\n", "Rule": "分支规则 eg:master,feature/*\n", "Workspace": "通过分支创建的流水线环境\n"}, "CertificateListRequest": {"Query": "对Certificate名进行like查询\n"}, "CheckRun": {"Commit": "提交commitID\n", "CompletedAt": "完成时间\n", "CreatedAt": "创建时间\n", "ExternalID": "外部系统 ID\n", "MrID": "Merge-Request ID\n", "Name": "检查任务名称 golang-lint/java-lint/api-test\n", "PipelineID": "流水线 ID\n", "RepoID": "仓库ID\n", "Result": "运行结果 success：成功 failed：失败 cancel：取消 timeout：超时\n", "Status": "运行状态 in_progress：进行中 completed：已完成\n", "Type": "检查类型 CI\n"}, "CloudAccount": {"AccessKeyID": "云账号ak\n", "AccessSecret": "云账号as\n"}, "CloudAddonResourceDeleteRequest": {"AddonID": "optional (addon request needed)\n", "InstanceID": "optional (来自云管的请求需要填)\n", "ProjectID": "optional (addon request needed)\n", "RecordID": "optional (addon request needed)\n", "Source": "来自addon, 还是云管理（resource）\n", "Vendor": "optional (来自云管的请求需要填)\n"}, "CloudClusterContainerInfo": {"DockerRoot": "容器服务配置信息\n"}, "CloudClusterInfo": {"ClusterName": "边缘集群配置信息\n", "CollectorURL": "中心集群配置信息，自动获取\n"}, "CloudClusterNewCreateInfo": {"CloudVendor": "云供应商信息\n", "CloudVendorName": "从CloudVendor中解析\n", "K8sVersion": "k8s/ecs相关配置\n", "NatGatewayID": "nat网关配置\n", "Region": "云环境vpc配置信息\n", "VSwitchID": "从已有vswitch创建，指定该值；否则新建vswitch，指定VSwitchCIDR\n", "VpcID": "从已有vpc创建，指定该值；否则新建vpc，指定VpcCIDR\n"}, "CloudClusterRequest": {"OrgID": "企业信息\n"}, "CloudResourceMysqlAccount": {"Account": "以字母开头，以字母或数字结尾。\n由小写字母、数字或下划线组成。\n长度为2~16个字符。\n", "Password": "长度为8~32个字符。\n由大写字母、小写字母、数字、特殊字符中的任意三种组成。\n特殊字符为!@#$\u0026%^*()_+-=\n"}, "CloudResourceMysqlBasicData": {"Category": "Basic：基础版\nHighAvailability：高可用版\nFinance：三节点企业版\n"}, "CloudResourceMysqlDB": {"Accounts": "accounts for a databases\n", "AddonID": "addon bound to this database\n"}, "CloudResourceMysqlDBInfo": {"InstanceID": "mysql instance id\n"}, "CloudResourceMysqlDBRequest": {"DBName": "optional, if not specified, return all db info, 由小写字母、数字、下划线或中划线组成\n"}, "CloudResourceMysqlDetailInfoData": {"Category": "Basic：基础版, HighAvailability：高可用版, AlwaysOn：集群版, Finance：三节点企业版\n", "Host": "connection string\n"}, "CloudResourceMysqlDetailInfoRequest": {"InstanceID": "get from request path\n"}, "CloudResourceMysqlListAccountRequest": {"InstanceID": "get from request path\n"}, "CloudResourceMysqlListDatabaseRequest": {"InstanceID": "get from request path\n"}, "CloudResourceOnsBasicData": {"InstanceType": "实例类型。取值说明如下：\n1：后付费实例\n2：铂金版实例\n", "Status": "实例状态。取值说明如下：\n0：铂金版实例部署中\n2：后付费实例已欠费\n5：后付费实例或铂金版实例服务中\n7：铂金版实例升级中且服务可用\n"}, "CloudResourceOnsGroupBaseInfo": {"GroupId": "以 “GID_“ 或者 “GID-“ 开头，只能包含字母、数字、短横线（-）和下划线（_）,长度限制在 5–64 字节之间,\nGroup ID 一旦创建，将无法再修改\n", "GroupType": "tcp：默认值，表示创建的 Group ID 仅适用于 TCP 协议的消息收发\nhttp：表示创建的 Group ID 仅适用于 HTTP 协议的消息收发\n"}, "CloudResourceOnsGroupInfoRequest": {"GroupID": "optional, if not provide, return all group info\n", "GroupType": "optional, filter by group type\n"}, "CloudResourceOnsTopicInfo": {"List": "topics\n"}, "CloudResourceOnsTopicInfoRequest": {"TopicName": "optional, if not specified, return all topics info\n"}, "CloudResourceOssDetailInfoData": {"Acl": "Bucket的ACL权限: private、public-read、public-read-write\n", "Location": "Bucket的地域\n"}, "CloudResourceOverviewRequest": {"Region": "optional\n", "Vendor": "optional\n"}, "CloudResourceRedisDetailInfoData": {"ArchitectureType": "cluster（集群版）, standard（标准版）, SplitRW（读写分离版）\n", "Capacity": "存储容量，单位：MB\n", "Connections": "实例最大连接数\n", "ID": "实例ID\n", "Name": "名称\n", "NetworkType": "网络类型（vpc/vsw信息）\n", "PrivateHost": "私网地址\n", "PublicHost": "公网地址\n", "RegionId": "地域/可用区\n", "Spec": "实例类型\n", "Status": "状态\n", "Version": "版本\n"}, "CloudResourceSetTagRequest": {"InstanceID": "Tag一级资源时，InstanceID 为空\nTag二级资源时，此处指定InstanceID, 如指定ons id, 然后在resource ids 中指定ons_group/ons_topic\n", "ResourceType": "一级资源\n\tVPC：VPC实例\n\tVSWITCH：交换机实例\n\tEIP：弹性公网IP实例\n\tOSS\n\tONS\n二级资源\n\tONS_TOPIC\n\tONS_GROUP\n"}, "ClusterInfo": {"IsRelation": "是否关联集群，Y: 是，N: 否\n", "System": "Resource       *aliyun.AliyunResources `json:\"resource\"`  // TODO: 重构优化\n"}, "ClusterLabels": {"LabelsInfo": "返回key,value形式; 如key: label, value: labelInfo\n"}, "ClusterLabelsRequest": {"Cluster": "查询集群需要带的query参数\n"}, "ClusterQueryRequest": {"Cluster": "查询集群需要带的query参数\n"}, "ClusterResourceResponse": {"Data": "返回key,value形式; 主要包括:\nkey: projects, value: 10\nkey: applications, value: 10\nkey: runtimes, value: 10\nkey: hosts, value: 10\nkey: abnormalHosts, value: 10\n"}, "Comment": {"CommentID": "评论ID\n", "CommentType": "工单评论类型\n", "Content": "评论内容\n", "CreatedAt": "创建时间\n", "IRComment": "关联任务工单\n", "TicketID": "工单ID\n", "UpdatedAt": "更新时间\n", "UserID": "评论用户ID\n"}, "CommentCreateRequest": {"CommentType": "评论类型\n", "Content": "评论内容\n", "IRComment": "关联事件评论内容\n", "TicketID": "工单ID\n", "UserID": "评论用户ID\n"}, "CommentCreateResponse": {"Data": "评论ID\n"}, "CommentUpdateRequestBody": {"Content": "评论内容\n"}, "CommentUpdateResponse": {"Data": "评论ID\n"}, "Component": {"Data": "组件业务数据\n", "Name": "组件名字\n", "Operations": "组件相关操作（前端定义）\n", "Props": "table 动态字段\n", "State": "前端组件状态\n", "Type": "组件类型\n"}, "ComponentIngressUpdateRequest": {"ClusterName": "若为空，则使用当前集群名称\n", "IngressName": "若为空，则使用ComponentName\n", "Routes": "若为空，则清除ingress\n"}, "ComponentProtocolRequest": {"DebugOptions": "DebugOptions debug 选项\n", "Protocol": "初次请求为空，事件出发后，把包含状态的protocol传到后端\n"}, "ComponentProtocolResponseData": {"Protocol": "后端渲染后的protocol返回前端\n"}, "ComponentProtocolScenario": {"ScenarioKey": "场景Key\n", "ScenarioType": "场景类型, 没有则为空\n"}, "CreateCloudResourceBaseInfo": {"ClientToken": "optional\n", "ClusterName": "optional (addon request need)\n", "OrgID": "optional\n", "ProjectID": "optional (addon request need)\n", "Source": "请求来自addon还是云管（addon, resource）\n", "UserID": "optional\n", "VSwitchID": "optional\n", "VpcID": "optional, 一个region可能有多个vpc，需要选择一个，然后还需要据此添加白名单\n", "ZoneID": "optional, 根据资源密集度选择\n"}, "CreateCloudResourceChargeInfo": {"AutoRenew": "是否开启自动付费\n", "AutoRenewPeriod": "optional, auto generate based on charge period if not provide\n"}, "CreateCloudResourceMysqlAccountRequest": {"Password": "长度为8~32个字符。\n由大写字母、小写字母、数字、特殊字符中的任意三种组成。\n特殊字符为!@#$\u0026%^*()_+-=\n"}, "CreateCloudResourceMysqlRequest": {"Databases": "optional, 创建mysql addon时需要指定database信息\n", "SecurityIPList": "optional, 后端根据vpc信息填充\n", "SpecSize": "mysql instance spec\n", "SpecType": "普通版，高可用版\n", "StorageType": "optional, 后端填充\n", "Version": "支持版本5.7\n"}, "CreateCloudResourceOnsRequest": {"Remark": "备注说明\n", "Topics": "optional\n"}, "CreateCloudResourceRedisRequest": {"AddonID": "来自addon的请求需要\n", "Password": "optional, generated by backend\n实例密码。 长度为8－32位，需包含大写字母、小写字母、特殊字符和数字中的至少三种，允许的特殊字符包括!@#$%^\u0026*()_+-=\n", "Spec": "eg. redis.master.mid.default\t(标准版，双副本，2G)\n"}, "CreateHookRequest": {"Active": "是否激活\n", "Events": "webhook 所关心事件的列表\n", "Name": "webhook 名字\n", "URL": "webhook URL, 后续的事件触发时，会POST到该URL\n"}, "CreateNotifyItemResponse": {"Data": "创建通知项的id\n"}, "CreateRepoRequest": {"IsLocked": "是否锁定\n", "OnlyCheck": "做仓库创建检测，不实际创建\n"}, "CreateRepoResponseData": {"RepoPath": "仓库相对路基\n"}, "CustomAddonCreateRequest": {"AddonName": "addon名称\n", "Configs": "环境变量 custom addon的环境变量配置\n", "CustomAddonType": "三方addon类型 custom或者cloud，云服务就是cloud\n", "Name": "实例名称\n", "OperatorID": "操作人\n", "Options": "补充信息，云addon的信息都放在这里\n", "ProjectID": "项目ID\n", "Tag": "标签\n", "Workspace": "所属环境\n"}, "CustomAddonUpdateRequest": {"Configs": "环境变量 custom addon的环境变量配置\n", "Options": "补充信息，云addon的信息都放在这里\n"}, "DashBoardDTO": {"DrawerInfoMap": "绘制信息\n", "Id": "记录主键id\n", "Layout": "布局信息\n", "UniqueId": "唯一标识\n"}, "DashboardCreateRequest": {"DrawerInfoMap": "绘制相关json\n", "Layout": "布局相关json\n"}, "DashboardDetailRequest": {"Id": "配置id\n"}, "DeploymentListRequest": {"OrgID": "Org ID, 获取 'orgid' 下的所有 runtime 的 deployments\n", "RuntimeID": "应用实例 ID\n", "StatusIn": "通过 Status 过滤，不传为默认不过滤\n"}, "DeploymentStatusDTO": {"DeploymentID": "发布Id\n", "FailCause": "失败原因\n", "ModuleErrMsg": "模块错误信息\n", "Phase": "发布过程\n", "Status": "状态\n"}, "DereferenceClusterRequest": {"Cluster": "查询集群需要带的query参数\n", "OrgID": "企业ID\n"}, "Dice": {"ID": "name of dice, namespace + name is unique\nID is the hash string identity for dice info like 'x389vj1l23...'\n", "Labels": "labels for extension and some tags\n", "ProjectNamespace": "Namespace indicates namespace for kubernetes\n", "ServiceDiscoveryKind": "service discovery kind: VIP, PROXY, NONE\n", "ServiceDiscoveryMode": "Defines the way dice do env injection.\n\nGLOBAL:\n  each service can see every services\nDEPEND:\n  each service can see what he depends (XXX_HOST, XXX_PORT)\n", "Services": "bunch of services running together with dependencies each other\n", "Type": "namespace of dice, namespace + name is unique\nType indicates the type of dice, it contains services, group-addon ...\nType and ID will compose the unique namespaces for kubernetes when Namespaces is empty\n"}, "DiffFile": {"Index": "40-byte SHA, Changed/New: new SHA; Deleted: old SHA\n"}, "DomainListRequest": {"RuntimeID": "应用实例 ID\n"}, "DomainUpdateRequest": {"RuntimeID": "应用实例 ID\n"}, "DrainNodeRequest": {"DeleteLocalData": "Continue even if there are pods using emptyDir (local data that will be deleted when the node is drained)\n", "DisableEviction": "DisableEviction forces drain to use delete rather than evict\n", "Force": "Continue even if there are pods not managed by a ReplicationController, ReplicaSet, Job, DaemonSet or StatefulSet\n", "GracePeriodSeconds": "Period of time in seconds given to each pod to terminate gracefully. If negative, the default value specified in the pod will be use\n", "IgnoreAllDaemonSets": "Ignore DaemonSet-managed pods\n", "PodSelector": "Label selector to filter pods on the node\n", "SkipWaitForDeleteTimeoutSeconds": "SkipWaitForDeleteTimeoutSeconds ignores pods that have a\nDeletionTimeStamp \u003e N seconds. It's up to the user to decide when this\noption is appropriate; examples include the Node is unready and the pods\nwon't drain otherwise\n", "Timeout": "The length of time to wait before giving up, zero means infinite\n"}, "EditActionItem": {"Action": "支持操作 add/delete\n", "PathType": "支持类型 tree/blob\n"}, "EffectivenessRequest": {"PluginParamDto": "插件参数\n"}, "EndpointDomainsItem": {"Domain": "Domain 域名\n", "Type": "Type 域名类型,CUSTOM or DEFAULT\n"}, "EnvConfig": {"ConfigType": "ENV, FILE\n", "Operations": "Operations 配置项操作，若为 nil，则使用默认配置: canDownload=false, canEdit=true, canDelete=true\n"}, "ErrorLogListRequest": {"ResourceID": "+required 资源id\n", "ResourceType": "+required 资源类型\n", "ScopeID": "+required 鉴权需要\n", "ScopeType": "+required 鉴权需要\n", "StartTime": "+option 根据时间过滤错误日志\n"}, "EventHeader": {"TimeStamp": "Content   PlaceHolder `json:\"content\"`\n"}, "ExecHealthCheck": {"Duration": "单位是秒\n"}, "ExistsMysqlExec": {"MysqlHost": "MysqlHost host地址\n", "MysqlPort": "MysqlPort mysqlPort\n", "Options": "Options 额外信息\n", "Password": "Password 登录密码\n", "User": "User 登录用户\n"}, "ExtensionCreateRequest": {"Type": "类型 addon|action\n"}, "ExtensionQueryRequest": {"All": "默认false查询公开的扩展, true查询所有扩展\n", "Labels": "根据标签查询 key:value 查询满足条件的 ^key:value 查询不满足条件的\n", "Type": "可选值: action、addon\n"}, "ExtensionSearchRequest": {"Extensions": "支持格式 name:获取默认版本 name@version:获取指定版本\n"}, "ExtensionVersionCreateRequest": {"All": "是否一起更新ext和version,默认只更新version,只在forceUpdate=true有效\n", "ForceUpdate": "为true的情况如果已经存在相同版本会覆盖更新,不会报错\n", "Public": "是否公开\n"}, "ExtensionVersionQueryRequest": {"All": "默认false查询有效版本, true查询所有版本\n"}, "FuzzyQueryNotifiesBySourceRequest": {"PageNo": "查询条件\n"}, "GetRuntimeAddonConfigRequest": {"ClusterName": "集群名称\n", "ProjectID": "项目Id\n", "RuntimeID": "runtimeId\n", "Workspace": "环境\n"}, "GitRepoConfig": {"Password": "仓库密码\n", "Type": "类型, 支持类型:general\n", "Url": "仓库地址\n", "Username": "仓库用户名\n"}, "GittarCommitsRequest": {"Search": "commit message过滤条件\n"}, "GittarCreateBranchRequest": {"Ref": "引用, branch/tag/commit\n"}, "GittarCreateCommitRequest": {"Actions": "变更操作列表\n", "Branch": "更新到的分支\n"}, "GittarCreateTagRequest": {"Ref": "引用, branch/tag/commit\n"}, "GittarFileData": {"Binary": "是否为二进制文件,如果是lines不会有内容\n"}, "GittarLinesData": {"Binary": "是否为二进制文件,如果是lines不会有内容\n"}, "GittarMergeStatusRequest": {"SourceBranch": "源分支\n", "TargetBranch": "将要合并到的目标分支\n"}, "GittarQueryMrRequest": {"AssigneeId": "分配人\n", "AuthorId": "创建人\n", "Page": "页数\n", "Query": "查询title模糊匹配或者merge_id精确匹配\n", "Score": "评分\n", "Size": "每页数量\n", "State": "状态 open/closed/merged\n"}, "GittarStatsData": {"ContributorCount": "提交的人数\n", "Empty": "仓库是否为空\n", "MergeRequestCount": "open状态的mr数量\n"}, "GittarTreeSearchRequest": {"Pattern": "文件通配符 例如 *.workflow\n", "Ref": "支持引用名: branch/tag/commit\n"}, "HealthCheck": {"Command": "command for COMMAND\n", "Kind": "healthCheck kinds: HTTP, HTTPS, TCP, COMMAND\n", "Path": "path for HTTP, HTTPS\n", "Port": "port for HTTP, HTTPS, TCP\n"}, "Hierarchy": {"Structure": "structure的结构可能是list、map\n"}, "Hook": {"ID": "webhook ID\n", "Secret": "用于计算后续发送的事件内容的sha值，目前没有用\n"}, "HookLocation": {"Application": "webhook 所属 applicationID\n", "Env": "webhook 所关心环境, nil 代表所有\n", "Org": "webhook 所属 orgID\n", "Project": "webhook 所属 projectID\n"}, "HttpHealthCheck": {"Duration": "单位是秒\n"}, "IdentityInfo": {"InternalClient": "InternalClient records the internal client, such as: bundle.\nCannot be null if UserID is null.\n+optional\n", "UserID": "UserID is user id. It must be provided in some cases.\nCannot be null if InternalClient is null.\n+optional\n"}, "ImageCreateRequest": {"ReleaseID": "关联release\n"}, "ImageListRequest": {"PageNum": "当前页号，默认值1\n", "PageSize": "分页大小,默认值20\n"}, "ImageSearchRequest": {"PageNum": "当前页号，默认值1\n", "PageSize": "分页大小,默认值20\n", "Query": "查询参数，eg:app:test\n"}, "InstanceDetailRes": {"AddonName": "addon名称\n", "AttachCount": "被引用次数\n", "CanDel": "是否可被删除\n", "ClusterName": "集群名称\n", "Config": "环境变量\n", "CreateAt": "创建时间\n", "Env": "所属环境\n", "EnvCn": "所属环境中文描述\n", "InstanceName": "addon实例名称\n", "LogoURL": "logo图片地址\n", "PlanCnName": "规格中文说明\n", "Platform": "是否平台属性\n", "ProjectID": "项目ID\n", "ProjectName": "项目名称\n", "ReferenceInfo": "引用信息\n", "Status": "addon状态\n", "Version": "版本\n"}, "InstanceInfoRequest": {"InstanceIP": "ip1,ip2,ip3\n", "Phases": "enum: unhealthy, healthy, dead, running\n", "ServiceType": "enum: addon, stateless-service, job\n", "Workspace": "enum: dev, test, staging, prod\n"}, "InstanceReferenceRes": {"ApplicationID": "应用ID\n", "ApplicationName": "应用名称\n", "OrgID": "企业ID\n", "ProjectID": "项目ID\n", "ProjectName": "项目名称\n", "RuntimeID": "runtime ID\n", "RuntimeName": "runtime名称\n"}, "InstanceStatusData": {"Host": "宿主机ip\n", "ID": "事件id\nk8s 中是 containerID\nmarathon 中是 taskID\n", "IP": "容器ip\n", "InstanceStatus": "包含Running,Killed,Failed,Healthy,UnHealthy等状态\n", "Message": "事件额外描述，可能为空\n", "Timestamp": "时间戳到纳秒级\n"}, "Issue": {"FinishTime": "切换到已完成状态的时间 （等事件可以记录历史信息了 删除该字段）\n"}, "IssueBatchUpdateRequest": {"CurrentIterationID": "以下字段用于鉴权, 不可更改\n"}, "IssueCreateRequest": {"AppID": "+optional 所属应用 ID\n", "Assignee": "+required 当前处理人\n", "BugStage": "+optionaln bug阶段\n", "Complexity": "+optional 复杂度\n", "Content": "+optional 内容\n", "Creator": "+optional 第三方创建时头里带不了userid，用这个参数显式指定一下\n", "External": "用来区分是通过ui还是bundle创建的\n", "IterationID": "+required 所属迭代 ID\n", "Labels": "+optional 标签名称列表\n", "ManHour": "+optional 工时信息，当事件类型为任务和缺陷时生效\n", "Owner": "+optionaln 负责人\n", "PlanFinishedAt": "+optional 计划结束时间\n", "PlanStartedAt": "+optional 计划开始时间\n", "Priority": "+optional 优先级\n", "ProjectID": "+required 所属项目 ID\n", "Severity": "+optional 严重程度\n", "Source": "+optional 创建来源，目前只有工单使用了该字段\n", "TaskType": "+optionaln 任务类型\n", "TestPlanCaseRelIDs": "+optional 关联的测试计划用例关联 ID 列表\n", "Title": "+required 标题\n", "Type": "+required issue 类型\n"}, "IssueListRequest": {"AppID": "+optional\n", "Asc": "+optional 是否升序排列\n", "Assignees": "+optional\n", "BugStage": "+optionaln bug阶段\n", "Complexity": "+optional 复杂度\n", "Creators": "+optional\n", "EndCreatedAt": "+optional ms\n", "EndFinishedAt": "+optional ms\n", "ExceptIDs": "+optional 排除的id\n", "External": "用来区分是通过ui还是bundle创建的\n", "IDs": "+optional 包含的ID\n", "IsEmptyPlanFinishedAt": "+optional 是否只筛选截止日期为空的事项\n", "IterationID": "+required 迭代id为-1时，即是显示待办事件\n", "IterationIDs": "+required 支持多迭代查询\n", "Label": "+optional\n", "OrderBy": "+optional 排序字段, 支持 planStartedAt \u0026 planFinishedAt\n", "Owner": "+optionaln 负责人\n", "Priority": "+optional 优先级\n", "ProjectID": "+required\n", "RelatedIssueIDs": "+optional\n", "RequirementID": "+optional\n", "Severity": "+optional 严重程度\n", "Source": "+optional 来源\n", "StartCreatedAt": "+optional ms\n", "StartFinishedAt": "+optional ms\n", "State": "+optional\n", "StateBelongs": "+optional\n", "TaskType": "+optionaln 任务类型\n", "Title": "+optional\n", "Type": "+optional\n", "WithProcessSummary": "+optional 是否需要进度统计\n"}, "IssueManHourSumResponse": {"DesignManHour": "Header\n"}, "IssuePagingRequest": {"OrgID": "+required 企业id\n", "PageNo": "+optional default 1\n", "PageSize": "+optional default 10\n"}, "IssueUpdateRequest": {"ManHour": "工时信息，当事件类型为任务和缺陷时生效\n"}, "IterationCreateRequest": {"Content": "+optional\n", "FinishedAt": "+optional\n", "ProjectID": "+required\n", "StartedAt": "+optional\n", "Title": "+required\n"}, "IterationPagingRequest": {"Deadline": "+optional 根据迭代结束时间过滤\n", "PageNo": "+optional default 1\n", "PageSize": "+optional default 10\n", "ProjectID": "+required\n", "State": "+optional 根据归档状态过滤\n", "WithoutIssueSummary": "+optional 是否查询事项概览，默认查询\n"}, "IterationUpdateRequest": {"Content": "+required\n", "FinishedAt": "+required\n", "StartedAt": "+required\n", "State": "+required\n", "Title": "+required\n"}, "JobVolume": {"Name": "用于生成volume id = \u003cnamespace\u003e-\u003cname:\u003e\n", "Type": "nfs | local\n"}, "LibReferenceListRequest": {"AppID": "+optional\n", "ApprovalStatus": "+optional\n", "LibID": "+optional\n", "PageNo": "+optional\n", "PageSize": "+optional\n"}, "ListCloudAddonBasicRequest": {"ProjectID": "optional, by project, e.g addon\n", "VpcID": "optional, by vpc\n", "Workspace": "optional (addon ons request need: DEV/TEST/STAGING/PRO)\n"}, "ListCloudResourceECSRequest": {"Cluster": "optional\n", "InnerIpAddress": "optional\n", "Region": "optional\n", "Vendor": "enum: aliyun\n"}, "ListCloudResourceVPCRequest": {"Cluster": "optional\n", "Region": "optional\n", "Vendor": "enum: aliyun\n"}, "ListLabelsData": {"IsPrefix": "类似 org-, 都是前缀 label\n"}, "ListSchemasQueryParams": {"AppID": "如果不传 inode, 就必须传 appID 和 branch\n", "Inode": "branch 节点的 inode,\n在实现中, 如果传了 branch 以下节点的 inode, 也会被处理成 branch 节点的 inode\n"}, "Member": {"Deleted": "uc注销用户的标记，用于分页查询member时的返回\n", "Labels": "成员标签，多标签\n", "Removed": "被移除标记, 延迟删除\n", "Roles": "成员角色，多角色\n", "Scope": "成员的归属\n", "Status": "Deprecated: 当前用户的状态,兼容老数据\n"}, "MemberAddOptions": {"Rewrite": "是否覆盖已存在的成员\n"}, "MemberAddRequest": {"Labels": "成员标签，多标签\n", "Options": "Deprecated: 可选选项\n", "Roles": "成员角色，多角色\n", "Scope": "成员的归属\n", "TargetScopeType": "TargetScopeType，TargetScopeIDs 要加入的scope，当这个参数有时，scope 参数只用来鉴权，不作为目标scope加入\n", "UserIDs": "要添加的用户id列表\n", "VerifyCode": "邀请成员加入验证码\n"}, "MemberDestroyRequest": {"UserIDs": "要添加的用户id列表\n"}, "MemberLabelList": {"List": "角色标签\n"}, "MemberListRequest": {"Labels": "过滤标签\n", "Q": "查询参数\n", "Roles": "过滤角色\n", "ScopeID": "对应的 orgId, projectId, applicationId\n", "ScopeType": "类型 sys, org, project, app\n"}, "MemberRemoveRequest": {"Scope": "成员的归属\n", "UserIDs": "要添加的用户id列表\n"}, "MergeTemplatesResponseData": {"Branch": "所在分支\n", "Names": "模板文件列表\n", "Path": "模板所在目录\n"}, "MicroProjectMenuRes": {"AddonDisplayName": "addon展示名称\n", "AddonName": "addon名称\n", "ConsoleURL": "console地址\n", "InstanceID": "实例Id\n", "ProjectName": "项目名称\n", "TerminusKey": "监控terminus key\n"}, "MicroProjectRes": {"Envs": "所属环境\n", "LogoURL": "project logo信息\n", "MicroTotal": "数量\n", "ProjectID": "项目ID\n", "ProjectName": "项目名称\n"}, "MiddlewareFetchResponseData": {"AddonName": "AddonName addon 名称\n", "AttachCount": "AttachCount 引用数量\n", "Category": "Category addon 类别: 微服务/数据库/配置中心，etc\n", "Cluster": "Cluster 集群名称\n", "Config": "Config addon 使用配置, eg: 地址/端口/账号\n", "CreatedAt": "CreatedAt 创建时间\n", "InstanceID": "InstanceID 实例ID\n", "LogoURL": "LogoURL addon logo\n", "Plan": "Plan addon 规格, basic/professional, etc\n", "ProjectID": "项目ID\n", "ReferenceInfos": "ReferenceInfos 引用详情\n", "Status": "Status addon 状态\n", "UpdatedAt": "UpdatedAt 更新时间\n", "Version": "Version addon 版本\n", "Workspace": "Workspace， DEV/TEST/STAGING/PROD\n"}, "MiddlewareListItem": {"AddonName": "addon名称\n", "AttachCount": "引用数\n", "CPU": "cpu\n", "ClusterName": "环境\n", "Env": "环境\n", "InstanceID": "实例ID\n", "Mem": "内存\n", "Name": "名称\n", "Nodes": "节点数\n", "ProjectID": "项目ID\n", "ProjectName": "项目名称\n"}, "MiddlewareListRequest": {"AddonName": "AddonName addon 名称\n", "EndTime": "EndTime 截止时间\n", "InstanceID": "InstanceID addon真实例ID\n", "InstanceIP": "ip1,ip2,ip3\n", "PageNo": "PageNo 当前页，默认值: 1\n", "PageSize": "PageSize 分页大小，默认值: 20\n", "ProjectID": "ProjectID 项目Id\n", "Workspace": "Workspace 工作环境，可选值: DEV/TEST/STAGING/PROD\n"}, "MiddlewareResourceFetchResponseData": {"InstanceID": "InstanceID 实例ID\n"}, "MigrationStatusDesc": {"Desc": "Desc 说明信息\n", "Status": "Status 返回的运行状态\n"}, "MultiLevelStatus": {"More": "More 是扩展字段，比如存储runtime下每个服务的名字及状态\n", "Name": "Name 指 runtime name\n", "Namespace": "Namespace 指 runtime namespace\n", "Status": "Status 指 runtime status\n"}, "MysqlDataBaseInfo": {"CharacterSetName": "optional, default uft8mb4\n"}, "MysqlExec": {"CreateDbs": "CreateDbs 要创建的数据库\n", "Host": "Host mysql host\n", "OssURL": "OssURL init.sql地址\n", "Password": "Password 登录密码\n", "Sqls": "Sqls 执行语句\n", "URL": "URL url\n", "User": "User 登录用户\n"}, "NamespaceCreateRequest": {"Dynamic": "该namespace下配置是否推送至远程配置中心\n", "IsDefault": "是否为default namespace\n", "Name": "namespace名称\n", "ProjectID": "项目ID\n"}, "NamespaceRelationCreateRequest": {"DefaultNamespace": "default namespace\n", "RelatedNamespaces": "dev/test/staging/prod四个环境namespace\n"}, "NexusUserEnsureRequest": {"ClusterName": "ClusterName 属于哪个集群的 nexus\n+required\n", "ForceUpdatePassword": "+optional\n是否强制更新密码，ensure 场景一般需要保留原密码，因为原密码可能正在被打包使用中\n", "OrgID": "OrgID 关联 org 信息\n+optional\n", "Password": "+required\n", "RepoID": "RepoID 关联 repo 信息\n+optional\n", "RepoPrivileges": "RepoPrivileges 关联的 repo 权限\n+optional\n", "SyncConfigToPipelineCM": "+optional\n", "UserName": "+required\n"}, "NodeResourceInfo": {"IgnoreLabels": "dcos, edas 缺少一些 label 或无法获取 label, 所以告诉上层忽略 labels\n", "Labels": "only 'dice-' prefixed labels\n"}, "NoticeListRequest": {"Content": "+optional\n", "OrgID": "+required 后端赋值\n", "PageNo": "+optional\n", "PageSize": "+optional\n", "Status": "+optional\n"}, "NotifyHistory": {"NotifyItemDisplayName": "todo json key名需要cdp前端配合修改后再改\n"}, "NotifyItem": {"CalledShowNumber": "语音通知的被叫显号，语音模版属于公共号码池外呼的时候，被叫显号必须是空\n属于专属号码外呼的时候，被叫显号不能为空\n", "VMSTemplate": "语音通知模版\n"}, "OneDataAnalysisBussProcRequest": {"FilePath": "本地仓库文件绝对路径\n"}, "OneDataAnalysisBussProcsRequest": {"BusinessDomain": "业务板块\n", "DataDomain": "数据域\n", "KeyWord": "搜索关键字\n", "PageNo": "页码\n", "PageSize": "行数\n", "RemoteUri": "模型远程仓库地址\n"}, "OneDataAnalysisDimRequest": {"FilePath": "本地仓库文件绝对路径\n"}, "OneDataAnalysisFuzzyAttrsRequest": {"FilePath": "本地仓库文件绝对路径\n", "KeyWord": "搜索关键字\n", "PageNo": "页码\n", "PageSize": "行数\n"}, "OneDataAnalysisOutputTablesRequest": {"BusinessDomain": "业务板块\n", "KeyWord": "搜索关键字\n", "MarketDomain": "集市域\n", "PageNo": "页码\n", "PageSize": "行数\n", "RemoteUri": "模型远程仓库地址\n"}, "OneDataAnalysisRequest": {"RemoteUri": "模型远程仓库地址\n"}, "OneDataAnalysisStarRequest": {"FilePath": "本地仓库文件绝对路径\n"}, "OnsEndpoints": {"HttpInternalEndpoint": "Http 协议客户端接入点\n", "TcpEndpoint": "Tcp 协议客户端接入点\n"}, "OnsTopic": {"MessageType": "消息类型\n", "RelationName": "权限\n", "Remark": "描述\n", "TopicName": "Topic 名称\n"}, "OrgCreateRequest": {"Admins": "创建组织时作为admin的用户id列表\n", "PublisherName": "发布商名称\n"}, "OrgDTO": {"Domain": "企业域名\n", "EnableReleaseCrossCluster": "开关：制品是否允许跨集群部署\n", "Operation": "操作者id\n", "PublisherID": "发布商 ID\n", "Selected": "用户是否选中当前企业\n", "Status": "组织状态\n"}, "OrgNexusGetRequest": {"Formats": "+optional\n", "Types": "+optional\n"}, "OrgResourceInfo": {"TotalCpu": "单位: c\n", "TotalMem": "单位: GB\n"}, "OrgRunningTasksListRequest": {"AppName": "应用名称，选填\n", "Cluster": "集群名称参数，选填\n", "EndTime": "截止时间戳(ms)，选填 默认为当前时间\n", "Env": "环境，选填\n", "PageNo": "页号, 默认值:1\n", "PageSize": "分页大小, 默认值20\n", "PipelineID": "pipeline ID，选填\n", "ProjectName": "项目名称，选填\n", "StartTime": "起始时间戳(ms)，选填\n", "Status": "状态，选填\n", "Type": "task类型参数: job或者deployment, 选填\n", "UserID": "创建人，选填\n"}, "OrgSearchRequest": {"PageNo": "分页参数\n", "Q": "用此对组织名进行模糊查询\n"}, "PageInfo": {"PageNO": "页码\n", "PageSize": "每页大小\n"}, "PagePipeline": {"CostTimeSec": "时间\n", "IsSnippet": "嵌套流水线相关信息\n", "Type": "运行时相关信息\n"}, "Parameter": {"Name": "参数名\n"}, "PermissionCheckRequest": {"Action": "Action Create/Update/Delete/\n", "Resource": "Resource 资源类型， eg: ticket/release\n", "ResourceRole": "resource 角色: Creator, Assignee\n", "Scope": "Scope 可选值: org/project/app\n", "ScopeID": "ScopeID scope具体值\n"}, "PermissionList": {"ContactsWhenNoPermission": "无权限（access=false）时，该字段返回联系人 ID 列表，例如无应用权限时，返回应用管理员列表\n", "Exist": "当项目/应用被删除时，鉴权为false，用于告诉前端是被删除了\n"}, "PipelineButton": {"CanCancel": "取消\n", "CanDelete": "删除\n", "CanManualRun": "手动开始\n", "CanPause": "TODO 暂停\n", "CanRerun": "重试\n", "CanStartCron": "定时\n"}, "PipelineCmsConfigValue": {"Comment": "Comment\n", "EncryptInDB": "EncryptInDB 在数据库中是否加密存储\n", "From": "From 配置项来源，可为空。例如：证书管理同步\n", "Operations": "Operations 配置项操作，若为 nil，则使用默认配置: canDownload=false, canEdit=true, canDelete=true\n", "TimeCreated": "创建或更新时以下字段无需填写\n", "Type": "Type\nif not specified, default type is `kv`;\nif type is `dice-file`, value is uuid of `dice-file`.\n", "Value": "Value\n更新时，Value 为 realValue\n获取时，若 Decrypt=true，Value=decrypt(dbValue)；若 Decrypt=false，Value=dbValue\n"}, "PipelineCreateRequestV2": {"AutoRun": "AutoRun represents whether auto run the created pipeline.\nDefault is false.\n+optional\nDeprecated, please use `AutoRunAtOnce` or `AutoStartCron`.\nAlias for AutoRunAtOnce.\n", "AutoRunAtOnce": "AutoRunAtOnce alias for `AutoRun`.\nAutoRunAtOnce represents whether auto run the created pipeline.\nDefault is false.\n+optional\n", "AutoStartCron": "AutoStartCron represents whether auto start cron.\nIf a pipeline doesn't have `cron` field, ignore.\nDefault is false.\n+optional\n", "ClusterName": "ClusterName represents the cluster the pipeline will be executed.\n+required\n", "ConfigManageNamespaces": "ConfigManageNamespaces pipeline fetch configs from cms by namespaces in order.\nPipeline won't generate default ns.\n+optional\n", "CronStartFrom": "CronStartFrom specify time when to start\n+optional\n", "Envs": "Envs is Map of string keys and values.\n+optional\n", "ForceRun": "ForceRun represents stop other running pipelines to run.\nDefault is false.\n+optional\n", "GC": "GC represents pipeline gc configs.\nIf config is empty, will use default config.\n+optional\n", "Labels": "Labels is Map of string keys and values, can be used to filter pipeline.\nIf label key or value is too long, it will be moved to NormalLabels automatically and overwrite value if key already exists in NormalLabels.\n+optional\n", "NormalLabels": "NormalLabels is Map of string keys and values, cannot be used to filter pipeline.\n+optional\n", "PipelineSource": "PipelineSource represents the source where pipeline created from.\nEqual to `Namespace`.\n+required\n", "PipelineYml": "PipelineYml is pipeline yaml content.\n+required\n", "PipelineYmlName": "PipelineYmlName\nEqual to `Name`.\nDefault is `pipeline.yml`.\n+optional\n", "RunParams": "RunPipelineParams represents pipeline params runtime input\nif pipeline have params runPipelineParams can not be empty\n+optional\n"}, "PipelineDBGCItem": {"NeedArchive": "NeedArchive means whether this record need be archived:\nIf true, archive record to specific archive table;\nIf false, delete record and cannot be found anymore.\n", "TTLSecond": "TTLSecond means when to do archive or delete operation.\n"}, "PipelineDTO": {"Branch": "分支相关信息\n", "CostTimeSec": "时间\n", "ID": "应用相关信息\n", "Namespace": "运行时相关信息\n", "Source": "pipeline.yml 相关信息\n"}, "PipelineDatabaseGC": {"Analyzed": "Analyzed contains gc strategy to analyzed pipeline.\n", "Finished": "Finished contains gc strategy to finished(success/failed) pipeline.\n"}, "PipelineDetailDTO": {"PipelineButton": "按钮\n", "PipelineTaskActionDetails": "task 的 action 详情\n"}, "PipelineIDSelectByLabelRequest": {"AllowNoPipelineSources": "AllowNoPipelineSources, default is false.\n默认查询必须带上 pipeline source，增加区分度\n", "AnyMatchLabels": "ANY match\n", "MustMatchLabels": "MUST match\n", "OrderByPipelineIDASC": "OrderByPipelineIDASC 根据 pipeline_id 升序，默认为 false，即降序\n"}, "PipelineInvokedCombo": {"PipelineID": "其他前端展示需要的字段\n"}, "PipelineInvokedComboRequest": {"AppID": "app id\n", "Branches": "comma-separated value, such as: develop,master\n", "Sources": "comma-separated value, such as: dice,bigdata\n", "YmlNames": "comma-separated value, such as: pipeline.yml,path1/path2/demo.workflow\n"}, "PipelinePageListRequest": {"AnyMatchLabels": "直接构造对象 请赋值该字段\n", "AnyMatchLabelsJSON": "Deprecated\n供 CDP 工作流明细查询使用，JSON(map[string]string)\n", "AnyMatchLabelsQueryParams": "?anyMatchLabel=key1=value1\n\u0026anyMatchLabel=key1=value2\n\u0026anyMatchLabel=key2=value3\n", "AppID": "Deprecated, use mustMatchLabels, key=appID\n", "Branches": "Deprecated, use mustMatchLabels, key=branch\n", "CommaBranches": "Deprecated, use schema `branch`\n", "CommaSources": "Deprecated, use schema `source`\n", "CommaStatuses": "Deprecated, use schema `status`\n", "CommaYmlNames": "Deprecated, use schema `ymlName`\n", "EndTimeBegin": "开始执行时间 右闭区间\n", "EndTimeBeginCST": "Deprecated, use `StartedAtTimestamp`.\nformat: 2006-01-02T15:04:05, TZ: CST\n", "EndTimeBeginTimestamp": "http GET query param 请赋值该字段\n", "EndTimeCreated": "创建时间 右闭区间\n", "EndTimeCreatedTimestamp": "http GET query param 请赋值该字段\n", "IncludeSnippet": "IncludeSnippet 是否展示嵌套流水线，默认不展示。\n嵌套流水线一般来说只需要在详情中展示即可。\n", "MustMatchLabels": "直接构造对象 请赋值该字段\n", "MustMatchLabelsJSON": "Deprecated\n供 CDP 工作流明细查询使用，JSON(map[string]string)\n", "MustMatchLabelsQueryParams": "?mustMatchLabel=key1=value1\n\u0026mustMatchLabel=key1=value2\n\u0026mustMatchLabel=key2=value3\n", "SelectCols": "internal use\n", "StartTimeBegin": "开始执行时间 左闭区间\n", "StartTimeBeginCST": "Deprecated, use `StartedAtTimestamp`.\nformat: 2006-01-02T15:04:05, TZ: CST\n", "StartTimeBeginTimestamp": "http GET query param 请赋值该字段\n", "StartTimeCreated": "创建时间 左闭区间\n", "StartTimeCreatedTimestamp": "http GET query param 请赋值该字段\n"}, "PipelineReportSetPagingRequest": {"EndTimeBeginTimestamp": "开始执行时间 右闭区间\n", "EndTimeCreatedTimestamp": "创建时间 右闭区间\n", "MustMatchLabelsQueryParams": "///////////////////////\npipeline 分页查询参数 //\n///////////////////////\nlabels\n\u0026mustMatchLabel=key2=value3\n", "StartTimeBeginTimestamp": "times\n开始执行时间 左闭区间\n", "StartTimeCreatedTimestamp": "创建时间 左闭区间\n"}, "PipelineResourceGC": {"FailedTTLSecond": "FailedTTLSecond means when to release resource if pipeline status is Failed.\nNormally failed ttl should larger than SuccessTTLSecond, because you may want to rerun this failed pipeline,\nwhich need these resource.\nDefault is 1800s.\n", "SuccessTTLSecond": "SuccessTTLSecond means when to release resource if pipeline status is Success.\nNormally success ttl should be small even to zero, because everything is ok and don't need to rerun.\nDefault is 1800s(30min)\n"}, "PipelineTaskSnippetDetail": {"DirectSnippetTasksNum": "直接子任务数，即 snippet pipeline 的任务数，不会递归查询\n-1 表示未知，具体数据在 reconciler 调度时赋值\n", "RecursiveSnippetTasksNum": "递归子任务数，即该节点下所有子任务数\n-1 表示未知，具体数据由 aop 上报\n"}, "PipelineYml": {"NeedUpgrade": "1.0 升级相关\n", "Version": "用于构造 pipeline yml\n", "YmlContent": "YmlContent:\n1) 当 needUpgrade 为 true  时，ymlContent 返回升级后的 yml\n2) 当 needUpgrade 为 false 时：\n   1) 用户传入的为 YAML(apistructs.PipelineYml) 时，ymlContent 返回 YAML(spec.PipelineYml)\n   2) 用户传入的为 YAML(spec.PipelineYml) 时，返回优化后的 YAML(spec.PipelineYml)\n"}, "PluginParamDto": {"DataSourceId": "数据源Id\n", "FilterColumns": "筛选字段列表\n", "GroupByColumns": "聚合字段列表\n", "Limit": "返回记录数\n", "Offset": "查询其实位置\n", "TableName": "数据表名称\n", "TargetColumns": "目标字段列表\n", "Widget": "展示图形类型，可选:default,line,bar,area,pie,cards,radar,gauge,map,dot\n"}, "PodInfoRequest": {"Phases": "enum: Pending, Running, Succeeded, Failed, Unknown\n", "ServiceType": "enum: addon, stateless-service, job\n", "Workspace": "enum: dev, test, staging, prod\n"}, "ProjectCreateRequest": {"ClusterConfig": "项目各环境集群配置\n", "CpuQuota": "+required 单位: c\n", "Creator": "创建者的用户id\n", "DdHook": "项目级别的dd回调地址\n", "MemQuota": "+required 单位: GB\n", "OrgID": "组织id\n", "RollbackConfig": "项目回滚点配置\n", "Template": "+required 项目模版\n"}, "ProjectDTO": {"ActiveTime": "项目活跃时间\n", "BlockStatus": "解封状态: unblocking | unblocked (目前只有 /api/projects/actions/list-my-projects api 有这个值)\n", "CanUnblock": "当前用户是否可以解封该 project (目前只有 /api/projects/actions/list-my-projects api 有这个值)\n", "ClusterConfig": "项目各环境集群配置\n", "CreatedAt": "项目创建时间\n", "Joined": "用户是否已加入项目\n", "Owners": "项目所有者\n", "Stats": "项目统计信息\n", "UpdatedAt": "项目更新时间\n"}, "ProjectDetailRequest": {"OrgID": "当传入projectName时，需要传入orgId或orgName\n", "OrgName": "当传入projectName时，需要传入orgId或orgName\n", "ProjectIDOrName": "支持项目id/项目名查询\n"}, "ProjectListRequest": {"Asc": "是否升序\n", "Joined": "是否只展示已加入的项目\n", "OrderBy": "排序支持activeTime,memQuota和cpuQuota\n", "Query": "对项目名进行like查询\n"}, "ProjectResourceResponse": {"Data": "key 为 projectID\n"}, "ProjectStats": {"CountApplications": "应用数\n", "CountMembers": "总成员数\n", "DoneBugCount": "已解决bug数\n", "DoneBugPercent": "bug解决率·\n", "PlanningIterationsCount": "规划中的迭代数\n", "PlanningManHourCount": "总规划工时\n", "RunningIterationsCount": "进行中的迭代数\n", "TotalApplicationsCount": "new states\n总应用数\n", "TotalBugCount": "总bug数\n", "TotalIterationsCount": "总迭代数\n", "TotalManHourCount": "总预计工时\n", "TotalMembersCount": "总成员数\n", "UsedManHourCount": "总已记录工时\n"}, "ProjectUpdateBody": {"ClusterConfig": "项目各环境集群配置\n", "CpuQuota": "+required 单位: c\n", "ID": "路径上有可以不传\n", "MemQuota": "+required 单位: GB\n", "RollbackConfig": "项目回滚点配置\n"}, "PublishItemStatisticsDetailRequest": {"EndTime": "EndTime 截止时间\n"}, "PublishItemStatisticsDetailResponse": {"ActiveUsers": "ActiveUsers 活跃用户\n", "ActiveUsersGrowth": "ActiveUsersGrowth 活跃用户占比\n", "Key": "Key 版本、渠道信息\n", "Launches": "Launches 启动次数\n", "NewUsers": "NewUsers 新增用户\n", "TotalUsers": "totalUsers 截止今日累计用户\n", "TotalUsersGrowth": "TotalUsersGrowth 截止今日累计用户占比\n", "UpgradeUser": "UpgradeUser 升级用户\n"}, "PublishItemStatisticsErrListResponse": {"AffectUsers": "AffectUsers 影响用户数\n", "AppVersion": "AppVersion 版本信息\n", "ErrSummary": "errSummary 错误摘要\n", "TimeOfFirst": "TimeOfFirst 首次发生时间\n", "TimeOfRecent": "TimeOfRecent 最近发生时间\n", "TotalErr": "TotalErr 累计错误计数\n"}, "PublishItemStatisticsErrTrendResponse": {"AffectUsers": "AffectUsers 影响用户数\n", "AffectUsersProportion": "AffectUsersProportion 影响用户占比\n", "AffectUsersProportionGrowth": "AffectUsersProportionGrowth 影响用户占比同比增长率\n", "CrashRate": "CrashRate 崩溃率\n", "CrashRateGrowth": "CrashRateGrowth 崩溃率同比增长率\n", "CrashTimes": "CrashTimes 崩溃次数\n"}, "PublishItemStatisticsTrendResponse": {"MonthTotalActiveUsers": "MonthTotalActiveUsers 30日总活跃用户\n", "MonthTotalActiveUsersGrowth": "MonthTotalActiveUsersGrowth 30日总活跃用户同比增长率\n", "SevenDayAvgActiveUsers": "SevenDayAvgActiveUsers 七日平均活跃用户\n", "SevenDayAvgActiveUsersGrowth": "SevenDayAvgActiveUsersGrowth 七日平均活跃用户同比增长率\n", "SevenDayAvgDuration": "SevenDayAvgDuration 七日平均使用时长\n", "SevenDayAvgDurationGrowth": "SevenDayAvgDurationGrowth 七日平均使用时长同比增长率\n", "SevenDayAvgNewUsers": "SevenDayAvgNewUsers 七日平均新用户\n", "SevenDayAvgNewUsersGrowth": "SevenDayAvgNewUsersGrowth 七日平均新用户同比增长率\n", "SevenDayAvgNewUsersRetention": "SevenDayAvgNewUsersRetention 七日平均新用户次日留存率\n", "SevenDayAvgNewUsersRetentionGrowth": "SevenDayAvgNewUsersRetentionGrowth 七日平均新用户次日留存率同比增长率\n", "SevenDayTotalActiveUsers": "SevenDayTotalActiveUsers 七日总活跃用户\n", "SevenDayTotalActiveUsersGrowth": "SevenDayTotalActiveUsersGrowth 七日总活跃用户同比增长率\n", "TotalCrashRate": "TotalCrashRate 总崩溃率\n", "TotalUsers": "TotalUsers 总用户数\n"}, "PublisherListRequest": {"Joined": "是否只展示已加入的 Publisher\n", "Query": "对Publisher名进行like查询\n"}, "PwdSecurityConfig": {"CaptchaChallengeNumber": "密码错误弹出图片验证码次数\n", "ContinuousPwdErrorNumber": "连续密码错误次数\n", "MaxPwdErrorNumber": "24小时内累计密码错误次数\n", "ResetPassWordPeriod": "强制重密码周期,单位:月\n"}, "QueryNotifyGroupRequest": {"Names": "通知组名字\n"}, "RecordsRequest": {"ClusterName": "optional\n", "PipelineIDs": "optional\n", "RecordIDs": "optional\n", "RecordType": "enum: addNodes, setLabels\n多个值为'或'关系\noptional\n", "Status": "enum: success, failed, processing\n多个值为'或'关系\noptional\n", "UserIDs": "optional\n"}, "RegistryManifestsRemoveResponseData": {"Failed": "删除元数据失败的镜像列表和失败原因\n", "Succeed": "删除元数据成功的镜像列表\n"}, "ReleaseCreateRequest": {"Addon": "Addon addon注册时，release包含dice.yml与addon.yml，选填\n", "ApplicationID": "ApplicationID 应用标识符，描述release所属应用，选填\n", "ApplicationName": "ApplicationName 应用名称，描述release所属应用，选填\n", "ClusterName": "ClusterName 集群名称，描述release所属集群，最大长度80，选填\n", "CrossCluster": "CrossCluster 跨集群\n", "Desc": "Desc 详细描述此release功能, 选填\n", "Dice": "Dice 资源类型为diceyml时, 存储dice.yml内容, 选填\n", "Labels": "Labels 用于release分类，描述release类别，map类型, 最大长度1000, 选填\n", "OrgID": "OrgID 企业标识符，描述release所属企业，选填\n", "ProjectID": "ProjectID 项目标识符，描述release所属项目，选填\n", "ProjectName": "ProjectName 项目名称，描述release所属项目，选填\n", "ReleaseName": "ReleaseName 任意字符串，便于用户识别，最大长度255，必填\n", "Resources": "Resources release包含的资源，包含类型、名称、资源存储路径, 为兼容现有diceyml，先选填\n", "UserID": "UserID 用户标识符, 描述release所属用户，最大长度50，选填\n", "Version": "Version 存储release版本信息, 同一企业同一项目同一应用下唯一，最大长度100，选填\n"}, "ReleaseGetResponseData": {"ApplicationID": "应用Id\n", "ApplicationName": "应用Name\n", "ClusterName": "集群名称\n", "CrossCluster": "CrossCluster 是否可以跨集群\n", "OrgID": "企业标识\n", "ProjectID": "项目Id\n", "ProjectName": "项目Name\n", "Reference": "当前被部署次数\n", "UserID": "操作用户Id\n"}, "ReleaseListRequest": {"ApplicationID": "应用Id\n", "Branch": "分支名\n", "Cluster": "集群名称\n", "CrossCluster": "跨集群\n", "CrossClusterOrSpecifyCluster": "跨集群或指定集群\n", "EndTime": "结束时间,ms\n", "IsVersion": "只列出有 version 的 release\n", "PageNum": "当前页号，默认值1\n", "PageSize": "分页大小,默认值20\n", "ProjectID": "项目ID\n", "Query": "查询参数，releaseId/releaseName/version\n", "ReleaseName": "release 名字精确匹配\n", "StartTime": "开始时间, ms\n"}, "ReleaseListResponseData": {"Total": "release总数，用于分页\n"}, "ReleaseNameListRequest": {"ApplicationID": "应用Id\n"}, "ReleaseResource": {"Name": "资源名称\n", "Type": "资源类型\n", "URL": "资源URL, 可wget获取\n"}, "ReleaseUpdateRequestData": {"ApplicationID": "应用Id\n", "OrgID": "企业标识\n", "ProjectID": "项目Id\n"}, "ResourceReferenceData": {"AddonReference": "addon引用数\n", "ServiceReference": "服务引用数\n"}, "ResourceReferenceResp": {"Data": "key 为 projectID\n"}, "Resources": {"Cpu": "cpu sharing\n", "Disk": "disk usage\n", "Mem": "memory usage\n"}, "RmNodesRequest": {"Force": "skip addon-exist-on-nodes check\n"}, "Role": {"Permissions": "权限列表\n", "Scope": "范围\n"}, "RoleChangeBody": {"TargetId": "目标id, 对应的applicationId, projectId, orgId\n", "TargetType": "目标类型 APPLICATION,PROJECT,ORG\n"}, "RoleChangeRequest": {"Role": "用户角色\n"}, "RouteOptions": {"Annotations": "参考: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/\n", "EnableTLS": "是否开启TLS，不填时，默认为true\n", "LocationSnippet": "参考: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#configuration-snippet\n", "RewriteHost": "重写转发域名\n", "RewritePath": "重写转发路径\n", "UseRegex": "Path中是否使用了正则\n"}, "RuntimeCreateRequestExtra": {"AddonActions": "for addon actions\n", "ClusterId": "Deprecated\n"}, "RuntimeInspectDTO": {"ModuleErrMsg": "模块发布错误信息\n", "Name": "runtime名称\n", "Status": "状态\n"}, "RuntimeInspectRequest": {"ApplicationID": "应用 ID, idOrName 为 Name 时必传, 为 ID 时不必传\n", "IDOrName": "应用实例 ID / Name\n", "Workspace": "环境, idOrName 为 Name 时必传, 为 ID 时不必传\n"}, "RuntimeReleaseCreateRequest": {"ApplicationID": "应用ID\n", "ProjectID": "项目ID\n", "ReleaseID": "制品ID\n", "Workspace": "环境\n"}, "RuntimeServiceRequest": {"AppID": "AppID 应用ID\n", "AppName": "AppName 应用名称\n", "ClusterName": "CluserName 集群名称\n", "OrgID": "OrgID 企业ID\n", "ProjectID": "ProjectID 项目ID\n", "ReleaseID": "ReleaseId\n", "RuntimeID": "RuntimeID runtimeID\n", "RuntimeName": "RuntimeName runtime名称\n", "ServiceGroupName": "ServiceGroupName\n", "ServiceGroupNamespace": "ServiceGroupNamespace\n", "Services": "Services 服务组成的列表\n", "UseApigw": "UseApigw 是否通过addon依赖了api网关\n", "Workspace": "Workspace 所属环境\n"}, "ScheduleInfo": {"ExclusiveLikes": "不与 \"any\" 标签共存的 Like\n", "Flag": "currently only for \"any\" label\n", "HostUnique": "服务（包括JOB）打散在不同 host\nHostUnique: 是否启用 host 打散\nHostUniqueInfo: service 分组\n", "InclusiveLikes": "元素是或集合，组合到一条约束语句中\n", "IsPlatform": "是否需要调度到 `平台` 所属机器\n", "IsUnLocked": "是否需要调度到 `非 locked` 机器\n", "LikePrefixs": "调度喜好对应的以该值为前缀的群体\n", "Likes": "调度喜好对应的个体\n", "Location": "Location 允许调度目的节点类型列表\n\ne.g.\n\nLocation: map[string]      interface{}\n          map[servicename] diceyml.Selector\n\nTODO: 目前 map value 是 interface{} 是因为 apistructs 没有 import diceyml，\n      需要把 diceyml 结构体移动到 apistructs\n", "SpecificHost": "指定 host, 列表中的host为‘或’关系\n"}, "ScheduleInfo2": {"HasHostUnique": "服务（包括JOB）打散在不同 host\nHasHostUnique: 是否启用 host 打散\nHostUnique: service 分组\n", "HasOrg": "HasOrg 表示 Org 字段是否有意义\n1. '集群配置' 中未开启: HasOrg = false\n2. '集群配置' 中开启，`LabelInfo.Label` 中没有 `labelconfig.ORG_KEY` label \u0026 selectors 中没有 `org`:\n     HasOrg = false\n3. '集群配置' 中开启，`LabelInfo.Label` 中存在 `labelconfig.ORG_KEY` label | selectors 中存在 `org`:\n     HasOrg = true, Org = \"\u003corgname\u003e\"\n", "HasProject": "Project label\n=DEPRECATED= k8s 中忽略该字段\n", "HasWorkSpace": "HasWorkSpace 表示 WorkSpace 字段是否有意义\n1. HasOrg = false\t\t: HasWorkSpace = false\n2. '集群配置' 中未开启\t: HasWorkSpace = false\n3. '集群配置' 中开启，`LabelInfo.Label` 中没有 `labelconfig.WORKSPACE_KEY` label \u0026 selectors 中没有 `org`  ：\n     HasWorkSpace = false\n4. '集群配置' 中开启，`LabelInfo.Label` 中存在 `labelconfig.ORG_KEY` label | selectors 中存在 `org`:\n     HasWorkSpace = true, WorkSpace = [\"\u003cworkspace\u003e\", ...]\n", "IsPlatform": "是否需要调度到 `平台` 所属机器\n", "IsUnLocked": "是否需要调度到 `非 locked` 机器\n总是 true\n", "Location": "Location 允许调度目的节点类型列表\n\ne.g.\n\nLocation: map[string]      interface{}\n          map[servicename] diceyml.Selector\n\nTODO: 目前 map value 是 interface{} 是因为 apistructs 没有 import diceyml，\n      需要把 diceyml 结构体移动到 apistructs\n", "PreferJob": "PreferJob\nk8s      忽略该字段\nmarathon 中生成的约束为 job | any\n", "PreferPack": "PreferPack\nk8s      忽略该字段\nmarathon 中生成的约束为 pack | any\n", "PreferStateful": "PreferStateful\nk8s      忽略该字段\nmarathon 中生成的约束为 stateful | any\n", "PreferStateless": "PreferStateless\nk8s      忽略该字段\nmarathon 中生成的约束为 stateless | any\n", "SpecificHost": "指定 host, 列表中的host为 ‘或’ 关系\n", "WorkSpaces": "WorkSpaces 列表中的 workspace 为 `或` 关系\n[a, b, c] =\u003e a | b | c\n"}, "ScheduleLabelListData": {"Labels": "map-key: label name\nmap-value: is this label a prefix?\n"}, "ScheduleLabelSetRequest": {"Tags": "对于 dcos 的 tag, 由于只有 key, 则 tag 中的 value 都为空\n"}, "Scope": {"ID": "范围对应的实例 ID (orgID, projectID, applicationID ...)\n比如 type == \"org\" 时, id 即为 orgID\n", "Type": "范围类型\n可选值: sys, org, project, app\n"}, "ScopeResource": {"Action": "Action Create/Update/Delete/\n", "Resource": "Resource 资源类型， eg: ticket/release\n", "ResourceRole": "resource 角色: Creator, Assignee\n"}, "ScriptInfo": {"ScriptBlackList": "脚本名逗号分隔，ALL代表终止全部脚本\n"}, "Service": {"Binds": "disk bind (mount) configuration, hostPath only\n", "Cmd": "docker's CMD\n", "Depends": "list of service names depends by this service, used for dependency scheduling\n", "DeploymentLabels": "deploymentLabels 会转化到 pod spec label 中, dcos 忽略此字段\n", "Env": "environment variables inject into container\n", "HealthCheck": "health check\n", "Hosts": "hosts append into /etc/hosts\n", "Image": "docker's image url\n", "ImagePassword": "docker's image password\n", "ImageUsername": "docker's image username\n", "InstanceInfos": "instance info, only for display\nmarathon 中对应一个task, k8s中对应一个pod\n", "Labels": "labels for extension and some tags\n", "MeshEnable": "service mesh 的服务级别开关\n", "Name": "unique name between services in one Dice (ServiceGroup)\n", "Namespace": "namespace of service, equal to the namespace in Dice\n", "Ports": "port list user-defined, we export these ports on our VIP\n", "ProxyIp": "only exists if serviceDiscoveryKind is PROXY\ncan not modify directly, assigned by dice\n", "ProxyPorts": "only exists if serviceDiscoveryKind is PROXY\ncan not modify directly, assigned by dice\n", "PublicIp": "TODO: refactor it, currently only work with label X_ENABLE_PUBLIC_IP=true\n", "Resources": "resources like cpu, mem, disk\n", "Scale": "instances of containers should running\n", "Selectors": "Selectors see also diceyml.Service.Selectors\n\nTODO: 把 ServiceGroup structure  移动到 scheduler 内部，Selectors 类型换为 diceyml.Selectors\n", "ShortVIP": "ShortVIP 短域名，为解决 DCOS, K8S等短域名不一致问题\n", "TrafficSecurity": "对应 istio 的流量加密策略\n", "Vip": "virtual ip\ncan not modify directly, assigned by dice\n", "Volumes": "Volumes intends to replace Binds\n", "WorkLoad": "WorkLoad indicates the type of service，\nsupport Kubernetes workload DaemonSet(Per-Node), Statefulset and Deployment\n"}, "ServiceBind": {"Persistent": "TODO: refactor it, currently just copy the marathon struct\n"}, "ServiceGroup": {"ClusterName": "substitute for \"Executor\" field\n", "CreatedTime": "runtime create time\n", "Executor": "executor for scheduling (e.g. marathon)\n", "Extra": "current usage for Extra:\n1. record last restart time, to implement RESTART api through PUT api\n", "LastModifiedTime": "last modified (update) time\n", "ScheduleInfo": "根据集群配置以及 label 所计算出的调度规则\nTODO: DEPRECATED\n", "ScheduleInfo2": "将会代替 ScheduleInfo\n", "Version": "version to tracing changes (create, update, etc.)\n"}, "ServiceGroupCreateV2Data": {"Version": "目前没用\n"}, "ServiceGroupCreateV2Request": {"ClusterName": "DiceYml              json.RawMessage   `json:\"diceyml\"`\n", "GroupLabels": "DEPRECATED, 放在 diceyml.meta 中\n", "ServiceDiscoveryMode": "DEPRECATED, 放在 diceyml.meta 中\n", "Volumes": "DEPRECATED\nmap[servicename]volumeinfo\n"}, "ServiceGroupPrecheckData": {"Nodes": "key: servicename\n"}, "ServiceItem": {"InnerAddress": "InnerAddress 服务内部地址\n", "ServiceName": "ServiceName 服务名称\n"}, "ServicePort": {"Port": "Port is port for service connection\n", "Protocol": "Protocol support kubernetes orn Protocol Type. It\ncontains ProtocolTCP， ProtocolUDP，ProtocolSCTP\n"}, "StatusDesc": {"LastMessage": "LastMessage 描述状态的额外信息\n", "Status": "Status 描述状态\n", "UnScheduledReasons": "[DEPRECATED] UnScheduledReasons 描述具体资源不足的信息\n"}, "Target": {"Secret": "目前只有钉钉用\n"}, "TestCallBackRequest": {"Totals": "Totals is the aggregated results of all tests.\n"}, "TestCaseListRequest": {"AllowEmptyTestSetIDs": "AllowEmptyTestSetIDs 是否允许 testSetIDs 为空，默认为 false\n"}, "TestCasePagingRequest": {"Labels": "TODO 用例类型\n", "OrderFields": "order by field\n", "PageNo": "分页参数\n", "ProjectID": "项目 ID，目前必填，因为测试用例的 testSetID 可以为 0，若无 projectID 只有 testSetID，会查到别的 project\n", "TimestampSecUpdatedAtBegin": "更新时间，外部传参使用时间戳\n", "UpdatedAtBeginInclude": "更新时间，内部使用直接赋值\n"}, "TestPlanCaseRelCreateRequest": {"TestSetIDs": "若 TestSetIDs 不为空，则添加测试集下所有测试用例到测试集下，与 TestCaseIDs 取合集\n"}, "TestPlanCaseRelPagingRequest": {"OrderByPriorityAsc": "order by field\n", "TimestampSecUpdatedAtBegin": "更新时间，外部传参使用时间戳\n", "UpdatedAtBeginInclude": "更新时间，内部使用直接赋值\n"}, "TestPlanCreateRequest": {"IsAutoTest": "是否是自动化测试计划\n"}, "TestPlanPagingRequest": {"OwnerIDs": "member about\n", "PageNo": "+optional default 1\n", "PageSize": "+optional default 10\n"}, "TestPlanTestCaseRelDeleteRequest": {"ProjectID": "+required\n", "TestPlanID": "+required\n"}, "TestPlanV2PagingRequest": {"IDs": "ids\n", "PageNo": "+optional default 1\n", "PageSize": "+optional default 20\n"}, "TestSet": {"CreatorID": "创建人ID\n", "Directory": "显示的目录地址\n", "ID": "测试集ID\n", "Name": "测试集名称\n", "Order": "排序\n", "ParentID": "父测试集ID\n", "ProjectID": "项目 ID\n", "Recycled": "是否回收\n", "UpdaterID": "更新人ID\n"}, "TestSetCreateRequest": {"Name": "名称\n", "ParentID": "父测试集ID\n", "ProjectID": "项目ID\n"}, "TestSetListRequest": {"NoSubTestSets": "是否不递归，默认 false，即默认递归\n", "ParentID": "父测试集 ID\n", "ProjectID": "项目 ID\n", "Recycled": "是否回收\n", "TestSetIDs": "指定 id 列表\n"}, "TestSetRecycleRequest": {"IsRoot": "IsRoot 表示递归回收测试集时是否是最外层的根测试集\n如果是根测试集，且 parentID != 0，回收时需要将 parentID 置为 0，否则在回收站中无法找到\n"}, "TestSetUpdateRequest": {"Name": "待更新项\n", "TestSetID": "基础信息\n"}, "TestSetWithPlanCaseRels": {"TestCaseCountWithoutFilter": "当前测试集下所有测试用例的个数，不考虑过滤条件；\n场景：分页查询，当前页只能显示部分用例，批量删除这部分用例后，前端需要根据这个参数值判断当前测试集下是否还有用例。\n     若已全部删除，则前端删除目录栏里的当前目录。\n"}, "TestSuite": {"Name": "Name is a descriptor given to the suite.\n", "Package": "Package is an additional descriptor for the hierarchy of the suite.\n", "Properties": "Properties is a mapping of key-value pairs that were available when the\ntests were run.\n", "SystemErr": "SystemErr is textual test error output for the suite. Usually output that is\nwritten to stderr.\n", "SystemOut": "SystemOut is textual test output for the suite. Usually output that is\nwritten to stdout.\n", "Tests": "Tests is an ordered collection of tests with associated results.\n", "Totals": "Totals is the aggregated results of all tests.\n"}, "Ticket": {"ClosedAt": "关闭时间\n", "Content": "工单内容\n", "Count": "累计告警次数\n", "CreatedAt": "创建时间\n", "Creator": "工单创建者ID\n", "Key": "告警工单 key，选填， 用于定位告警类工单\n", "Label": "标签\n", "LastComment": "工单最新评论，仅主动监控使用\n", "LastOperator": "工单最近操作者ID\n", "Metric": "告警指标，告警使用，其他类型工单不传\n", "Priority": "工单优先级，可选值: high/medium/low\n", "Status": "工单状态，可选值: open/closed\n", "TargetID": "工单目标ID\n", "TargetType": "工单目标类型，可选值: cluster/project/application\n", "TicketID": "工单ID\n", "Title": "工单标题\n", "TriggeredAt": "触发时间\n", "Type": "工单类型，可选值: bug/vulnerability/codeSmell/task\n", "UpdatedAt": "更新时间\n"}, "TicketCloseResponse": {"Data": "工单ID\n"}, "TicketCreateRequest": {"ClosedAt": "告警恢复时间\n", "Content": "工单内容\n", "Key": "告警工单使用，作为唯一 key 定位工单\n", "Label": "标签\n", "Metric": "告警指标，告警使用，其他类型工单不传\n", "OrgID": "企业ID\n", "Priority": "工单优先级，可选值: high/medium/low\n", "TargetID": "工单目标ID\n", "TargetType": "工单目标类型，可选值: machine/addon/project/application\n", "Title": "工单标题\n", "TriggeredAt": "触发时间\n", "Type": "工单类型 可选值: task/bug/vulnerability/codeSmell/machine/component/addon/trace/glance/exception\n", "UserID": "用户ID\n"}, "TicketCreateResponse": {"Data": "工单ID\n"}, "TicketDeleteResponse": {"Data": "工单ID\n"}, "TicketListRequest": {"Comment": "是否包含工单最新评论，默认false\n", "EndTime": "截止时间戳(ms)，选填 默认为当前时间\n", "Key": "告警工单 key，选填， 用于定位告警类工单\n", "Metric": "告警维度，选填(仅供告警类工单使用) eg: cpu/mem/load\n", "MetricID": "告警维度取值, 选填\n", "OrgID": "企业ID, 选填，集群类告警时使用\n", "PageNo": "页号, 默认值:1\n", "PageSize": "分页大小, 默认值20\n", "Priority": "工单优先级，选填 可选值: high/medium/low\n", "Q": "查询参数，按title/label模糊匹配\n", "StartTime": "起始时间戳(ms)，选填\n", "Status": "工单状态，选填 可选值: open/closed\n", "TargetID": "工单关联目标ID，选填\n", "TargetType": "工单关联目标类型, 选填 可选值: cluster/project/application\n", "Type": "工单类型，选填 可选值: task/bug/vulnerability/codeSmell/machine/component/addon/trace/glance/exception\n"}, "TicketReopenResponse": {"Data": "工单ID\n"}, "TicketUpdateRequestBody": {"Content": "工单内容\n", "Priority": "工单优先级，可选值: high/medium/low\n", "Title": "工单标题\n", "Type": "工单类型，可选值: task/bug/vulnerability/codeSmell/machine/component/addon/trace/glance/exception\n"}, "TicketUpdateResponse": {"Data": "工单ID\n"}, "UnifiedFileTreeNodeFuzzySearchRequest": {"FromPinode": "从哪个父节点开始搜索\n", "PrefixFuzzy": "fuzzy search\n", "Recursive": "是否需要递归，若不递归，则只返回当前层\n"}, "UpdateCustomAddonRequest": {"Body": "更新custom addon请求体\n"}, "User": {"ThirdPart": "三方用户如wechat,qq等\n", "ThirdUid": "三方用户的id\n"}, "UserListRequest": {"Plaintext": "用户信息是否明文\n", "Query": "查询关键字，可根据用户名/手机号/邮箱模糊匹配\n", "UserIDs": "支持批量查询，传参形式: userID=xxx\u0026userID=yyy\n"}, "UserProfile": {"Authorizes": "用户权限列表\n", "Roles": "用户角色列表\n"}, "UserRole": {"RoleKey": "角色key\n"}, "VClusterCreateRequest": {"ClusterID": "物理集群Id\n", "ClusterName": "物理集群名称\n", "Name": "集群名称\n", "OrgID": "集群对应组织Id\n", "OrgName": "集群对应组织名称\n", "Owner": "集群拥有者\n"}, "VClusterCreateResponse": {"Data": "集群Id\n"}, "VClusterFetchResponseData": {"ClusterID": "物理集群Id\n", "ClusterName": "物理集群名称\n", "Name": "集群名称\n", "OrgID": "集群对应组织ID\n", "OrgName": "集群对应组织名称\n", "Owner": "集群拥有者\n", "UUID": "集群uuid\n"}, "ValidBranch": {"ArtifactWorkspace": "制品可部署的环境\n", "Workspace": "通过分支创建的流水线环境\n"}, "Volume": {"ContainerPath": "挂载到容器中的卷路径\n", "ID": "volume ID\n", "Size": "单位 G\n", "Storage": "TODO: k8s.go 需要这个字段，现在对于k8s先不使用其插件中实现的volume相关实现（现在也没有用的地方）\nk8s plugin 重构的时候才去实现 k8s 特定的 volume 逻辑\n", "VolumePath": "由volume driver来填 volume 所在地址\n对于 localvolume: hostpath\n对于 nasvolume: nas网盘地址(/netdata/xxx/...)\n"}, "VolumeCreateRequest": {"Size": "单位 G\n"}, "VolumeInfo": {"ID": "volume ID, 可能是 uuid 也可能是 unique name\n"}, "WebhookDeleteRequest": {"ID": "webhook ID\n"}, "WebhookInspectRequest": {"ID": "所查询的 webhook ID\n"}, "WebhookListRequest": {"ApplicationID": "列出 orgid \u0026 projectID \u0026 applicationID \u0026 env 下的 webhook\n", "Env": "列出 orgid \u0026 projectID \u0026 applicationID \u0026 env 下的 webhook, env格式：test,prod,dev\n", "OrgID": "列出 orgid \u0026 projectID \u0026 applicationID \u0026 env 下的 webhook\n", "ProjectID": "列出 orgid \u0026 projectID \u0026 applicationID \u0026 env 下的 webhook\n"}, "WebhookPingRequest": {"ID": "webhook ID\n"}, "WebhookUpdateRequest": {"ID": "webhook ID\n"}, "WebhookUpdateRequestBody": {"Active": "是否激活，如果没有该参数，默认为false\n", "AddEvents": "从 webhook event 列表中增加\n", "Events": "全量更新这个 webhook 关心的 event 列表\n", "RemoveEvents": "从 webhook event 列表中删除\n", "URL": "该 webhook 对应的 URL， 所关心事件触发后会POST到该URL\n"}, "WidgetResponse": {"Datas": "字段数据列表\n", "Name": "图形名称\n", "Names": "字段名列表\n", "Titles": "图形标题\n"}}

func (CloudClusterContainerInfo) Desc_CloudClusterContainerInfo(s string) string {
	if structDescMap["CloudClusterContainerInfo"] == nil {
		return ""
	}
	return structDescMap["CloudClusterContainerInfo"][s]
}

func (RoleChangeBody) Desc_RoleChangeBody(s string) string {
	if structDescMap["RoleChangeBody"] == nil {
		return ""
	}
	return structDescMap["RoleChangeBody"][s]
}

func (AddonNameResponse) Desc_AddonNameResponse(s string) string {
	if structDescMap["AddonNameResponse"] == nil {
		return ""
	}
	return structDescMap["AddonNameResponse"][s]
}

func (Component) Desc_Component(s string) string {
	if structDescMap["Component"] == nil {
		return ""
	}
	return structDescMap["Component"][s]
}

func (Audit) Desc_Audit(s string) string {
	if structDescMap["Audit"] == nil {
		return ""
	}
	return structDescMap["Audit"][s]
}

func (Target) Desc_Target(s string) string {
	if structDescMap["Target"] == nil {
		return ""
	}
	return structDescMap["Target"][s]
}

func (CloudResourceMysqlListDatabaseRequest) Desc_CloudResourceMysqlListDatabaseRequest(s string) string {
	if structDescMap["CloudResourceMysqlListDatabaseRequest"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlListDatabaseRequest"][s]
}

func (RoleChangeRequest) Desc_RoleChangeRequest(s string) string {
	if structDescMap["RoleChangeRequest"] == nil {
		return ""
	}
	return structDescMap["RoleChangeRequest"][s]
}

func (EnvConfig) Desc_EnvConfig(s string) string {
	if structDescMap["EnvConfig"] == nil {
		return ""
	}
	return structDescMap["EnvConfig"][s]
}

func (TicketCloseResponse) Desc_TicketCloseResponse(s string) string {
	if structDescMap["TicketCloseResponse"] == nil {
		return ""
	}
	return structDescMap["TicketCloseResponse"][s]
}

func (MiddlewareListRequest) Desc_MiddlewareListRequest(s string) string {
	if structDescMap["MiddlewareListRequest"] == nil {
		return ""
	}
	return structDescMap["MiddlewareListRequest"][s]
}

func (ActionCreateRequest) Desc_ActionCreateRequest(s string) string {
	if structDescMap["ActionCreateRequest"] == nil {
		return ""
	}
	return structDescMap["ActionCreateRequest"][s]
}

func (JobVolume) Desc_JobVolume(s string) string {
	if structDescMap["JobVolume"] == nil {
		return ""
	}
	return structDescMap["JobVolume"][s]
}

func (IterationCreateRequest) Desc_IterationCreateRequest(s string) string {
	if structDescMap["IterationCreateRequest"] == nil {
		return ""
	}
	return structDescMap["IterationCreateRequest"][s]
}

func (RecordsRequest) Desc_RecordsRequest(s string) string {
	if structDescMap["RecordsRequest"] == nil {
		return ""
	}
	return structDescMap["RecordsRequest"][s]
}

func (Volume) Desc_Volume(s string) string {
	if structDescMap["Volume"] == nil {
		return ""
	}
	return structDescMap["Volume"][s]
}

func (OrgCreateRequest) Desc_OrgCreateRequest(s string) string {
	if structDescMap["OrgCreateRequest"] == nil {
		return ""
	}
	return structDescMap["OrgCreateRequest"][s]
}

func (IssueListRequest) Desc_IssueListRequest(s string) string {
	if structDescMap["IssueListRequest"] == nil {
		return ""
	}
	return structDescMap["IssueListRequest"][s]
}

func (AddonTenantCreateRequest) Desc_AddonTenantCreateRequest(s string) string {
	if structDescMap["AddonTenantCreateRequest"] == nil {
		return ""
	}
	return structDescMap["AddonTenantCreateRequest"][s]
}

func (GittarCreateTagRequest) Desc_GittarCreateTagRequest(s string) string {
	if structDescMap["GittarCreateTagRequest"] == nil {
		return ""
	}
	return structDescMap["GittarCreateTagRequest"][s]
}

func (AuditSetCleanCronRequest) Desc_AuditSetCleanCronRequest(s string) string {
	if structDescMap["AuditSetCleanCronRequest"] == nil {
		return ""
	}
	return structDescMap["AuditSetCleanCronRequest"][s]
}

func (TestPlanV2PagingRequest) Desc_TestPlanV2PagingRequest(s string) string {
	if structDescMap["TestPlanV2PagingRequest"] == nil {
		return ""
	}
	return structDescMap["TestPlanV2PagingRequest"][s]
}

func (Issue) Desc_Issue(s string) string {
	if structDescMap["Issue"] == nil {
		return ""
	}
	return structDescMap["Issue"][s]
}

func (CertificateListRequest) Desc_CertificateListRequest(s string) string {
	if structDescMap["CertificateListRequest"] == nil {
		return ""
	}
	return structDescMap["CertificateListRequest"][s]
}

func (TestPlanPagingRequest) Desc_TestPlanPagingRequest(s string) string {
	if structDescMap["TestPlanPagingRequest"] == nil {
		return ""
	}
	return structDescMap["TestPlanPagingRequest"][s]
}

func (AddonAvailableRequest) Desc_AddonAvailableRequest(s string) string {
	if structDescMap["AddonAvailableRequest"] == nil {
		return ""
	}
	return structDescMap["AddonAvailableRequest"][s]
}

func (PublishItemStatisticsErrListResponse) Desc_PublishItemStatisticsErrListResponse(s string) string {
	if structDescMap["PublishItemStatisticsErrListResponse"] == nil {
		return ""
	}
	return structDescMap["PublishItemStatisticsErrListResponse"][s]
}

func (MemberRemoveRequest) Desc_MemberRemoveRequest(s string) string {
	if structDescMap["MemberRemoveRequest"] == nil {
		return ""
	}
	return structDescMap["MemberRemoveRequest"][s]
}

func (PipelineTaskSnippetDetail) Desc_PipelineTaskSnippetDetail(s string) string {
	if structDescMap["PipelineTaskSnippetDetail"] == nil {
		return ""
	}
	return structDescMap["PipelineTaskSnippetDetail"][s]
}

func (ComponentProtocolResponseData) Desc_ComponentProtocolResponseData(s string) string {
	if structDescMap["ComponentProtocolResponseData"] == nil {
		return ""
	}
	return structDescMap["ComponentProtocolResponseData"][s]
}

func (ExtensionCreateRequest) Desc_ExtensionCreateRequest(s string) string {
	if structDescMap["ExtensionCreateRequest"] == nil {
		return ""
	}
	return structDescMap["ExtensionCreateRequest"][s]
}

func (LibReferenceListRequest) Desc_LibReferenceListRequest(s string) string {
	if structDescMap["LibReferenceListRequest"] == nil {
		return ""
	}
	return structDescMap["LibReferenceListRequest"][s]
}

func (CreateCloudResourceChargeInfo) Desc_CreateCloudResourceChargeInfo(s string) string {
	if structDescMap["CreateCloudResourceChargeInfo"] == nil {
		return ""
	}
	return structDescMap["CreateCloudResourceChargeInfo"][s]
}

func (TestPlanCaseRelPagingRequest) Desc_TestPlanCaseRelPagingRequest(s string) string {
	if structDescMap["TestPlanCaseRelPagingRequest"] == nil {
		return ""
	}
	return structDescMap["TestPlanCaseRelPagingRequest"][s]
}

func (OneDataAnalysisBussProcsRequest) Desc_OneDataAnalysisBussProcsRequest(s string) string {
	if structDescMap["OneDataAnalysisBussProcsRequest"] == nil {
		return ""
	}
	return structDescMap["OneDataAnalysisBussProcsRequest"][s]
}

func (AddonStrategy) Desc_AddonStrategy(s string) string {
	if structDescMap["AddonStrategy"] == nil {
		return ""
	}
	return structDescMap["AddonStrategy"][s]
}

func (RuntimeInspectRequest) Desc_RuntimeInspectRequest(s string) string {
	if structDescMap["RuntimeInspectRequest"] == nil {
		return ""
	}
	return structDescMap["RuntimeInspectRequest"][s]
}

func (Member) Desc_Member(s string) string {
	if structDescMap["Member"] == nil {
		return ""
	}
	return structDescMap["Member"][s]
}

func (OnsEndpoints) Desc_OnsEndpoints(s string) string {
	if structDescMap["OnsEndpoints"] == nil {
		return ""
	}
	return structDescMap["OnsEndpoints"][s]
}

func (AppCertificateListRequest) Desc_AppCertificateListRequest(s string) string {
	if structDescMap["AppCertificateListRequest"] == nil {
		return ""
	}
	return structDescMap["AppCertificateListRequest"][s]
}

func (WidgetResponse) Desc_WidgetResponse(s string) string {
	if structDescMap["WidgetResponse"] == nil {
		return ""
	}
	return structDescMap["WidgetResponse"][s]
}

func (AddonCreateRequest) Desc_AddonCreateRequest(s string) string {
	if structDescMap["AddonCreateRequest"] == nil {
		return ""
	}
	return structDescMap["AddonCreateRequest"][s]
}

func (PipelineButton) Desc_PipelineButton(s string) string {
	if structDescMap["PipelineButton"] == nil {
		return ""
	}
	return structDescMap["PipelineButton"][s]
}

func (RuntimeInspectDTO) Desc_RuntimeInspectDTO(s string) string {
	if structDescMap["RuntimeInspectDTO"] == nil {
		return ""
	}
	return structDescMap["RuntimeInspectDTO"][s]
}

func (UserRole) Desc_UserRole(s string) string {
	if structDescMap["UserRole"] == nil {
		return ""
	}
	return structDescMap["UserRole"][s]
}

func (TicketDeleteResponse) Desc_TicketDeleteResponse(s string) string {
	if structDescMap["TicketDeleteResponse"] == nil {
		return ""
	}
	return structDescMap["TicketDeleteResponse"][s]
}

func (TicketReopenResponse) Desc_TicketReopenResponse(s string) string {
	if structDescMap["TicketReopenResponse"] == nil {
		return ""
	}
	return structDescMap["TicketReopenResponse"][s]
}

func (AddonDependsRelation) Desc_AddonDependsRelation(s string) string {
	if structDescMap["AddonDependsRelation"] == nil {
		return ""
	}
	return structDescMap["AddonDependsRelation"][s]
}

func (Scope) Desc_Scope(s string) string {
	if structDescMap["Scope"] == nil {
		return ""
	}
	return structDescMap["Scope"][s]
}

func (ExistsMysqlExec) Desc_ExistsMysqlExec(s string) string {
	if structDescMap["ExistsMysqlExec"] == nil {
		return ""
	}
	return structDescMap["ExistsMysqlExec"][s]
}

func (CloudResourceMysqlDBInfo) Desc_CloudResourceMysqlDBInfo(s string) string {
	if structDescMap["CloudResourceMysqlDBInfo"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlDBInfo"][s]
}

func (TestCaseListRequest) Desc_TestCaseListRequest(s string) string {
	if structDescMap["TestCaseListRequest"] == nil {
		return ""
	}
	return structDescMap["TestCaseListRequest"][s]
}

func (MultiLevelStatus) Desc_MultiLevelStatus(s string) string {
	if structDescMap["MultiLevelStatus"] == nil {
		return ""
	}
	return structDescMap["MultiLevelStatus"][s]
}

func (AddonConfigRes) Desc_AddonConfigRes(s string) string {
	if structDescMap["AddonConfigRes"] == nil {
		return ""
	}
	return structDescMap["AddonConfigRes"][s]
}

func (DiffFile) Desc_DiffFile(s string) string {
	if structDescMap["DiffFile"] == nil {
		return ""
	}
	return structDescMap["DiffFile"][s]
}

func (RmNodesRequest) Desc_RmNodesRequest(s string) string {
	if structDescMap["RmNodesRequest"] == nil {
		return ""
	}
	return structDescMap["RmNodesRequest"][s]
}

func (CloudResourceOssDetailInfoData) Desc_CloudResourceOssDetailInfoData(s string) string {
	if structDescMap["CloudResourceOssDetailInfoData"] == nil {
		return ""
	}
	return structDescMap["CloudResourceOssDetailInfoData"][s]
}

func (CreateCloudResourceOnsRequest) Desc_CreateCloudResourceOnsRequest(s string) string {
	if structDescMap["CreateCloudResourceOnsRequest"] == nil {
		return ""
	}
	return structDescMap["CreateCloudResourceOnsRequest"][s]
}

func (CloudAccount) Desc_CloudAccount(s string) string {
	if structDescMap["CloudAccount"] == nil {
		return ""
	}
	return structDescMap["CloudAccount"][s]
}

func (VolumeInfo) Desc_VolumeInfo(s string) string {
	if structDescMap["VolumeInfo"] == nil {
		return ""
	}
	return structDescMap["VolumeInfo"][s]
}

func (HealthCheck) Desc_HealthCheck(s string) string {
	if structDescMap["HealthCheck"] == nil {
		return ""
	}
	return structDescMap["HealthCheck"][s]
}

func (GittarStatsData) Desc_GittarStatsData(s string) string {
	if structDescMap["GittarStatsData"] == nil {
		return ""
	}
	return structDescMap["GittarStatsData"][s]
}

func (CloudResourceOnsGroupInfoRequest) Desc_CloudResourceOnsGroupInfoRequest(s string) string {
	if structDescMap["CloudResourceOnsGroupInfoRequest"] == nil {
		return ""
	}
	return structDescMap["CloudResourceOnsGroupInfoRequest"][s]
}

func (ApplicationDTO) Desc_ApplicationDTO(s string) string {
	if structDescMap["ApplicationDTO"] == nil {
		return ""
	}
	return structDescMap["ApplicationDTO"][s]
}

func (PipelineDetailDTO) Desc_PipelineDetailDTO(s string) string {
	if structDescMap["PipelineDetailDTO"] == nil {
		return ""
	}
	return structDescMap["PipelineDetailDTO"][s]
}

func (CloudClusterRequest) Desc_CloudClusterRequest(s string) string {
	if structDescMap["CloudClusterRequest"] == nil {
		return ""
	}
	return structDescMap["CloudClusterRequest"][s]
}

func (OneDataAnalysisDimRequest) Desc_OneDataAnalysisDimRequest(s string) string {
	if structDescMap["OneDataAnalysisDimRequest"] == nil {
		return ""
	}
	return structDescMap["OneDataAnalysisDimRequest"][s]
}

func (GittarCommitsRequest) Desc_GittarCommitsRequest(s string) string {
	if structDescMap["GittarCommitsRequest"] == nil {
		return ""
	}
	return structDescMap["GittarCommitsRequest"][s]
}

func (ServiceGroupPrecheckData) Desc_ServiceGroupPrecheckData(s string) string {
	if structDescMap["ServiceGroupPrecheckData"] == nil {
		return ""
	}
	return structDescMap["ServiceGroupPrecheckData"][s]
}

func (CommentCreateResponse) Desc_CommentCreateResponse(s string) string {
	if structDescMap["CommentCreateResponse"] == nil {
		return ""
	}
	return structDescMap["CommentCreateResponse"][s]
}

func (ProjectResourceResponse) Desc_ProjectResourceResponse(s string) string {
	if structDescMap["ProjectResourceResponse"] == nil {
		return ""
	}
	return structDescMap["ProjectResourceResponse"][s]
}

func (ApplicationUpdateRequestBody) Desc_ApplicationUpdateRequestBody(s string) string {
	if structDescMap["ApplicationUpdateRequestBody"] == nil {
		return ""
	}
	return structDescMap["ApplicationUpdateRequestBody"][s]
}

func (ApplicationWorkspace) Desc_ApplicationWorkspace(s string) string {
	if structDescMap["ApplicationWorkspace"] == nil {
		return ""
	}
	return structDescMap["ApplicationWorkspace"][s]
}

func (AddonExtension) Desc_AddonExtension(s string) string {
	if structDescMap["AddonExtension"] == nil {
		return ""
	}
	return structDescMap["AddonExtension"][s]
}

func (AutoTestSpace) Desc_AutoTestSpace(s string) string {
	if structDescMap["AutoTestSpace"] == nil {
		return ""
	}
	return structDescMap["AutoTestSpace"][s]
}

func (IssuePagingRequest) Desc_IssuePagingRequest(s string) string {
	if structDescMap["IssuePagingRequest"] == nil {
		return ""
	}
	return structDescMap["IssuePagingRequest"][s]
}

func (ScheduleLabelSetRequest) Desc_ScheduleLabelSetRequest(s string) string {
	if structDescMap["ScheduleLabelSetRequest"] == nil {
		return ""
	}
	return structDescMap["ScheduleLabelSetRequest"][s]
}

func (ScheduleInfo) Desc_ScheduleInfo(s string) string {
	if structDescMap["ScheduleInfo"] == nil {
		return ""
	}
	return structDescMap["ScheduleInfo"][s]
}

func (CommentUpdateRequestBody) Desc_CommentUpdateRequestBody(s string) string {
	if structDescMap["CommentUpdateRequestBody"] == nil {
		return ""
	}
	return structDescMap["CommentUpdateRequestBody"][s]
}

func (ListCloudResourceVPCRequest) Desc_ListCloudResourceVPCRequest(s string) string {
	if structDescMap["ListCloudResourceVPCRequest"] == nil {
		return ""
	}
	return structDescMap["ListCloudResourceVPCRequest"][s]
}

func (WebhookUpdateRequest) Desc_WebhookUpdateRequest(s string) string {
	if structDescMap["WebhookUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["WebhookUpdateRequest"][s]
}

func (GittarTreeSearchRequest) Desc_GittarTreeSearchRequest(s string) string {
	if structDescMap["GittarTreeSearchRequest"] == nil {
		return ""
	}
	return structDescMap["GittarTreeSearchRequest"][s]
}

func (ListLabelsData) Desc_ListLabelsData(s string) string {
	if structDescMap["ListLabelsData"] == nil {
		return ""
	}
	return structDescMap["ListLabelsData"][s]
}

func (CloudResourceRedisDetailInfoData) Desc_CloudResourceRedisDetailInfoData(s string) string {
	if structDescMap["CloudResourceRedisDetailInfoData"] == nil {
		return ""
	}
	return structDescMap["CloudResourceRedisDetailInfoData"][s]
}

func (VClusterCreateResponse) Desc_VClusterCreateResponse(s string) string {
	if structDescMap["VClusterCreateResponse"] == nil {
		return ""
	}
	return structDescMap["VClusterCreateResponse"][s]
}

func (ReleaseResource) Desc_ReleaseResource(s string) string {
	if structDescMap["ReleaseResource"] == nil {
		return ""
	}
	return structDescMap["ReleaseResource"][s]
}

func (TicketListRequest) Desc_TicketListRequest(s string) string {
	if structDescMap["TicketListRequest"] == nil {
		return ""
	}
	return structDescMap["TicketListRequest"][s]
}

func (AuditsListRequest) Desc_AuditsListRequest(s string) string {
	if structDescMap["AuditsListRequest"] == nil {
		return ""
	}
	return structDescMap["AuditsListRequest"][s]
}

func (ErrorLogListRequest) Desc_ErrorLogListRequest(s string) string {
	if structDescMap["ErrorLogListRequest"] == nil {
		return ""
	}
	return structDescMap["ErrorLogListRequest"][s]
}

func (ImageSearchRequest) Desc_ImageSearchRequest(s string) string {
	if structDescMap["ImageSearchRequest"] == nil {
		return ""
	}
	return structDescMap["ImageSearchRequest"][s]
}

func (OneDataAnalysisFuzzyAttrsRequest) Desc_OneDataAnalysisFuzzyAttrsRequest(s string) string {
	if structDescMap["OneDataAnalysisFuzzyAttrsRequest"] == nil {
		return ""
	}
	return structDescMap["OneDataAnalysisFuzzyAttrsRequest"][s]
}

func (TicketCreateRequest) Desc_TicketCreateRequest(s string) string {
	if structDescMap["TicketCreateRequest"] == nil {
		return ""
	}
	return structDescMap["TicketCreateRequest"][s]
}

func (ImageCreateRequest) Desc_ImageCreateRequest(s string) string {
	if structDescMap["ImageCreateRequest"] == nil {
		return ""
	}
	return structDescMap["ImageCreateRequest"][s]
}

func (ProjectListRequest) Desc_ProjectListRequest(s string) string {
	if structDescMap["ProjectListRequest"] == nil {
		return ""
	}
	return structDescMap["ProjectListRequest"][s]
}

func (MemberAddOptions) Desc_MemberAddOptions(s string) string {
	if structDescMap["MemberAddOptions"] == nil {
		return ""
	}
	return structDescMap["MemberAddOptions"][s]
}

func (AddonCreateOptions) Desc_AddonCreateOptions(s string) string {
	if structDescMap["AddonCreateOptions"] == nil {
		return ""
	}
	return structDescMap["AddonCreateOptions"][s]
}

func (ProjectDTO) Desc_ProjectDTO(s string) string {
	if structDescMap["ProjectDTO"] == nil {
		return ""
	}
	return structDescMap["ProjectDTO"][s]
}

func (Hierarchy) Desc_Hierarchy(s string) string {
	if structDescMap["Hierarchy"] == nil {
		return ""
	}
	return structDescMap["Hierarchy"][s]
}

func (AddNodesRequest) Desc_AddNodesRequest(s string) string {
	if structDescMap["AddNodesRequest"] == nil {
		return ""
	}
	return structDescMap["AddNodesRequest"][s]
}

func (ClusterQueryRequest) Desc_ClusterQueryRequest(s string) string {
	if structDescMap["ClusterQueryRequest"] == nil {
		return ""
	}
	return structDescMap["ClusterQueryRequest"][s]
}

func (TestPlanCreateRequest) Desc_TestPlanCreateRequest(s string) string {
	if structDescMap["TestPlanCreateRequest"] == nil {
		return ""
	}
	return structDescMap["TestPlanCreateRequest"][s]
}

func (OneDataAnalysisBussProcRequest) Desc_OneDataAnalysisBussProcRequest(s string) string {
	if structDescMap["OneDataAnalysisBussProcRequest"] == nil {
		return ""
	}
	return structDescMap["OneDataAnalysisBussProcRequest"][s]
}

func (ProjectCreateRequest) Desc_ProjectCreateRequest(s string) string {
	if structDescMap["ProjectCreateRequest"] == nil {
		return ""
	}
	return structDescMap["ProjectCreateRequest"][s]
}

func (OrgNexusGetRequest) Desc_OrgNexusGetRequest(s string) string {
	if structDescMap["OrgNexusGetRequest"] == nil {
		return ""
	}
	return structDescMap["OrgNexusGetRequest"][s]
}

func (ServiceItem) Desc_ServiceItem(s string) string {
	if structDescMap["ServiceItem"] == nil {
		return ""
	}
	return structDescMap["ServiceItem"][s]
}

func (ServicePort) Desc_ServicePort(s string) string {
	if structDescMap["ServicePort"] == nil {
		return ""
	}
	return structDescMap["ServicePort"][s]
}

func (ReleaseCreateRequest) Desc_ReleaseCreateRequest(s string) string {
	if structDescMap["ReleaseCreateRequest"] == nil {
		return ""
	}
	return structDescMap["ReleaseCreateRequest"][s]
}

func (MysqlDataBaseInfo) Desc_MysqlDataBaseInfo(s string) string {
	if structDescMap["MysqlDataBaseInfo"] == nil {
		return ""
	}
	return structDescMap["MysqlDataBaseInfo"][s]
}

func (CreateCloudResourceRedisRequest) Desc_CreateCloudResourceRedisRequest(s string) string {
	if structDescMap["CreateCloudResourceRedisRequest"] == nil {
		return ""
	}
	return structDescMap["CreateCloudResourceRedisRequest"][s]
}

func (VolumeCreateRequest) Desc_VolumeCreateRequest(s string) string {
	if structDescMap["VolumeCreateRequest"] == nil {
		return ""
	}
	return structDescMap["VolumeCreateRequest"][s]
}

func (AddonPlanItem) Desc_AddonPlanItem(s string) string {
	if structDescMap["AddonPlanItem"] == nil {
		return ""
	}
	return structDescMap["AddonPlanItem"][s]
}

func (InstanceStatusData) Desc_InstanceStatusData(s string) string {
	if structDescMap["InstanceStatusData"] == nil {
		return ""
	}
	return structDescMap["InstanceStatusData"][s]
}

func (MicroProjectRes) Desc_MicroProjectRes(s string) string {
	if structDescMap["MicroProjectRes"] == nil {
		return ""
	}
	return structDescMap["MicroProjectRes"][s]
}

func (PermissionList) Desc_PermissionList(s string) string {
	if structDescMap["PermissionList"] == nil {
		return ""
	}
	return structDescMap["PermissionList"][s]
}

func (DereferenceClusterRequest) Desc_DereferenceClusterRequest(s string) string {
	if structDescMap["DereferenceClusterRequest"] == nil {
		return ""
	}
	return structDescMap["DereferenceClusterRequest"][s]
}

func (AddonPlanRes) Desc_AddonPlanRes(s string) string {
	if structDescMap["AddonPlanRes"] == nil {
		return ""
	}
	return structDescMap["AddonPlanRes"][s]
}

func (GittarLinesData) Desc_GittarLinesData(s string) string {
	if structDescMap["GittarLinesData"] == nil {
		return ""
	}
	return structDescMap["GittarLinesData"][s]
}

func (GittarFileData) Desc_GittarFileData(s string) string {
	if structDescMap["GittarFileData"] == nil {
		return ""
	}
	return structDescMap["GittarFileData"][s]
}

func (CreateRepoResponseData) Desc_CreateRepoResponseData(s string) string {
	if structDescMap["CreateRepoResponseData"] == nil {
		return ""
	}
	return structDescMap["CreateRepoResponseData"][s]
}

func (EditActionItem) Desc_EditActionItem(s string) string {
	if structDescMap["EditActionItem"] == nil {
		return ""
	}
	return structDescMap["EditActionItem"][s]
}

func (CloudResourceMysqlDetailInfoRequest) Desc_CloudResourceMysqlDetailInfoRequest(s string) string {
	if structDescMap["CloudResourceMysqlDetailInfoRequest"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlDetailInfoRequest"][s]
}

func (ClusterLabelsRequest) Desc_ClusterLabelsRequest(s string) string {
	if structDescMap["ClusterLabelsRequest"] == nil {
		return ""
	}
	return structDescMap["ClusterLabelsRequest"][s]
}

func (ResourceReferenceData) Desc_ResourceReferenceData(s string) string {
	if structDescMap["ResourceReferenceData"] == nil {
		return ""
	}
	return structDescMap["ResourceReferenceData"][s]
}

func (GittarQueryMrRequest) Desc_GittarQueryMrRequest(s string) string {
	if structDescMap["GittarQueryMrRequest"] == nil {
		return ""
	}
	return structDescMap["GittarQueryMrRequest"][s]
}

func (AddonFetchResponseData) Desc_AddonFetchResponseData(s string) string {
	if structDescMap["AddonFetchResponseData"] == nil {
		return ""
	}
	return structDescMap["AddonFetchResponseData"][s]
}

func (PublisherListRequest) Desc_PublisherListRequest(s string) string {
	if structDescMap["PublisherListRequest"] == nil {
		return ""
	}
	return structDescMap["PublisherListRequest"][s]
}

func (OrgRunningTasksListRequest) Desc_OrgRunningTasksListRequest(s string) string {
	if structDescMap["OrgRunningTasksListRequest"] == nil {
		return ""
	}
	return structDescMap["OrgRunningTasksListRequest"][s]
}

func (HttpHealthCheck) Desc_HttpHealthCheck(s string) string {
	if structDescMap["HttpHealthCheck"] == nil {
		return ""
	}
	return structDescMap["HttpHealthCheck"][s]
}

func (CommentUpdateResponse) Desc_CommentUpdateResponse(s string) string {
	if structDescMap["CommentUpdateResponse"] == nil {
		return ""
	}
	return structDescMap["CommentUpdateResponse"][s]
}

func (AddonRes) Desc_AddonRes(s string) string {
	if structDescMap["AddonRes"] == nil {
		return ""
	}
	return structDescMap["AddonRes"][s]
}

func (MemberDestroyRequest) Desc_MemberDestroyRequest(s string) string {
	if structDescMap["MemberDestroyRequest"] == nil {
		return ""
	}
	return structDescMap["MemberDestroyRequest"][s]
}

func (MicroProjectMenuRes) Desc_MicroProjectMenuRes(s string) string {
	if structDescMap["MicroProjectMenuRes"] == nil {
		return ""
	}
	return structDescMap["MicroProjectMenuRes"][s]
}

func (ClusterLabels) Desc_ClusterLabels(s string) string {
	if structDescMap["ClusterLabels"] == nil {
		return ""
	}
	return structDescMap["ClusterLabels"][s]
}

func (Service) Desc_Service(s string) string {
	if structDescMap["Service"] == nil {
		return ""
	}
	return structDescMap["Service"][s]
}

func (NotifyHistory) Desc_NotifyHistory(s string) string {
	if structDescMap["NotifyHistory"] == nil {
		return ""
	}
	return structDescMap["NotifyHistory"][s]
}

func (ReleaseNameListRequest) Desc_ReleaseNameListRequest(s string) string {
	if structDescMap["ReleaseNameListRequest"] == nil {
		return ""
	}
	return structDescMap["ReleaseNameListRequest"][s]
}

func (ComponentProtocolScenario) Desc_ComponentProtocolScenario(s string) string {
	if structDescMap["ComponentProtocolScenario"] == nil {
		return ""
	}
	return structDescMap["ComponentProtocolScenario"][s]
}

func (IssueUpdateRequest) Desc_IssueUpdateRequest(s string) string {
	if structDescMap["IssueUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["IssueUpdateRequest"][s]
}

func (DomainUpdateRequest) Desc_DomainUpdateRequest(s string) string {
	if structDescMap["DomainUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["DomainUpdateRequest"][s]
}

func (CloudResourceSetTagRequest) Desc_CloudResourceSetTagRequest(s string) string {
	if structDescMap["CloudResourceSetTagRequest"] == nil {
		return ""
	}
	return structDescMap["CloudResourceSetTagRequest"][s]
}

func (CreateCloudResourceBaseInfo) Desc_CreateCloudResourceBaseInfo(s string) string {
	if structDescMap["CreateCloudResourceBaseInfo"] == nil {
		return ""
	}
	return structDescMap["CreateCloudResourceBaseInfo"][s]
}

func (ExecHealthCheck) Desc_ExecHealthCheck(s string) string {
	if structDescMap["ExecHealthCheck"] == nil {
		return ""
	}
	return structDescMap["ExecHealthCheck"][s]
}

func (TicketUpdateResponse) Desc_TicketUpdateResponse(s string) string {
	if structDescMap["TicketUpdateResponse"] == nil {
		return ""
	}
	return structDescMap["TicketUpdateResponse"][s]
}

func (ResourceReferenceResp) Desc_ResourceReferenceResp(s string) string {
	if structDescMap["ResourceReferenceResp"] == nil {
		return ""
	}
	return structDescMap["ResourceReferenceResp"][s]
}

func (CloudResourceOnsGroupBaseInfo) Desc_CloudResourceOnsGroupBaseInfo(s string) string {
	if structDescMap["CloudResourceOnsGroupBaseInfo"] == nil {
		return ""
	}
	return structDescMap["CloudResourceOnsGroupBaseInfo"][s]
}

func (WebhookUpdateRequestBody) Desc_WebhookUpdateRequestBody(s string) string {
	if structDescMap["WebhookUpdateRequestBody"] == nil {
		return ""
	}
	return structDescMap["WebhookUpdateRequestBody"][s]
}

func (IterationUpdateRequest) Desc_IterationUpdateRequest(s string) string {
	if structDescMap["IterationUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["IterationUpdateRequest"][s]
}

func (DashboardDetailRequest) Desc_DashboardDetailRequest(s string) string {
	if structDescMap["DashboardDetailRequest"] == nil {
		return ""
	}
	return structDescMap["DashboardDetailRequest"][s]
}

func (PipelinePageListRequest) Desc_PipelinePageListRequest(s string) string {
	if structDescMap["PipelinePageListRequest"] == nil {
		return ""
	}
	return structDescMap["PipelinePageListRequest"][s]
}

func (NexusUserEnsureRequest) Desc_NexusUserEnsureRequest(s string) string {
	if structDescMap["NexusUserEnsureRequest"] == nil {
		return ""
	}
	return structDescMap["NexusUserEnsureRequest"][s]
}

func (OrgResourceInfo) Desc_OrgResourceInfo(s string) string {
	if structDescMap["OrgResourceInfo"] == nil {
		return ""
	}
	return structDescMap["OrgResourceInfo"][s]
}

func (MigrationStatusDesc) Desc_MigrationStatusDesc(s string) string {
	if structDescMap["MigrationStatusDesc"] == nil {
		return ""
	}
	return structDescMap["MigrationStatusDesc"][s]
}

func (EndpointDomainsItem) Desc_EndpointDomainsItem(s string) string {
	if structDescMap["EndpointDomainsItem"] == nil {
		return ""
	}
	return structDescMap["EndpointDomainsItem"][s]
}

func (ClusterResourceResponse) Desc_ClusterResourceResponse(s string) string {
	if structDescMap["ClusterResourceResponse"] == nil {
		return ""
	}
	return structDescMap["ClusterResourceResponse"][s]
}

func (Hook) Desc_Hook(s string) string {
	if structDescMap["Hook"] == nil {
		return ""
	}
	return structDescMap["Hook"][s]
}

func (BranchRule) Desc_BranchRule(s string) string {
	if structDescMap["BranchRule"] == nil {
		return ""
	}
	return structDescMap["BranchRule"][s]
}

func (CloudResourceMysqlBasicData) Desc_CloudResourceMysqlBasicData(s string) string {
	if structDescMap["CloudResourceMysqlBasicData"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlBasicData"][s]
}

func (OneDataAnalysisOutputTablesRequest) Desc_OneDataAnalysisOutputTablesRequest(s string) string {
	if structDescMap["OneDataAnalysisOutputTablesRequest"] == nil {
		return ""
	}
	return structDescMap["OneDataAnalysisOutputTablesRequest"][s]
}

func (PipelineYml) Desc_PipelineYml(s string) string {
	if structDescMap["PipelineYml"] == nil {
		return ""
	}
	return structDescMap["PipelineYml"][s]
}

func (BaseResource) Desc_BaseResource(s string) string {
	if structDescMap["BaseResource"] == nil {
		return ""
	}
	return structDescMap["BaseResource"][s]
}

func (NodeResourceInfo) Desc_NodeResourceInfo(s string) string {
	if structDescMap["NodeResourceInfo"] == nil {
		return ""
	}
	return structDescMap["NodeResourceInfo"][s]
}

func (CommentCreateRequest) Desc_CommentCreateRequest(s string) string {
	if structDescMap["CommentCreateRequest"] == nil {
		return ""
	}
	return structDescMap["CommentCreateRequest"][s]
}

func (ApplicationInitRequest) Desc_ApplicationInitRequest(s string) string {
	if structDescMap["ApplicationInitRequest"] == nil {
		return ""
	}
	return structDescMap["ApplicationInitRequest"][s]
}

func (CloudResourceOnsBasicData) Desc_CloudResourceOnsBasicData(s string) string {
	if structDescMap["CloudResourceOnsBasicData"] == nil {
		return ""
	}
	return structDescMap["CloudResourceOnsBasicData"][s]
}

func (InstanceReferenceRes) Desc_InstanceReferenceRes(s string) string {
	if structDescMap["InstanceReferenceRes"] == nil {
		return ""
	}
	return structDescMap["InstanceReferenceRes"][s]
}

func (MemberLabelList) Desc_MemberLabelList(s string) string {
	if structDescMap["MemberLabelList"] == nil {
		return ""
	}
	return structDescMap["MemberLabelList"][s]
}

func (InstanceInfoRequest) Desc_InstanceInfoRequest(s string) string {
	if structDescMap["InstanceInfoRequest"] == nil {
		return ""
	}
	return structDescMap["InstanceInfoRequest"][s]
}

func (CloudResourceMysqlListAccountRequest) Desc_CloudResourceMysqlListAccountRequest(s string) string {
	if structDescMap["CloudResourceMysqlListAccountRequest"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlListAccountRequest"][s]
}

func (DeploymentListRequest) Desc_DeploymentListRequest(s string) string {
	if structDescMap["DeploymentListRequest"] == nil {
		return ""
	}
	return structDescMap["DeploymentListRequest"][s]
}

func (AddonProviderDataResp) Desc_AddonProviderDataResp(s string) string {
	if structDescMap["AddonProviderDataResp"] == nil {
		return ""
	}
	return structDescMap["AddonProviderDataResp"][s]
}

func (PipelineCmsConfigValue) Desc_PipelineCmsConfigValue(s string) string {
	if structDescMap["PipelineCmsConfigValue"] == nil {
		return ""
	}
	return structDescMap["PipelineCmsConfigValue"][s]
}

func (Dice) Desc_Dice(s string) string {
	if structDescMap["Dice"] == nil {
		return ""
	}
	return structDescMap["Dice"][s]
}

func (EffectivenessRequest) Desc_EffectivenessRequest(s string) string {
	if structDescMap["EffectivenessRequest"] == nil {
		return ""
	}
	return structDescMap["EffectivenessRequest"][s]
}

func (HookLocation) Desc_HookLocation(s string) string {
	if structDescMap["HookLocation"] == nil {
		return ""
	}
	return structDescMap["HookLocation"][s]
}

func (PipelineIDSelectByLabelRequest) Desc_PipelineIDSelectByLabelRequest(s string) string {
	if structDescMap["PipelineIDSelectByLabelRequest"] == nil {
		return ""
	}
	return structDescMap["PipelineIDSelectByLabelRequest"][s]
}

func (ActivitiyListRequest) Desc_ActivitiyListRequest(s string) string {
	if structDescMap["ActivitiyListRequest"] == nil {
		return ""
	}
	return structDescMap["ActivitiyListRequest"][s]
}

func (AddonHandlerCreateItem) Desc_AddonHandlerCreateItem(s string) string {
	if structDescMap["AddonHandlerCreateItem"] == nil {
		return ""
	}
	return structDescMap["AddonHandlerCreateItem"][s]
}

func (ActionCache) Desc_ActionCache(s string) string {
	if structDescMap["ActionCache"] == nil {
		return ""
	}
	return structDescMap["ActionCache"][s]
}

func (DashboardCreateRequest) Desc_DashboardCreateRequest(s string) string {
	if structDescMap["DashboardCreateRequest"] == nil {
		return ""
	}
	return structDescMap["DashboardCreateRequest"][s]
}

func (AddonConfigUpdateRequest) Desc_AddonConfigUpdateRequest(s string) string {
	if structDescMap["AddonConfigUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["AddonConfigUpdateRequest"][s]
}

func (ExtensionQueryRequest) Desc_ExtensionQueryRequest(s string) string {
	if structDescMap["ExtensionQueryRequest"] == nil {
		return ""
	}
	return structDescMap["ExtensionQueryRequest"][s]
}

func (ServiceGroup) Desc_ServiceGroup(s string) string {
	if structDescMap["ServiceGroup"] == nil {
		return ""
	}
	return structDescMap["ServiceGroup"][s]
}

func (AddonDirectCreateRequest) Desc_AddonDirectCreateRequest(s string) string {
	if structDescMap["AddonDirectCreateRequest"] == nil {
		return ""
	}
	return structDescMap["AddonDirectCreateRequest"][s]
}

func (ValidBranch) Desc_ValidBranch(s string) string {
	if structDescMap["ValidBranch"] == nil {
		return ""
	}
	return structDescMap["ValidBranch"][s]
}

func (OneDataAnalysisStarRequest) Desc_OneDataAnalysisStarRequest(s string) string {
	if structDescMap["OneDataAnalysisStarRequest"] == nil {
		return ""
	}
	return structDescMap["OneDataAnalysisStarRequest"][s]
}

func (ExtensionVersionQueryRequest) Desc_ExtensionVersionQueryRequest(s string) string {
	if structDescMap["ExtensionVersionQueryRequest"] == nil {
		return ""
	}
	return structDescMap["ExtensionVersionQueryRequest"][s]
}

func (MemberListRequest) Desc_MemberListRequest(s string) string {
	if structDescMap["MemberListRequest"] == nil {
		return ""
	}
	return structDescMap["MemberListRequest"][s]
}

func (FuzzyQueryNotifiesBySourceRequest) Desc_FuzzyQueryNotifiesBySourceRequest(s string) string {
	if structDescMap["FuzzyQueryNotifiesBySourceRequest"] == nil {
		return ""
	}
	return structDescMap["FuzzyQueryNotifiesBySourceRequest"][s]
}

func (TestSetWithPlanCaseRels) Desc_TestSetWithPlanCaseRels(s string) string {
	if structDescMap["TestSetWithPlanCaseRels"] == nil {
		return ""
	}
	return structDescMap["TestSetWithPlanCaseRels"][s]
}

func (ApplicationListRequest) Desc_ApplicationListRequest(s string) string {
	if structDescMap["ApplicationListRequest"] == nil {
		return ""
	}
	return structDescMap["ApplicationListRequest"][s]
}

func (GitRepoConfig) Desc_GitRepoConfig(s string) string {
	if structDescMap["GitRepoConfig"] == nil {
		return ""
	}
	return structDescMap["GitRepoConfig"][s]
}

func (TestSetRecycleRequest) Desc_TestSetRecycleRequest(s string) string {
	if structDescMap["TestSetRecycleRequest"] == nil {
		return ""
	}
	return structDescMap["TestSetRecycleRequest"][s]
}

func (ComponentProtocolRequest) Desc_ComponentProtocolRequest(s string) string {
	if structDescMap["ComponentProtocolRequest"] == nil {
		return ""
	}
	return structDescMap["ComponentProtocolRequest"][s]
}

func (CloudClusterNewCreateInfo) Desc_CloudClusterNewCreateInfo(s string) string {
	if structDescMap["CloudClusterNewCreateInfo"] == nil {
		return ""
	}
	return structDescMap["CloudClusterNewCreateInfo"][s]
}

func (IdentityInfo) Desc_IdentityInfo(s string) string {
	if structDescMap["IdentityInfo"] == nil {
		return ""
	}
	return structDescMap["IdentityInfo"][s]
}

func (ReleaseListRequest) Desc_ReleaseListRequest(s string) string {
	if structDescMap["ReleaseListRequest"] == nil {
		return ""
	}
	return structDescMap["ReleaseListRequest"][s]
}

func (ApplicationCreateRequest) Desc_ApplicationCreateRequest(s string) string {
	if structDescMap["ApplicationCreateRequest"] == nil {
		return ""
	}
	return structDescMap["ApplicationCreateRequest"][s]
}

func (CloudResourceOverviewRequest) Desc_CloudResourceOverviewRequest(s string) string {
	if structDescMap["CloudResourceOverviewRequest"] == nil {
		return ""
	}
	return structDescMap["CloudResourceOverviewRequest"][s]
}

func (CloudResourceMysqlDB) Desc_CloudResourceMysqlDB(s string) string {
	if structDescMap["CloudResourceMysqlDB"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlDB"][s]
}

func (OrgDTO) Desc_OrgDTO(s string) string {
	if structDescMap["OrgDTO"] == nil {
		return ""
	}
	return structDescMap["OrgDTO"][s]
}

func (PublishItemStatisticsDetailResponse) Desc_PublishItemStatisticsDetailResponse(s string) string {
	if structDescMap["PublishItemStatisticsDetailResponse"] == nil {
		return ""
	}
	return structDescMap["PublishItemStatisticsDetailResponse"][s]
}

func (TestPlanCaseRelCreateRequest) Desc_TestPlanCaseRelCreateRequest(s string) string {
	if structDescMap["TestPlanCaseRelCreateRequest"] == nil {
		return ""
	}
	return structDescMap["TestPlanCaseRelCreateRequest"][s]
}

func (Parameter) Desc_Parameter(s string) string {
	if structDescMap["Parameter"] == nil {
		return ""
	}
	return structDescMap["Parameter"][s]
}

func (ApplicationStats) Desc_ApplicationStats(s string) string {
	if structDescMap["ApplicationStats"] == nil {
		return ""
	}
	return structDescMap["ApplicationStats"][s]
}

func (GittarCreateBranchRequest) Desc_GittarCreateBranchRequest(s string) string {
	if structDescMap["GittarCreateBranchRequest"] == nil {
		return ""
	}
	return structDescMap["GittarCreateBranchRequest"][s]
}

func (CreateRepoRequest) Desc_CreateRepoRequest(s string) string {
	if structDescMap["CreateRepoRequest"] == nil {
		return ""
	}
	return structDescMap["CreateRepoRequest"][s]
}

func (ListCloudAddonBasicRequest) Desc_ListCloudAddonBasicRequest(s string) string {
	if structDescMap["ListCloudAddonBasicRequest"] == nil {
		return ""
	}
	return structDescMap["ListCloudAddonBasicRequest"][s]
}

func (ImageListRequest) Desc_ImageListRequest(s string) string {
	if structDescMap["ImageListRequest"] == nil {
		return ""
	}
	return structDescMap["ImageListRequest"][s]
}

func (NoticeListRequest) Desc_NoticeListRequest(s string) string {
	if structDescMap["NoticeListRequest"] == nil {
		return ""
	}
	return structDescMap["NoticeListRequest"][s]
}

func (IssueBatchUpdateRequest) Desc_IssueBatchUpdateRequest(s string) string {
	if structDescMap["IssueBatchUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["IssueBatchUpdateRequest"][s]
}

func (TestSetUpdateRequest) Desc_TestSetUpdateRequest(s string) string {
	if structDescMap["TestSetUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["TestSetUpdateRequest"][s]
}

func (NotifyItem) Desc_NotifyItem(s string) string {
	if structDescMap["NotifyItem"] == nil {
		return ""
	}
	return structDescMap["NotifyItem"][s]
}

func (OrgSearchRequest) Desc_OrgSearchRequest(s string) string {
	if structDescMap["OrgSearchRequest"] == nil {
		return ""
	}
	return structDescMap["OrgSearchRequest"][s]
}

func (CloudAddonResourceDeleteRequest) Desc_CloudAddonResourceDeleteRequest(s string) string {
	if structDescMap["CloudAddonResourceDeleteRequest"] == nil {
		return ""
	}
	return structDescMap["CloudAddonResourceDeleteRequest"][s]
}

func (AttachDest) Desc_AttachDest(s string) string {
	if structDescMap["AttachDest"] == nil {
		return ""
	}
	return structDescMap["AttachDest"][s]
}

func (TicketCreateResponse) Desc_TicketCreateResponse(s string) string {
	if structDescMap["TicketCreateResponse"] == nil {
		return ""
	}
	return structDescMap["TicketCreateResponse"][s]
}

func (ProjectUpdateBody) Desc_ProjectUpdateBody(s string) string {
	if structDescMap["ProjectUpdateBody"] == nil {
		return ""
	}
	return structDescMap["ProjectUpdateBody"][s]
}

func (AddonReferenceRes) Desc_AddonReferenceRes(s string) string {
	if structDescMap["AddonReferenceRes"] == nil {
		return ""
	}
	return structDescMap["AddonReferenceRes"][s]
}

func (AddonScaleRequest) Desc_AddonScaleRequest(s string) string {
	if structDescMap["AddonScaleRequest"] == nil {
		return ""
	}
	return structDescMap["AddonScaleRequest"][s]
}

func (NamespaceRelationCreateRequest) Desc_NamespaceRelationCreateRequest(s string) string {
	if structDescMap["NamespaceRelationCreateRequest"] == nil {
		return ""
	}
	return structDescMap["NamespaceRelationCreateRequest"][s]
}

func (RuntimeReleaseCreateRequest) Desc_RuntimeReleaseCreateRequest(s string) string {
	if structDescMap["RuntimeReleaseCreateRequest"] == nil {
		return ""
	}
	return structDescMap["RuntimeReleaseCreateRequest"][s]
}

func (IssueCreateRequest) Desc_IssueCreateRequest(s string) string {
	if structDescMap["IssueCreateRequest"] == nil {
		return ""
	}
	return structDescMap["IssueCreateRequest"][s]
}

func (StatusDesc) Desc_StatusDesc(s string) string {
	if structDescMap["StatusDesc"] == nil {
		return ""
	}
	return structDescMap["StatusDesc"][s]
}

func (Resources) Desc_Resources(s string) string {
	if structDescMap["Resources"] == nil {
		return ""
	}
	return structDescMap["Resources"][s]
}

func (TicketUpdateRequestBody) Desc_TicketUpdateRequestBody(s string) string {
	if structDescMap["TicketUpdateRequestBody"] == nil {
		return ""
	}
	return structDescMap["TicketUpdateRequestBody"][s]
}

func (PwdSecurityConfig) Desc_PwdSecurityConfig(s string) string {
	if structDescMap["PwdSecurityConfig"] == nil {
		return ""
	}
	return structDescMap["PwdSecurityConfig"][s]
}

func (ScopeResource) Desc_ScopeResource(s string) string {
	if structDescMap["ScopeResource"] == nil {
		return ""
	}
	return structDescMap["ScopeResource"][s]
}

func (MemberAddRequest) Desc_MemberAddRequest(s string) string {
	if structDescMap["MemberAddRequest"] == nil {
		return ""
	}
	return structDescMap["MemberAddRequest"][s]
}

func (CloudResourceMysqlDetailInfoData) Desc_CloudResourceMysqlDetailInfoData(s string) string {
	if structDescMap["CloudResourceMysqlDetailInfoData"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlDetailInfoData"][s]
}

func (Authorize) Desc_Authorize(s string) string {
	if structDescMap["Authorize"] == nil {
		return ""
	}
	return structDescMap["Authorize"][s]
}

func (ScriptInfo) Desc_ScriptInfo(s string) string {
	if structDescMap["ScriptInfo"] == nil {
		return ""
	}
	return structDescMap["ScriptInfo"][s]
}

func (CloudResourceOnsTopicInfo) Desc_CloudResourceOnsTopicInfo(s string) string {
	if structDescMap["CloudResourceOnsTopicInfo"] == nil {
		return ""
	}
	return structDescMap["CloudResourceOnsTopicInfo"][s]
}

func (CreateCloudResourceMysqlAccountRequest) Desc_CreateCloudResourceMysqlAccountRequest(s string) string {
	if structDescMap["CreateCloudResourceMysqlAccountRequest"] == nil {
		return ""
	}
	return structDescMap["CreateCloudResourceMysqlAccountRequest"][s]
}

func (UserProfile) Desc_UserProfile(s string) string {
	if structDescMap["UserProfile"] == nil {
		return ""
	}
	return structDescMap["UserProfile"][s]
}

func (CreateHookRequest) Desc_CreateHookRequest(s string) string {
	if structDescMap["CreateHookRequest"] == nil {
		return ""
	}
	return structDescMap["CreateHookRequest"][s]
}

func (PagePipeline) Desc_PagePipeline(s string) string {
	if structDescMap["PagePipeline"] == nil {
		return ""
	}
	return structDescMap["PagePipeline"][s]
}

func (NamespaceCreateRequest) Desc_NamespaceCreateRequest(s string) string {
	if structDescMap["NamespaceCreateRequest"] == nil {
		return ""
	}
	return structDescMap["NamespaceCreateRequest"][s]
}

func (ListSchemasQueryParams) Desc_ListSchemasQueryParams(s string) string {
	if structDescMap["ListSchemasQueryParams"] == nil {
		return ""
	}
	return structDescMap["ListSchemasQueryParams"][s]
}

func (Bind) Desc_Bind(s string) string {
	if structDescMap["Bind"] == nil {
		return ""
	}
	return structDescMap["Bind"][s]
}

func (AddonCreateItem) Desc_AddonCreateItem(s string) string {
	if structDescMap["AddonCreateItem"] == nil {
		return ""
	}
	return structDescMap["AddonCreateItem"][s]
}

func (CheckRun) Desc_CheckRun(s string) string {
	if structDescMap["CheckRun"] == nil {
		return ""
	}
	return structDescMap["CheckRun"][s]
}

func (CustomAddonUpdateRequest) Desc_CustomAddonUpdateRequest(s string) string {
	if structDescMap["CustomAddonUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["CustomAddonUpdateRequest"][s]
}

func (PipelineDatabaseGC) Desc_PipelineDatabaseGC(s string) string {
	if structDescMap["PipelineDatabaseGC"] == nil {
		return ""
	}
	return structDescMap["PipelineDatabaseGC"][s]
}

func (ListCloudResourceECSRequest) Desc_ListCloudResourceECSRequest(s string) string {
	if structDescMap["ListCloudResourceECSRequest"] == nil {
		return ""
	}
	return structDescMap["ListCloudResourceECSRequest"][s]
}

func (MiddlewareFetchResponseData) Desc_MiddlewareFetchResponseData(s string) string {
	if structDescMap["MiddlewareFetchResponseData"] == nil {
		return ""
	}
	return structDescMap["MiddlewareFetchResponseData"][s]
}

func (IterationPagingRequest) Desc_IterationPagingRequest(s string) string {
	if structDescMap["IterationPagingRequest"] == nil {
		return ""
	}
	return structDescMap["IterationPagingRequest"][s]
}

func (DashBoardDTO) Desc_DashBoardDTO(s string) string {
	if structDescMap["DashBoardDTO"] == nil {
		return ""
	}
	return structDescMap["DashBoardDTO"][s]
}

func (PipelineDBGCItem) Desc_PipelineDBGCItem(s string) string {
	if structDescMap["PipelineDBGCItem"] == nil {
		return ""
	}
	return structDescMap["PipelineDBGCItem"][s]
}

func (QueryNotifyGroupRequest) Desc_QueryNotifyGroupRequest(s string) string {
	if structDescMap["QueryNotifyGroupRequest"] == nil {
		return ""
	}
	return structDescMap["QueryNotifyGroupRequest"][s]
}

func (RouteOptions) Desc_RouteOptions(s string) string {
	if structDescMap["RouteOptions"] == nil {
		return ""
	}
	return structDescMap["RouteOptions"][s]
}

func (PipelineReportSetPagingRequest) Desc_PipelineReportSetPagingRequest(s string) string {
	if structDescMap["PipelineReportSetPagingRequest"] == nil {
		return ""
	}
	return structDescMap["PipelineReportSetPagingRequest"][s]
}

func (PublishItemStatisticsErrTrendResponse) Desc_PublishItemStatisticsErrTrendResponse(s string) string {
	if structDescMap["PublishItemStatisticsErrTrendResponse"] == nil {
		return ""
	}
	return structDescMap["PublishItemStatisticsErrTrendResponse"][s]
}

func (ActionCallback) Desc_ActionCallback(s string) string {
	if structDescMap["ActionCallback"] == nil {
		return ""
	}
	return structDescMap["ActionCallback"][s]
}

func (WebhookListRequest) Desc_WebhookListRequest(s string) string {
	if structDescMap["WebhookListRequest"] == nil {
		return ""
	}
	return structDescMap["WebhookListRequest"][s]
}

func (AddonProviderRequest) Desc_AddonProviderRequest(s string) string {
	if structDescMap["AddonProviderRequest"] == nil {
		return ""
	}
	return structDescMap["AddonProviderRequest"][s]
}

func (APIOperation) Desc_APIOperation(s string) string {
	if structDescMap["APIOperation"] == nil {
		return ""
	}
	return structDescMap["APIOperation"][s]
}

func (Ticket) Desc_Ticket(s string) string {
	if structDescMap["Ticket"] == nil {
		return ""
	}
	return structDescMap["Ticket"][s]
}

func (WebhookInspectRequest) Desc_WebhookInspectRequest(s string) string {
	if structDescMap["WebhookInspectRequest"] == nil {
		return ""
	}
	return structDescMap["WebhookInspectRequest"][s]
}

func (ProjectStats) Desc_ProjectStats(s string) string {
	if structDescMap["ProjectStats"] == nil {
		return ""
	}
	return structDescMap["ProjectStats"][s]
}

func (AuditListCleanCronRequest) Desc_AuditListCleanCronRequest(s string) string {
	if structDescMap["AuditListCleanCronRequest"] == nil {
		return ""
	}
	return structDescMap["AuditListCleanCronRequest"][s]
}

func (PipelineDTO) Desc_PipelineDTO(s string) string {
	if structDescMap["PipelineDTO"] == nil {
		return ""
	}
	return structDescMap["PipelineDTO"][s]
}

func (ScheduleInfo2) Desc_ScheduleInfo2(s string) string {
	if structDescMap["ScheduleInfo2"] == nil {
		return ""
	}
	return structDescMap["ScheduleInfo2"][s]
}

func (ServiceGroupCreateV2Request) Desc_ServiceGroupCreateV2Request(s string) string {
	if structDescMap["ServiceGroupCreateV2Request"] == nil {
		return ""
	}
	return structDescMap["ServiceGroupCreateV2Request"][s]
}

func (OneDataAnalysisRequest) Desc_OneDataAnalysisRequest(s string) string {
	if structDescMap["OneDataAnalysisRequest"] == nil {
		return ""
	}
	return structDescMap["OneDataAnalysisRequest"][s]
}

func (CreateNotifyItemResponse) Desc_CreateNotifyItemResponse(s string) string {
	if structDescMap["CreateNotifyItemResponse"] == nil {
		return ""
	}
	return structDescMap["CreateNotifyItemResponse"][s]
}

func (CloudResourceOnsTopicInfoRequest) Desc_CloudResourceOnsTopicInfoRequest(s string) string {
	if structDescMap["CloudResourceOnsTopicInfoRequest"] == nil {
		return ""
	}
	return structDescMap["CloudResourceOnsTopicInfoRequest"][s]
}

func (InstanceDetailRes) Desc_InstanceDetailRes(s string) string {
	if structDescMap["InstanceDetailRes"] == nil {
		return ""
	}
	return structDescMap["InstanceDetailRes"][s]
}

func (TestSetListRequest) Desc_TestSetListRequest(s string) string {
	if structDescMap["TestSetListRequest"] == nil {
		return ""
	}
	return structDescMap["TestSetListRequest"][s]
}

func (ExtensionVersionCreateRequest) Desc_ExtensionVersionCreateRequest(s string) string {
	if structDescMap["ExtensionVersionCreateRequest"] == nil {
		return ""
	}
	return structDescMap["ExtensionVersionCreateRequest"][s]
}

func (PublishItemStatisticsDetailRequest) Desc_PublishItemStatisticsDetailRequest(s string) string {
	if structDescMap["PublishItemStatisticsDetailRequest"] == nil {
		return ""
	}
	return structDescMap["PublishItemStatisticsDetailRequest"][s]
}

func (PodInfoRequest) Desc_PodInfoRequest(s string) string {
	if structDescMap["PodInfoRequest"] == nil {
		return ""
	}
	return structDescMap["PodInfoRequest"][s]
}

func (ApplicationFetchRequest) Desc_ApplicationFetchRequest(s string) string {
	if structDescMap["ApplicationFetchRequest"] == nil {
		return ""
	}
	return structDescMap["ApplicationFetchRequest"][s]
}

func (EventHeader) Desc_EventHeader(s string) string {
	if structDescMap["EventHeader"] == nil {
		return ""
	}
	return structDescMap["EventHeader"][s]
}

func (ProjectDetailRequest) Desc_ProjectDetailRequest(s string) string {
	if structDescMap["ProjectDetailRequest"] == nil {
		return ""
	}
	return structDescMap["ProjectDetailRequest"][s]
}

func (TestPlanTestCaseRelDeleteRequest) Desc_TestPlanTestCaseRelDeleteRequest(s string) string {
	if structDescMap["TestPlanTestCaseRelDeleteRequest"] == nil {
		return ""
	}
	return structDescMap["TestPlanTestCaseRelDeleteRequest"][s]
}

func (VClusterCreateRequest) Desc_VClusterCreateRequest(s string) string {
	if structDescMap["VClusterCreateRequest"] == nil {
		return ""
	}
	return structDescMap["VClusterCreateRequest"][s]
}

func (ScheduleLabelListData) Desc_ScheduleLabelListData(s string) string {
	if structDescMap["ScheduleLabelListData"] == nil {
		return ""
	}
	return structDescMap["ScheduleLabelListData"][s]
}

func (ReleaseListResponseData) Desc_ReleaseListResponseData(s string) string {
	if structDescMap["ReleaseListResponseData"] == nil {
		return ""
	}
	return structDescMap["ReleaseListResponseData"][s]
}

func (PluginParamDto) Desc_PluginParamDto(s string) string {
	if structDescMap["PluginParamDto"] == nil {
		return ""
	}
	return structDescMap["PluginParamDto"][s]
}

func (ReleaseUpdateRequestData) Desc_ReleaseUpdateRequestData(s string) string {
	if structDescMap["ReleaseUpdateRequestData"] == nil {
		return ""
	}
	return structDescMap["ReleaseUpdateRequestData"][s]
}

func (ExtensionSearchRequest) Desc_ExtensionSearchRequest(s string) string {
	if structDescMap["ExtensionSearchRequest"] == nil {
		return ""
	}
	return structDescMap["ExtensionSearchRequest"][s]
}

func (APIAssetVersionInstanceCreateRequest) Desc_APIAssetVersionInstanceCreateRequest(s string) string {
	if structDescMap["APIAssetVersionInstanceCreateRequest"] == nil {
		return ""
	}
	return structDescMap["APIAssetVersionInstanceCreateRequest"][s]
}

func (PermissionCheckRequest) Desc_PermissionCheckRequest(s string) string {
	if structDescMap["PermissionCheckRequest"] == nil {
		return ""
	}
	return structDescMap["PermissionCheckRequest"][s]
}

func (PipelineInvokedComboRequest) Desc_PipelineInvokedComboRequest(s string) string {
	if structDescMap["PipelineInvokedComboRequest"] == nil {
		return ""
	}
	return structDescMap["PipelineInvokedComboRequest"][s]
}

func (TestSet) Desc_TestSet(s string) string {
	if structDescMap["TestSet"] == nil {
		return ""
	}
	return structDescMap["TestSet"][s]
}

func (DeploymentStatusDTO) Desc_DeploymentStatusDTO(s string) string {
	if structDescMap["DeploymentStatusDTO"] == nil {
		return ""
	}
	return structDescMap["DeploymentStatusDTO"][s]
}

func (PublishItemStatisticsTrendResponse) Desc_PublishItemStatisticsTrendResponse(s string) string {
	if structDescMap["PublishItemStatisticsTrendResponse"] == nil {
		return ""
	}
	return structDescMap["PublishItemStatisticsTrendResponse"][s]
}

func (UpdateCustomAddonRequest) Desc_UpdateCustomAddonRequest(s string) string {
	if structDescMap["UpdateCustomAddonRequest"] == nil {
		return ""
	}
	return structDescMap["UpdateCustomAddonRequest"][s]
}

func (TestCallBackRequest) Desc_TestCallBackRequest(s string) string {
	if structDescMap["TestCallBackRequest"] == nil {
		return ""
	}
	return structDescMap["TestCallBackRequest"][s]
}

func (ReleaseGetResponseData) Desc_ReleaseGetResponseData(s string) string {
	if structDescMap["ReleaseGetResponseData"] == nil {
		return ""
	}
	return structDescMap["ReleaseGetResponseData"][s]
}

func (WebhookPingRequest) Desc_WebhookPingRequest(s string) string {
	if structDescMap["WebhookPingRequest"] == nil {
		return ""
	}
	return structDescMap["WebhookPingRequest"][s]
}

func (MiddlewareListItem) Desc_MiddlewareListItem(s string) string {
	if structDescMap["MiddlewareListItem"] == nil {
		return ""
	}
	return structDescMap["MiddlewareListItem"][s]
}

func (GittarCreateCommitRequest) Desc_GittarCreateCommitRequest(s string) string {
	if structDescMap["GittarCreateCommitRequest"] == nil {
		return ""
	}
	return structDescMap["GittarCreateCommitRequest"][s]
}

func (Blame) Desc_Blame(s string) string {
	if structDescMap["Blame"] == nil {
		return ""
	}
	return structDescMap["Blame"][s]
}

func (UnifiedFileTreeNodeFuzzySearchRequest) Desc_UnifiedFileTreeNodeFuzzySearchRequest(s string) string {
	if structDescMap["UnifiedFileTreeNodeFuzzySearchRequest"] == nil {
		return ""
	}
	return structDescMap["UnifiedFileTreeNodeFuzzySearchRequest"][s]
}

func (IssueManHourSumResponse) Desc_IssueManHourSumResponse(s string) string {
	if structDescMap["IssueManHourSumResponse"] == nil {
		return ""
	}
	return structDescMap["IssueManHourSumResponse"][s]
}

func (RuntimeServiceRequest) Desc_RuntimeServiceRequest(s string) string {
	if structDescMap["RuntimeServiceRequest"] == nil {
		return ""
	}
	return structDescMap["RuntimeServiceRequest"][s]
}

func (VClusterFetchResponseData) Desc_VClusterFetchResponseData(s string) string {
	if structDescMap["VClusterFetchResponseData"] == nil {
		return ""
	}
	return structDescMap["VClusterFetchResponseData"][s]
}

func (MiddlewareResourceFetchResponseData) Desc_MiddlewareResourceFetchResponseData(s string) string {
	if structDescMap["MiddlewareResourceFetchResponseData"] == nil {
		return ""
	}
	return structDescMap["MiddlewareResourceFetchResponseData"][s]
}

func (RuntimeCreateRequestExtra) Desc_RuntimeCreateRequestExtra(s string) string {
	if structDescMap["RuntimeCreateRequestExtra"] == nil {
		return ""
	}
	return structDescMap["RuntimeCreateRequestExtra"][s]
}

func (User) Desc_User(s string) string {
	if structDescMap["User"] == nil {
		return ""
	}
	return structDescMap["User"][s]
}

func (ServiceGroupCreateV2Data) Desc_ServiceGroupCreateV2Data(s string) string {
	if structDescMap["ServiceGroupCreateV2Data"] == nil {
		return ""
	}
	return structDescMap["ServiceGroupCreateV2Data"][s]
}

func (Comment) Desc_Comment(s string) string {
	if structDescMap["Comment"] == nil {
		return ""
	}
	return structDescMap["Comment"][s]
}

func (DrainNodeRequest) Desc_DrainNodeRequest(s string) string {
	if structDescMap["DrainNodeRequest"] == nil {
		return ""
	}
	return structDescMap["DrainNodeRequest"][s]
}

func (GetRuntimeAddonConfigRequest) Desc_GetRuntimeAddonConfigRequest(s string) string {
	if structDescMap["GetRuntimeAddonConfigRequest"] == nil {
		return ""
	}
	return structDescMap["GetRuntimeAddonConfigRequest"][s]
}

func (PipelineResourceGC) Desc_PipelineResourceGC(s string) string {
	if structDescMap["PipelineResourceGC"] == nil {
		return ""
	}
	return structDescMap["PipelineResourceGC"][s]
}

func (PageInfo) Desc_PageInfo(s string) string {
	if structDescMap["PageInfo"] == nil {
		return ""
	}
	return structDescMap["PageInfo"][s]
}

func (ClusterInfo) Desc_ClusterInfo(s string) string {
	if structDescMap["ClusterInfo"] == nil {
		return ""
	}
	return structDescMap["ClusterInfo"][s]
}

func (RegistryManifestsRemoveResponseData) Desc_RegistryManifestsRemoveResponseData(s string) string {
	if structDescMap["RegistryManifestsRemoveResponseData"] == nil {
		return ""
	}
	return structDescMap["RegistryManifestsRemoveResponseData"][s]
}

func (TestSuite) Desc_TestSuite(s string) string {
	if structDescMap["TestSuite"] == nil {
		return ""
	}
	return structDescMap["TestSuite"][s]
}

func (DomainListRequest) Desc_DomainListRequest(s string) string {
	if structDescMap["DomainListRequest"] == nil {
		return ""
	}
	return structDescMap["DomainListRequest"][s]
}

func (CloudResourceMysqlDBRequest) Desc_CloudResourceMysqlDBRequest(s string) string {
	if structDescMap["CloudResourceMysqlDBRequest"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlDBRequest"][s]
}

func (OnsTopic) Desc_OnsTopic(s string) string {
	if structDescMap["OnsTopic"] == nil {
		return ""
	}
	return structDescMap["OnsTopic"][s]
}

func (CustomAddonCreateRequest) Desc_CustomAddonCreateRequest(s string) string {
	if structDescMap["CustomAddonCreateRequest"] == nil {
		return ""
	}
	return structDescMap["CustomAddonCreateRequest"][s]
}

func (CloudClusterInfo) Desc_CloudClusterInfo(s string) string {
	if structDescMap["CloudClusterInfo"] == nil {
		return ""
	}
	return structDescMap["CloudClusterInfo"][s]
}

func (ComponentIngressUpdateRequest) Desc_ComponentIngressUpdateRequest(s string) string {
	if structDescMap["ComponentIngressUpdateRequest"] == nil {
		return ""
	}
	return structDescMap["ComponentIngressUpdateRequest"][s]
}

func (TestCasePagingRequest) Desc_TestCasePagingRequest(s string) string {
	if structDescMap["TestCasePagingRequest"] == nil {
		return ""
	}
	return structDescMap["TestCasePagingRequest"][s]
}

func (Role) Desc_Role(s string) string {
	if structDescMap["Role"] == nil {
		return ""
	}
	return structDescMap["Role"][s]
}

func (MysqlExec) Desc_MysqlExec(s string) string {
	if structDescMap["MysqlExec"] == nil {
		return ""
	}
	return structDescMap["MysqlExec"][s]
}

func (PipelineCreateRequestV2) Desc_PipelineCreateRequestV2(s string) string {
	if structDescMap["PipelineCreateRequestV2"] == nil {
		return ""
	}
	return structDescMap["PipelineCreateRequestV2"][s]
}

func (MergeTemplatesResponseData) Desc_MergeTemplatesResponseData(s string) string {
	if structDescMap["MergeTemplatesResponseData"] == nil {
		return ""
	}
	return structDescMap["MergeTemplatesResponseData"][s]
}

func (ServiceBind) Desc_ServiceBind(s string) string {
	if structDescMap["ServiceBind"] == nil {
		return ""
	}
	return structDescMap["ServiceBind"][s]
}

func (GittarMergeStatusRequest) Desc_GittarMergeStatusRequest(s string) string {
	if structDescMap["GittarMergeStatusRequest"] == nil {
		return ""
	}
	return structDescMap["GittarMergeStatusRequest"][s]
}

func (CloudResourceMysqlAccount) Desc_CloudResourceMysqlAccount(s string) string {
	if structDescMap["CloudResourceMysqlAccount"] == nil {
		return ""
	}
	return structDescMap["CloudResourceMysqlAccount"][s]
}

func (CreateCloudResourceMysqlRequest) Desc_CreateCloudResourceMysqlRequest(s string) string {
	if structDescMap["CreateCloudResourceMysqlRequest"] == nil {
		return ""
	}
	return structDescMap["CreateCloudResourceMysqlRequest"][s]
}

func (UserListRequest) Desc_UserListRequest(s string) string {
	if structDescMap["UserListRequest"] == nil {
		return ""
	}
	return structDescMap["UserListRequest"][s]
}

func (WebhookDeleteRequest) Desc_WebhookDeleteRequest(s string) string {
	if structDescMap["WebhookDeleteRequest"] == nil {
		return ""
	}
	return structDescMap["WebhookDeleteRequest"][s]
}

func (PipelineInvokedCombo) Desc_PipelineInvokedCombo(s string) string {
	if structDescMap["PipelineInvokedCombo"] == nil {
		return ""
	}
	return structDescMap["PipelineInvokedCombo"][s]
}

func (TestSetCreateRequest) Desc_TestSetCreateRequest(s string) string {
	if structDescMap["TestSetCreateRequest"] == nil {
		return ""
	}
	return structDescMap["TestSetCreateRequest"][s]
}
