package pvolumes

import (
	"net/url"
	"path/filepath"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// ParseDiceYmlJobBinds 将 diceYmlJob 里老格式的 binds 转换为新的格式
func ParseDiceYmlJobBinds(diceYmlJob *diceyml.Job) ([]apistructs.Bind, error) {
	binds, err := diceyml.ParseBinds(diceYmlJob.Binds)
	if err != nil {
		return nil, err
	}
	var result []apistructs.Bind
	for _, bind := range binds {
		result = append(result, apistructs.Bind{
			ContainerPath: bind.ContainerPath,
			HostPath:      bind.HostPath,
			ReadOnly:      bind.Type == "r",
		})
	}
	return result, nil
}

// GenerateTaskCommonBinds 生成 task 通用 binds
func GenerateTaskCommonBinds(mountPoint string) []apistructs.Bind {

	const (
		dockerSock = "/var/run/docker.sock"
	)

	var binds []apistructs.Bind
	dockerSockBind := apistructs.Bind{
		HostPath:      dockerSock,
		ContainerPath: dockerSock,
		ReadOnly:      true,
	}
	binds = append(binds, dockerSockBind)

	storageURL := conf.StorageURL()
	URL, _ := url.Parse(storageURL)
	if URL.Scheme == "file" {
		var storageBind apistructs.Bind
		_path := filepath.Join(mountPoint, URL.Path)
		storageBind = apistructs.Bind{
			HostPath:      _path,
			ContainerPath: _path,
			ReadOnly:      false,
		}
		binds = append(binds, storageBind)
	}
	return binds
}
