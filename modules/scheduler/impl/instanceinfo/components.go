package instanceinfo

import (
	"github.com/erda-project/erda/apistructs"
	insinfo "github.com/erda-project/erda/modules/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/dbengine"
)

type ComponentInfo interface {
	Get() (apistructs.ComponentInfoDataList, error)
}

type ComponentInfoImpl struct {
	db *insinfo.Client
}

func NewComponentInfoImpl() *ComponentInfoImpl {
	return &ComponentInfoImpl{
		db: insinfo.New(dbengine.MustOpen()),
	}
}

func (c *ComponentInfoImpl) Get() (apistructs.ComponentInfoDataList, error) {
	r := c.db.InstanceReader()
	instances, err := r.ByMetaLike("dice_component=").Do()
	if err != nil {
		return apistructs.ComponentInfoDataList{}, err
	}
	result := apistructs.ComponentInfoDataList{}
	for _, ins := range instances {
		insInfo := apistructs.ComponentInfoData{}
		name, ok := ins.Metadata("dice_component")
		if !ok {
			continue
		}

		insInfo.Cluster = ins.Cluster
		insInfo.ComponentName = name
		insInfo.Phase = string(ins.Phase)
		insInfo.Message = ins.Message
		insInfo.ContainerID = ins.ContainerID
		insInfo.ContainerIP = ins.ContainerIP
		insInfo.HostIP = ins.HostIP
		insInfo.ExitCode = ins.ExitCode
		insInfo.CpuOrigin = ins.CpuOrigin
		insInfo.MemOrigin = ins.MemOrigin
		insInfo.CpuRequest = ins.CpuRequest
		insInfo.MemRequest = ins.MemRequest
		insInfo.CpuLimit = ins.CpuLimit
		insInfo.MemLimit = ins.MemLimit
		insInfo.Image = ins.Image
		insInfo.StartedAt = ins.StartedAt
		insInfo.FinishedAt = ins.FinishedAt
		result = append(result, insInfo)
	}
	return result, nil
}
