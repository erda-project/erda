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
	"github.com/erda-project/erda/apistructs"
)

type VolumeImpl struct {
	drivers map[apistructs.VolumeType]Volume
}

func NewVolumeImpl(drivers map[apistructs.VolumeType]Volume) Volume {
	return &VolumeImpl{drivers}
}

func (VolumeImpl) Type() VolumeType {
	panic("shouldn't call me")
}

func (i VolumeImpl) Create(config VolumeCreateConfig) (VolumeInfo, error) {
	return i.drivers[config.Type].Create(config)
}

func (i VolumeImpl) Info(ID VolumeIdentity) (VolumeInfo, error) {
	driver, err := GetDriver(i.drivers, ID)
	if err != nil {
		return VolumeInfo{}, err
	}
	return driver.Info(ID)
}

func (i VolumeImpl) Attach(ID VolumeIdentity, dst AttachDest) (AttachCallback, error) {
	driver, err := GetDriver(i.drivers, ID)
	if err != nil {
		return nil, err
	}
	return driver.Attach(ID, dst)
}
func (i VolumeImpl) UnAttach(ID VolumeIdentity, dst AttachDest) (VolumeInfo, error) {
	driver, err := GetDriver(i.drivers, ID)
	if err != nil {
		return VolumeInfo{}, err
	}
	return driver.UnAttach(ID, dst)
}

func (i VolumeImpl) Delete(ID VolumeIdentity, force bool) error {
	driver, err := GetDriver(i.drivers, ID)
	if err != nil {
		return err
	}
	return driver.Delete(ID, force)
}

func GetDriver(drivers map[VolumeType]Volume, ID VolumeIdentity) (Volume, error) {
	tp, err := DecodeVolumeType(string(ID))
	if err != nil {
		return nil, err
	}
	return drivers[tp], nil

}
