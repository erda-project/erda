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

package driver

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/impl/volume"
	"github.com/erda-project/erda/pkg/jsonstore"
)

var (
	BadVolumeTypeNotLocalVolume = errors.New("Bad VolumeType, not localvolume")
	BadAttachDest               = errors.New("Bad Attach dest")
	MultiAttachWithLocalVolume  = errors.New("multiple attach localvolume")
)

const (
	LocalVolumeHostPathPrefix = "/data/volumes"
)

type LocalVolumeDriver struct {
	js jsonstore.JsonStore
}

func NewLocalVolumeDriver(js jsonstore.JsonStore) volume.Volume {
	return &LocalVolumeDriver{js}
}

func (d LocalVolumeDriver) Type() volume.VolumeType {
	return apistructs.LocalVolume
}

func (d LocalVolumeDriver) Create(config volume.VolumeCreateConfig) (volume.VolumeInfo, error) {
	if config.Type != apistructs.LocalVolume {
		return volume.VolumeInfo{}, BadVolumeTypeNotLocalVolume
	}
	return defaultCreate(d.js, config)
}

func (d LocalVolumeDriver) Info(ID volume.VolumeIdentity) (volume.VolumeInfo, error) {
	return defaultInfo(d.js, ID)
}

func (d LocalVolumeDriver) Attach(ID volume.VolumeIdentity, dst volume.AttachDest) (volume.AttachCallback, error) {
	if err := dst.Validate(); err != nil {
		return nil, errors.Wrap(BadAttachDest, err.Error())
	}
	info, err := defaultInfo(d.js, ID)
	if err != nil {
		return nil, err
	}

	if info.Type != apistructs.LocalVolume {
		return nil, BadVolumeTypeNotLocalVolume
	}

	ref := volume.VolumeReference{apistructs.AttachDest(dst)}
	info.References = append(info.References, ref)

	_, err = defaultUpdate(d.js, ID, info)
	if err != nil {
		return nil, err
	}

	cb := func(runtime *apistructs.ServiceGroup) (volume.VolumeInfo, error) {
		for i, s := range runtime.Services {
			if s.Name == dst.Service {
				for idx := range s.Volumes {
					runtime.Services[i].Volumes[idx].VolumePath = mkLocalVolumeHostPath(info.ID)
				}
			}
		}
		return info, nil
	}

	return cb, nil
}

func (d LocalVolumeDriver) UnAttach(ID volume.VolumeIdentity, dst volume.AttachDest) (volume.VolumeInfo, error) {
	if err := dst.Validate(); err != nil {
		return volume.VolumeInfo{}, errors.Wrap(BadAttachDest, err.Error())
	}
	info, err := defaultInfo(d.js, ID)
	if err != nil {
		return volume.VolumeInfo{}, err
	}

	newReferences := []volume.VolumeReference{}
	for _, ref := range info.References {
		if !volume.AttachDest(ref.Info).Equal(dst) {
			newReferences = append(newReferences, ref)
		}
	}
	info.References = newReferences
	_, err = defaultUpdate(d.js, ID, info)
	if err != nil {
		return volume.VolumeInfo{}, err
	}
	return info, nil
}

// For localvolume, Delete does nothing, only clears the metadata. The specific cleanup work is implemented by the localvolume provided by the plugin.
// For example, for marathon, localpv is cleaned up by it
func (d LocalVolumeDriver) Delete(ID volume.VolumeIdentity, force bool) error {
	_, err := defaultDelete(d.js, ID)
	return err
}

func mkLocalVolumeHostPath(ID string) string {
	return ID
}
