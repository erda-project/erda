package model

// Container 容器实例元数据
type Container struct {
	BaseModel
	ContainerID         string  `json:"id" gorm:"column:container_id;type:varchar(64);index:container_id"` // 容器ID
	Deleted             bool    `json:"deleted"`                                                           // 资源是否被删除
	StartedAt           string  `json:"started_at"`                                                        // 容器启动时间
	FinishedAt          string  `json:"finished_at"`                                                       // 容器结束时间
	ExitCode            int     `json:"exit_code"`                                                         // 容器退出码
	Privileged          bool    `json:"privileged"`                                                        // 是否是特权容器
	Cluster             string  `json:"cluster_full_name"`                                                 // 集群名
	HostPrivateIPAddr   string  `json:"host_private_addr"`                                                 // 宿主机内网地址
	IPAddress           string  `json:"ip_addr"`                                                           // 容器IP地址
	Image               string  `json:"image_name"`                                                        // 容器镜像名
	CPU                 float64 `json:"cpu"`                                                               // 分配的cpu
	Memory              int64   `json:"memory"`                                                            // 分配的内存（字节）
	Disk                int64   `json:"disk"`                                                              // 分配的磁盘空间（字节）
	DiceOrg             string  `json:"dice_org"`                                                          // 所在的组织
	DiceProject         string  `gorm:"type:varchar(40);index:idx_project_id"`                             // 所在大项目
	DiceApplication     string  `json:"dice_application"`                                                  // 所在项目
	DiceRuntime         string  `gorm:"type:varchar(40);index:idx_runtime_id"`                             // 所在runtime
	DiceService         string  `json:"dice_service"`                                                      // 对应 service
	EdasAppID           string  `gorm:"type:varchar(64);index:idx_edas_app_id"`                            // EDAS 应用 ID，与 dice service 属于一个层级
	EdasAppName         string  `gorm:"type:varchar(128)"`
	EdasGroupID         string  `gorm:"type:varchar(64)"`
	DiceProjectName     string  `json:"dice_project_name"`               // 所在大项目名称
	DiceApplicationName string  `json:"dice_application_name"`           // 所在项目
	DiceRuntimeName     string  `json:"dice_runtime_name"`               // 所在runtime
	DiceComponent       string  `json:"dice_component"`                  // 组件名
	DiceAddon           string  `json:"dice_addon"`                      // 中间件id
	DiceAddonName       string  `json:"dice_addon_name"`                 // 中间件名称
	DiceWorkspace       string  `json:"dice_workspace"`                  // 部署环境
	DiceSharedLevel     string  `json:"dice_shared_level"`               // 中间件共享级别
	Status              string  `json:"status"`                          // 前期定义为docker状态（后期期望能表示服务状态）
	TimeStamp           int64   `json:"timestamp"`                       // 消息本身的时间戳
	TaskID              string  `gorm:"type:varchar(180);index:task_id"` // task id
	Env                 string  `json:"env,omitempty" gorm:"-"`          // 该容器由哪个环境发布(dev, test, staging, prod)
}

// TableName 设置模型对应数据库表名称
func (Container) TableName() string {
	return "cm_containers"
}
