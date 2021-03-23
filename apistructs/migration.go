package apistructs

// MigrationStatus addon规格信息返回res
type MigrationStatusDesc struct {
	// Status 返回的运行状态
	Status StatusCode `json:"status"`
	// Desc 说明信息
	Desc string `json:"desc"`
}

const (
	// MigrationStatusInit migration初始化
	MigrationStatusInit string = "INIT"
	// MigrationStatusPending migration等待
	MigrationStatusPending string = "PENDING"
	// MigrationStatusRunning migration running
	MigrationStatusRunning string = "RUNNING"
	// MigrationStatusFail migration 失败
	MigrationStatusFail string = "FAIL"
	// MigrationStatusFinish migration 完成
	MigrationStatusFinish string = "FINISH"
	// MigrationStatusDeleted migration 删除
	MigrationStatusDeleted string = "DELETE"
)
