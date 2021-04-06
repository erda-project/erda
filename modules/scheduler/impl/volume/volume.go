// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package volume

import (
	"errors"
	"fmt"
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/uuid"
)

type VolumeType = apistructs.VolumeType

type VolumeCreateConfig struct {
	Size int
	Type VolumeType
}

func (c VolumeCreateConfig) String() string {
	return fmt.Sprintf("volume config: size(%d G), type(%v)", c.Size, c.Type)
}

func VolumeCreateConfigFrom(r apistructs.VolumeCreateRequest) (VolumeCreateConfig, error) {
	tp, err := apistructs.VolumeTypeFromString(r.Type)
	if err != nil {
		return VolumeCreateConfig{}, err
	}
	return VolumeCreateConfig{
		Size: r.Size,
		Type: tp,
	}, nil
}

type VolumeReference = apistructs.VolumeReference

type VolumeInfo = apistructs.VolumeInfo

var IDLenLock sync.Once
var IDLen int

// VolumeIdentity 代表 volume ID 或者 volume name
type VolumeIdentity string

func (i VolumeIdentity) String() string {
	return string(i)
}

// TODO: 需要更好的区分name 还是 ID
func (i VolumeIdentity) IsNotUUID() bool {
	IDLenLock.Do(func() {
		IDLen = len(uuid.Generate())
	})
	return len(i) != IDLen
}

// EncodeVolumeType 根据 VolumeType 得到对应的 hex 编码表示
func EncodeVolumeType(t VolumeType) (string, error) {
	switch t {
	case apistructs.LocalVolume:
		return apistructs.LocalVolumeHex, nil
	case apistructs.NasVolume:
		return apistructs.NasVolumeHex, nil
	default:
		return "", errors.New("bad volumetype")
	}
}

// DecodeVolumeType 解析 `s` 的前缀，得到 VolumeType
func DecodeVolumeType(s string) (apistructs.VolumeType, error) {
	if strutil.HasPrefixes(s, apistructs.LocalVolumeHex) {
		return apistructs.LocalVolume, nil
	}
	if strutil.HasPrefixes(s, apistructs.NasVolumeHex) {
		return apistructs.NasVolume, nil
	}
	return apistructs.LocalVolume, errors.New("decode fail")
}

func NewVolumeID(config VolumeCreateConfig) (VolumeIdentity, error) {
	hex, err := EncodeVolumeType(config.Type)
	if err != nil {
		return VolumeIdentity(""), err
	}
	id := hex + uuid.Generate()
	if len(id) > 40 {
		return VolumeIdentity(id[:40]), nil
	}
	return VolumeIdentity(id), nil
}

func NewVolumeName(name string) VolumeIdentity {
	return VolumeIdentity(name)
}

type AttachDest apistructs.AttachDest

func (d AttachDest) Validate() error {
	if d.Namespace == "" {
		return fmt.Errorf("empty namespace")
	}
	if d.Service == "" {
		return fmt.Errorf("empty service")
	}
	if d.Path == "" {
		return fmt.Errorf("empty Path")
	}
	return nil
}

// 2个 AttachDest 是否相同
func (d AttachDest) Equal(d2 AttachDest) bool {
	return d.Namespace == d2.Namespace && d.Service == d2.Service && d.Path == d2.Path
}

func (d AttachDest) String() string {
	namespace := "unknownNamespace"
	if d.Namespace != "" {
		namespace = d.Namespace
	}
	service := "unknownService"
	if d.Service != "" {
		service = d.Service
	}
	path := "unknownPath"
	if d.Path != "" {
		path = d.Path
	}
	return fmt.Sprintf("<%s>:<%s>:<%s>", namespace, service, path)

}

type AttachCallback func(runtime *apistructs.ServiceGroup) (VolumeInfo, error)

type Volume interface {
	// 返回 volume 类型
	Type() VolumeType

	// 创建 volume
	// 创建 volume 不一定会实际创建，而只是创建元数据，比如 LocalVolume
	Create(config VolumeCreateConfig) (VolumeInfo, error)

	// 返回volume信息
	Info(ID VolumeIdentity) (VolumeInfo, error)

	// Attach 已存在的 volume
	Attach(ID VolumeIdentity, dst AttachDest) (AttachCallback, error)

	// UnAttach
	UnAttach(ID VolumeIdentity, dst AttachDest) (VolumeInfo, error)

	Delete(ID VolumeIdentity, force bool) error
}
