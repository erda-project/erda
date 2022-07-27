// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
)

//TODO webhook

const (
	ModeClassic = "Classic"
	ModeSingle  = "Single"
	ModeMulti   = "Multi"
)

func (r *Mysql) Default() {
	if r.Spec.Version == "" {
		r.Spec.Version = "v5.7"
	}

	if r.Spec.PrimaryMode == "" {
		r.Spec.PrimaryMode = ModeClassic
	}

	if r.Spec.Primaries == 0 {
		r.Spec.Primaries = 1
	}

	if r.Spec.PrimaryMode == ModeClassic {
		if r.Spec.Replicas == nil {
			r.Spec.Replicas = pointer.IntPtr(1)
		}

		if r.Spec.PrimaryId == nil {
			r.Spec.PrimaryId = pointer.IntPtr(0)
		}

		if r.Spec.AutoSwitch == nil {
			r.Spec.AutoSwitch = pointer.BoolPtr(true)
		}
	} else {
		if r.Spec.Replicas == nil {
			r.Spec.Replicas = pointer.IntPtr(0)
		}

		if r.Spec.PrimaryId == nil {
			r.Spec.PrimaryId = pointer.IntPtr(-1)
		}

		if r.Spec.AutoSwitch == nil {
			r.Spec.AutoSwitch = pointer.BoolPtr(false)
		}

		if r.Spec.GroupName == "" {
			r.Spec.GroupName = uuid.New().String()
		}
	}

	//TODO secret
	if r.Spec.LocalUsername == "" {
		r.Spec.LocalUsername = "root"
	}
	if r.Spec.LocalPassword == "" {
		r.Spec.LocalPassword = GeneratePassword(29)
	}
	if r.Spec.ReplicaUsername == "" {
		r.Spec.ReplicaUsername = "repl"
	}
	if r.Spec.ReplicaPassword == "" {
		r.Spec.ReplicaPassword = GeneratePassword(29)
	}

	if r.Spec.StorageClassName == "" {
		r.Spec.StorageClassName = "standard"
	}
	if r.Spec.StorageSize.IsZero() {
		r.Spec.StorageSize = resource.MustParse("10Gi")
	}

	if r.Spec.Image == "" {
		r.Spec.Image = "registry.erda.cloud/erda-addons/mylet:" + r.Spec.Version
	}
	if r.Spec.ImagePullPolicy == "" {
		r.Spec.ImagePullPolicy = corev1.PullIfNotPresent
	}

	if r.Spec.Port == 0 {
		r.Spec.Port = 3306
	}
	if r.Spec.MyletPort == 0 {
		r.Spec.MyletPort = 33080
	}
	if r.Spec.GroupPort == 0 {
		r.Spec.GroupPort = 33061
	}
	if r.Spec.ExporterPort == 0 {
		r.Spec.ExporterPort = 9104
	}

	if r.Spec.Mydir == "" {
		r.Spec.Mydir = "/mydir"
	}

	const MyctlPort = 33081
	if r.Spec.MyctlAddr == "" {
		r.Spec.MyctlAddr = "myctl." + Namespace + ".svc.cluster.local:" + strconv.Itoa(MyctlPort)
	}
	if r.Spec.HeadlessHost == "" {
		r.Spec.HeadlessHost = r.BuildName(HeadlessSuffix) + "." + r.Namespace + ".svc.cluster.local"
		r.Spec.ShortHeadlessHost = r.BuildName(HeadlessSuffix) + "." + r.Namespace
	}

	if r.Spec.EnableExporter {
		if r.Spec.ExporterImage == "" {
			r.Spec.ExporterImage = "registry.erda.cloud/retag/mysqld-exporter:v0.14.0"
		}
		if r.Spec.ExporterUsername == "" {
			r.Spec.ExporterUsername = "exporter"
		}
		if r.Spec.ExporterPassword == "" {
			r.Spec.ExporterPassword = GeneratePassword(29)
		}
	}
}

