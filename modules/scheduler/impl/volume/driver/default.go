package driver

import (
	"context"
	"path/filepath"
	"time"

	"github.com/erda-project/erda/modules/scheduler/impl/volume"
	"github.com/erda-project/erda/pkg/jsonstore"

	"github.com/pkg/errors"
)

var (
	ExistedVolumeID          = errors.New("Existed Volume ID")
	NotFoundVolume           = errors.New("Not found volume")
	VolumeNameReferNilVolume = errors.New("volume name refer to nil volume")
)

// etcd 中存储的 volume 信息
// key: /volume/<id>
// val: VolumeInfo
type etcdVolumeInfo volume.VolumeInfo

// defaultCreate 把 volume 元数据存储到 etcd
func defaultCreate(js jsonstore.JsonStore, config volume.VolumeCreateConfig) (volume.VolumeInfo, error) {
	id, err := volume.NewVolumeID(config)
	if err != nil {
		return volume.VolumeInfo{}, err
	}
	info := volume.VolumeInfo{
		ID:        id.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Size:      config.Size,
		Type:      config.Type,
	}
	// TODO(zj): 这里先Get 后 Set 存在 并发问题
	// 后续在jsonstore中加 STM 的支持, 以方便使用，避免直接用etcd库的stm
	IDPath := mkIndexEtcdPath(volume.ETCDVolumeMetadataDir, id.String())
	infoToEtcd := etcdVolumeInfo(info)
	notfound, err := js.Notfound(context.Background(), IDPath)
	if err != nil {
		return volume.VolumeInfo{}, err
	}
	if !notfound {
		return volume.VolumeInfo{}, ExistedVolumeID
	}

	if err = js.Put(context.Background(), IDPath, infoToEtcd); err != nil {
		return volume.VolumeInfo{}, err
	}
	return info, nil
}

func defaultInfo(js jsonstore.JsonStore, ID volume.VolumeIdentity) (volume.VolumeInfo, error) {
	path := mkIndexEtcdPath(volume.ETCDVolumeMetadataDir, ID.String())
	var v etcdVolumeInfo
	if err := js.Get(context.Background(), path, &v); err != nil {
		if err == jsonstore.NotFoundErr {
			return volume.VolumeInfo{}, NotFoundVolume
		}
		return volume.VolumeInfo{}, err
	}
	return volume.VolumeInfo(v), nil
}

func defaultDelete(js jsonstore.JsonStore, ID volume.VolumeIdentity) (volume.VolumeInfo, error) {
	path := mkIndexEtcdPath(volume.ETCDVolumeMetadataDir, ID.String())
	var info etcdVolumeInfo
	if err := js.Get(context.Background(), path, &info); err != nil {
		if err == jsonstore.NotFoundErr {
			return volume.VolumeInfo{}, NotFoundVolume
		}
		return volume.VolumeInfo{}, err
	}
	var unused interface{}
	if err := js.Remove(context.Background(), path, &unused); err != nil {
		return volume.VolumeInfo{}, err
	}
	return volume.VolumeInfo(info), nil
}

// defaultUpdate 目前只会更新 VolumeInfo 中 References 值,
// 返回 更新前 的 volumeinfo
func defaultUpdate(js jsonstore.JsonStore, ID volume.VolumeIdentity, vlm volume.VolumeInfo) (volume.VolumeInfo, error) {
	path := mkIndexEtcdPath(volume.ETCDVolumeMetadataDir, ID.String())
	var info etcdVolumeInfo
	if err := js.Get(context.Background(), path, &info); err != nil {
		if err == jsonstore.NotFoundErr {
			return volume.VolumeInfo{}, NotFoundVolume
		}
		return volume.VolumeInfo{}, err
	}
	old := volume.VolumeInfo(info)
	info.UpdatedAt = time.Now()
	info.References = vlm.References
	var err error
	if err = js.Put(context.Background(), path, info); err != nil {
		return volume.VolumeInfo{}, err
	}
	return old, nil
}

func defaultSoftDelete(js jsonstore.JsonStore, ID volume.VolumeIdentity) (volume.VolumeInfo, error) {
	path := mkIndexEtcdPath(volume.ETCDVolumeMetadataDir, ID.String())
	var info etcdVolumeInfo
	if err := js.Get(context.Background(), path, &info); err != nil {
		if err == jsonstore.NotFoundErr {
			return volume.VolumeInfo{}, NotFoundVolume
		}
		return volume.VolumeInfo{}, err
	}
	info.DeletedAt = time.Now()
	if err := js.Put(context.Background(), path, info); err != nil {
		return volume.VolumeInfo{}, err
	}
	return volume.VolumeInfo(info), nil
}

func mkIndexEtcdPath(dir string, name string) string {
	return filepath.Join(dir, name)
}
