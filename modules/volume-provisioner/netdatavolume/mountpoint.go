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

package netdatavolume

import (
	"bufio"
	"errors"
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/strutil"
)

var (
	NotFoundErr            = errors.New("not found mountpoint")
	MultipleAlternativeErr = errors.New("multiple alternative mountpoint found")
)

// DiscoverMountPoint auto discover netdata's mountpoint
// How to parse mountInfo please refer to 'http://man7.org/linux/man-pages/man5/proc.5.html'
func DiscoverMountPoint() (string, error) {
	mountInfo, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(mountInfo)
	mountInfoList := []string{}
	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		mountInfoList = append(mountInfoList, l)

	}

	discovered := []string{}

	for _, l := range mountInfoList {
		splited := strutil.Split(l, " ", true)
		if len(splited) < 9 {
			continue
		}
		// splited[8] is fstype
		// Currently only consider fstype to be 'fuse.glusterfs' or 'nfs*'
		if splited[3] == "/" && (strutil.Contains(splited[8], "nfs") ||
			strutil.Contains(splited[8], "fuse.glusterfs")) {
			discovered = append(discovered, splited[4])
		}
	}
	if len(discovered) == 0 {
		return "", NotFoundErr
	}
	if len(discovered) > 1 {
		logrus.Errorf("found multiple alternative: %v", discovered)
		return "", MultipleAlternativeErr
	}
	return discovered[0], nil
}