func (r *Mysql) Validate() (err error) {
	r.Status.Version, err = ParseVersion(r.Spec.Version)
	if err != nil {
		return err
	}

	if !Between(r.Spec.Primaries, 1, 9) {
		return fmt.Errorf("primaries not in [1, 9]: %d", r.Spec.Primaries)
	}
	if !Between(*r.Spec.Replicas, 0, 9) {
		return fmt.Errorf("replicas not in [0, 9]: %d", *r.Spec.Replicas)
	}

	switch r.Spec.PrimaryMode {
	case ModeClassic:
		if r.Spec.Primaries != 1 {
			return fmt.Errorf("%s mode primaries must equal 1", r.Spec.PrimaryMode)
		}

		if !Between(*r.Spec.PrimaryId, 0, *r.Spec.Replicas) {
			return fmt.Errorf("%s mode primary id not in [0, %d]: %d", r.Spec.PrimaryMode, *r.Spec.Replicas, *r.Spec.PrimaryId)
		}
	case ModeSingle, ModeMulti:
		if r.Spec.Primaries <= 1 {
			return fmt.Errorf("%s mode primaries must greater than 1", r.Spec.PrimaryMode)
		}

		if *r.Spec.PrimaryId != -1 {
			return fmt.Errorf("%s mode primary id unwanted", r.Spec.PrimaryMode)
		}

		if *r.Spec.AutoSwitch {
			return fmt.Errorf("%s mode auto switch unwanted", r.Spec.PrimaryMode)
		}

		_, err = uuid.Parse(r.Spec.GroupName)
		if err != nil {
			return fmt.Errorf("group name invalid: %s", err.Error())
		}
	default:
		return fmt.Errorf("primary mode invalid: %s", r.Spec.PrimaryMode)
	}

	if r.Spec.LocalUsername == "" {
		return fmt.Errorf("local username required")
	}
	if r.Spec.LocalPassword == "" {
		return fmt.Errorf("local password required")
	}
	if r.Spec.ReplicaUsername == "" {
		return fmt.Errorf("replica username required")
	}
	if r.Spec.ReplicaPassword == "" {
		return fmt.Errorf("replica password required")
	}
	if r.Spec.LocalUsername == r.Spec.ReplicaUsername {
		return fmt.Errorf("local username must not equal replica username")
	}
	if HasQuote(r.Spec.LocalUsername, r.Spec.LocalPassword, r.Spec.ReplicaUsername, r.Spec.ReplicaPassword) {
		return fmt.Errorf("username and password must not contains any quotation marks")
	}

	if !Between(r.Spec.Port, minPort, maxPort) {
		return fmt.Errorf("port not in [%d, %d]: %d", minPort, maxPort, r.Spec.Port)
	}
	if !Between(r.Spec.MyletPort, minPort, maxPort) {
		return fmt.Errorf("mylet port not in [%d, %d]: %d", minPort, maxPort, r.Spec.MyletPort)
	}
	if !Between(r.Spec.GroupPort, minPort, maxPort) {
		return fmt.Errorf("group port not in [%d, %d]: %d", minPort, maxPort, r.Spec.GroupPort)
	}
	if !Between(r.Spec.ExporterPort, minPort, maxPort) {
		return fmt.Errorf("exporter port not in [%d, %d]: %d", minPort, maxPort, r.Spec.ExporterPort)
	}
	if HasEqual(r.Spec.Port, r.Spec.MyletPort, r.Spec.GroupPort, r.Spec.ExporterPort) {
		return fmt.Errorf("ports must not equal")
	}

	if r.Spec.Mydir == "" {
		return fmt.Errorf("mydir required")
	}
	if host, _ := SplitHostPort(r.Spec.MyctlAddr); host == "" {
		return fmt.Errorf("myctl addr invalid: %s", r.Spec.MyctlAddr)
	}
	if host, port := SplitHostPort(r.Spec.HeadlessHost); host == "" {
		return fmt.Errorf("headless host invalid: %s", r.Spec.HeadlessHost)
	} else if port != "" {
		return fmt.Errorf("headless host must not contains port: %s", r.Spec.HeadlessHost)
	}
	if r.Spec.ShortHeadlessHost != "" {
		if host, port := SplitHostPort(r.Spec.ShortHeadlessHost); host == "" {
			return fmt.Errorf("short headless host invalid: %s", r.Spec.ShortHeadlessHost)
		} else if port != "" {
			return fmt.Errorf("short headless host must not contains port: %s", r.Spec.ShortHeadlessHost)
		}
	}

	if r.Spec.EnableExporter {
		if r.Spec.ExporterUsername == "" {
			return fmt.Errorf("exporter username required")
		}
		if r.Spec.ExporterPassword == "" {
			return fmt.Errorf("exporter password required")
		}
		if r.Spec.ExporterUsername == r.Spec.LocalUsername || r.Spec.ExporterUsername == r.Spec.ReplicaUsername {
			return fmt.Errorf("exporter username must not equal local username or replica username")
		}
		if HasQuote(r.Spec.ExporterUsername, r.Spec.ExporterPassword) {
			return fmt.Errorf("exporter username and password must not contains any quotation marks")
		}
	}

	n := r.Spec.Size()
	r.Status.Solos = make([]MysqlSolo, n)
	for i := 0; i < n; i++ {
		r.Status.Solos[i], err = r.NormalizeSolo(i)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	serverId = 729
	minPort  = 1
	maxPort  = 65535
)
