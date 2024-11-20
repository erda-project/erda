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

package addon

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const (
	InfluxDBDataPath     = "/var/lib/influxdb2"
	InfluxDBConfHost     = "INFLUX_HOST"
	InfluxDBConfUser     = "INFLUX_USERNAME"
	InfluxDBConfPassword = "INFLUX_PASSWORD"
	InfluxDBConfORG      = "INFLUX_ORG"
	InfluxDBServicePort  = "8086"
)

const (
	InfluxDBKMSPasswordKey = "influxdb-password"
)

const (
	InfluxDBInitPrefix = "DOCKER_INFLUXDB_INIT_"

	InfluxDBInitUserNameKey = InfluxDBInitPrefix + "USERNAME"
	InfluxDBInitUserName    = "influxdb"
	InfluxDBInitPasswordKey = InfluxDBInitPrefix + "PASSWORD"

	// InfluxDBInitModeKey default init mode: setup
	InfluxDBInitModeKey = InfluxDBInitPrefix + "MODE"
	InfluxDBInitMode    = "setup"

	// InfluxDBInitOrgKey erda project name -> influxdb org
	InfluxDBInitOrgKey = InfluxDBInitPrefix + "ORG"

	// InfluxDBInitBucketKey erda application name -> influxdb bucket
	InfluxDBInitBucketKey = InfluxDBInitPrefix + "BUCKET"
	InfluxDBInitBucket    = "erda"

	// InfluxDBInitRetentionKey default 1w
	InfluxDBInitRetentionKey = InfluxDBInitPrefix + "RETENTION"
	InfluxDBInitRetention    = "1w"
)

const (
	InfluxDBParamsOrg       = "org"
	InfluxDBParamsBucket    = "bucket"
	InfluxDBParamsRetention = "retention"
)

// BuildInfluxDBServiceItem build influxdb service item
func (a *Addon) BuildInfluxDBServiceItem(params *apistructs.AddonHandlerCreateItem, addonIns *dbclient.AddonInstance,
	addonSpec *apistructs.AddonExtension, addonDice *diceyml.Object) error {
	addonDeployPlan := addonSpec.Plan[params.Plan]
	serviceMap := diceyml.Services{}

	// InfluxDB support 1 node only now.
	serviceItem := *addonDice.Services[addonSpec.Name]
	serviceItem.Resources = diceyml.Resources{
		CPU:    addonDeployPlan.CPU,
		MaxCPU: addonDeployPlan.MaxCPU,
		Mem:    addonDeployPlan.Mem,
		MaxMem: addonDeployPlan.MaxMem,
	}

	// init config render
	if err := a.influxDBInitRender(params, addonIns, serviceItem.Envs); err != nil {
		return err
	}

	// label
	if len(serviceItem.Labels) == 0 {
		serviceItem.Labels = map[string]string{}
	}
	serviceItem.Labels["ADDON_GROUP_ID"] = addonSpec.Name
	SetlabelsFromOptions(params.Options, serviceItem.Labels)
	// binding data
	vol := SetAddonVolumes(params.Options, InfluxDBDataPath, false)
	serviceItem.Volumes = diceyml.Volumes{vol}

	serviceMap[strings.Join([]string{addonSpec.Name, strconv.Itoa(0)}, "-")] = &serviceItem
	addonDice.Services = serviceMap

	return nil
}

func (a *Addon) InfluxDBDeployStatus(addonIns *dbclient.AddonInstance, serviceGroup *apistructs.ServiceGroup) (map[string]string, error) {
	password, err := a.db.GetByInstanceIDAndField(addonIns.ID, InfluxDBKMSPasswordKey)
	if err != nil {
		return nil, err
	}

	configMap := map[string]string{}
	for _, value := range serviceGroup.Services {
		configMap[InfluxDBConfHost] = fmt.Sprintf("http://%s:%s", value.Vip, InfluxDBServicePort)
		configMap[InfluxDBConfUser] = InfluxDBInitUserName
		configMap[InfluxDBConfORG] = genInfluxDBOrg(addonIns)
		configMap[InfluxDBConfPassword] = password.Value
	}

	return configMap, nil
}

func (a *Addon) influxDBInitRender(params *apistructs.AddonHandlerCreateItem,
	addonIns *dbclient.AddonInstance, serviceEnv diceyml.EnvMap) error {
	// Init mode
	serviceEnv[InfluxDBInitModeKey] = InfluxDBInitMode

	// Username
	serviceEnv[InfluxDBInitUserNameKey] = InfluxDBInitUserName

	// Password
	password, err := a.savePassword(addonIns, InfluxDBKMSPasswordKey)
	if err != nil {
		return err
	}
	serviceEnv[InfluxDBInitPasswordKey] = password

	// Org
	org := params.Options[InfluxDBParamsOrg]
	if org == "" {
		org = genInfluxDBOrg(addonIns)
	}
	serviceEnv[InfluxDBInitOrgKey] = org

	// Bucket
	bucket := params.Options[InfluxDBParamsBucket]
	if bucket == "" {
		bucket = InfluxDBInitBucket
	}
	serviceEnv[InfluxDBInitBucketKey] = bucket

	// Retention
	retention := params.Options[InfluxDBParamsRetention]
	if retention == "" {
		retention = InfluxDBInitRetention
	} else {
		_, err := time.ParseDuration(retention)
		if err != nil {
			return errors.Wrapf(err, "failed to parse retention time %s", retention)
		}
	}
	serviceEnv[InfluxDBInitRetentionKey] = retention

	return nil
}

func genInfluxDBOrg(addonIns *dbclient.AddonInstance) string {
	// ORG_PROJECT_WORKSPACE
	influxDBOrgSource := fmt.Sprintf("%s_%s_%s", addonIns.OrgID, addonIns.ProjectID, addonIns.Workspace)
	return fmt.Sprintf("%x", md5.Sum([]byte(influxDBOrgSource)))
}
