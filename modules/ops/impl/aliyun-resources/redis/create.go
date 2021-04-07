package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	kvstore "github.com/aliyun/alibaba-cloud-sdk-go/services/r_kvstore"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/dbclient"
	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
	"github.com/erda-project/erda/modules/ops/impl/aliyun-resources/vpc"
	"github.com/erda-project/erda/pkg/uuid"
)

func CreateInstanceWithRecord(ctx aliyun_resources.Context, req apistructs.CreateCloudResourceRedisRequest, record *dbclient.Record) {
	var detail apistructs.CreateCloudResourceRecord

	// create instance step
	createInstanceStep := apistructs.CreateCloudResourceStep{
		Step:   string(dbclient.RecordTypeCreateAliCloudRedis),
		Status: string(dbclient.StatusTypeSuccess)}
	detail.Steps = append(detail.Steps, createInstanceStep)
	detail.ClientToken = req.ClientToken
	detail.InstanceName = req.InstanceName
	detail.Steps[len(detail.Steps)-1].Name = req.InstanceName

	// 重名检查
	regionids := aliyun_resources.ActiveRegionIDs(ctx)
	list, err := List(ctx, aliyun_resources.DefaultPageOption, regionids.ECS, "")
	if err != nil {
		err := fmt.Errorf("list redis failed, error:%v", err)
		logrus.Errorf("%s, request:%+v", err.Error(), req)
		aliyun_resources.ProcessFailedRecord(ctx, req.Source, req.ClientToken, record, &detail, err)
		return
	}
	for _, m := range list {
		if req.InstanceName == m.InstanceName {
			err := fmt.Errorf("redis instance already exist, region:%s, name:%s", m.RegionId, m.InstanceName)
			logrus.Errorf("%s, request:%+v", err.Error(), req)
			aliyun_resources.ProcessFailedRecord(ctx, req.Source, req.ClientToken, record, &detail, err)
			return
		}
	}

	// 来自addon的请求，不带region，通过cluster name，查找vpc，得到region、cidr等信息
	// 来自云管的请求，带region和vpc id，据此查询更详细的cidr，zoneID等信息
	if req.ZoneID == "" {
		ctx.Region = req.Region
		ctx.VpcID = req.VpcID
		v, err := vpc.GetVpcBaseInfo(ctx, req.ClusterName, req.VpcID)
		if err != nil {
			err := fmt.Errorf("get vpc info failed, error:%v", err)
			logrus.Errorf("%s, request:%+v", err.Error(), req)
			aliyun_resources.ProcessFailedRecord(ctx, req.Source, req.ClientToken, record, &detail, err)
			return
		}
		req.Region = v.Region
		req.VpcID = v.VpcID
		req.VSwitchID = v.VSwitchID
		req.ZoneID = v.ZoneID
	}
	ctx.Region = req.Region

	// auto generate password if not provide
	if req.Password == "" {
		req.Password = uuid.UUID()[:8] + "r@1" + uuid.UUID()[:8]
	}
	// auto generate AutoRenewPeriod by ChargePeriod
	if strings.ToLower(req.ChargeType) == aliyun_resources.ChargeTypePrepaid {
		p, err := strconv.Atoi(req.ChargePeriod)
		if err != nil {
			err := fmt.Errorf("invalid charge period, support format:%s, (month)", "1-9，12，24，36")
			logrus.Errorf("%s, request:%+v", err.Error(), req)
			aliyun_resources.ProcessFailedRecord(ctx, req.Source, req.ClientToken, record, &detail, err)
			return
		}
		if p >= 12 {
			req.AutoRenewPeriod = "12"
		} else if p >= 6 {
			req.AutoRenewPeriod = "6"
		} else if p >= 3 {
			req.AutoRenewPeriod = "3"
		} else if p <= 0 {
			req.AutoRenewPeriod = "1"
		}
	}

	logrus.Debugf("start to create instance, request: %+v", req)
	r, err := CreateInstance(ctx, req)
	if err != nil {
		err := fmt.Errorf("create redis instance failed, error:%v", err)
		logrus.Errorf("%s, request:%+v", err.Error(), req)
		aliyun_resources.ProcessFailedRecord(ctx, req.Source, req.ClientToken, record, &detail, err)
		return
	}
	detail.InstanceID = r.InstanceId

	// request come from addon
	if req.Source == apistructs.CloudResourceSourceAddon {
		// post resource config info to orchestrator
		cb := apistructs.AddonConfigCallBackResponse{
			Config: []apistructs.AddonConfigCallBackItemResponse{
				{
					Name:  "REDIS_HOST",
					Value: r.ConnectionDomain,
				},
				{
					Name:  "REDIS_PORT",
					Value: r.Port,
				},
				{
					Name:  "REDIS_PASSWORD",
					Value: req.Password,
				},
			},
		}

		// TODO: only support one addon in a request
		if req.AddonID == "" {
			req.AddonID = req.ClientToken
		}

		logrus.Debugf("start to addon config callback, addonid:%s", req.AddonID)
		_, err := ctx.Bdl.AddonConfigCallback(req.AddonID, cb)
		if err != nil {
			err := fmt.Errorf("redis addon call back failed, error: %v", err)
			logrus.Errorf(err.Error())
			aliyun_resources.ProcessFailedRecord(ctx, req.Source, req.ClientToken, record, &detail, err)
			return
		}

		_, err = ctx.Bdl.AddonConfigCallbackProvison(req.AddonID, apistructs.AddonCreateCallBackResponse{IsSuccess: true})
		if err != nil {
			err := fmt.Errorf("add call back provision failed, error:%v", err)
			logrus.Errorf(err.Error())
			aliyun_resources.ProcessFailedRecord(ctx, req.Source, req.ClientToken, record, &detail, err)
			return
		}

		// update resource routing record
		_, err = ctx.DB.ResourceRoutingWriter().Create(&dbclient.ResourceRouting{
			ResourceID:   r.InstanceId,
			ResourceName: req.InstanceName,
			ResourceType: dbclient.ResourceTypeRedis,
			Vendor:       req.Vendor,
			OrgID:        req.OrgID,
			ClusterName:  req.ClusterName,
			ProjectID:    req.ProjectID,
			AddonID:      req.AddonID,
			Status:       dbclient.ResourceStatusAttached,
			RecordID:     record.ID,
			Detail:       "",
		})
		if err != nil {
			err := fmt.Errorf("write resource routing to db failed, error:%v", err)
			logrus.Errorf(err.Error())
			aliyun_resources.ProcessFailedRecord(ctx, req.Source, req.ClientToken, record, &detail, err)
			return
		}
	}

	// success, update ops_record
	content, err := json.Marshal(detail)
	if err != nil {
		logrus.Errorf("marshal record detail failed, error:%+v", err)
	}
	record.Status = dbclient.StatusTypeSuccess
	record.Detail = string(content)
	if err := ctx.DB.RecordsWriter().Update(*record); err != nil {
		logrus.Errorf("failed to update record: %v", err)
	}
}

func CreateInstance(ctx aliyun_resources.Context, req apistructs.CreateCloudResourceRedisRequest) (*kvstore.CreateInstanceResponse, error) {
	client, err := kvstore.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create client error: %+v", err)
		return nil, err
	}

	request := kvstore.CreateCreateInstanceRequest()
	request.Scheme = "https"

	request.Token = req.ClientToken
	request.InstanceName = req.InstanceName
	// TODO: auto generate password, DONE
	request.Password = req.Password
	request.InstanceClass = req.Spec
	request.NetworkType = "VPC"
	if req.VpcID != "" {
		request.VpcId = req.VpcID
	}
	if req.VSwitchID != "" {
		request.VSwitchId = req.VSwitchID
	}
	if req.ZoneID != "" {
		request.ZoneId = req.ZoneID
	}
	// charge
	request.ChargeType = req.ChargeType
	if strings.ToLower(req.ChargeType) == aliyun_resources.ChargeTypePrepaid {
		request.Period = req.ChargePeriod
		// 1-9，12，24，36
		request.AutoRenew = strconv.FormatBool(req.AutoRenew)
		// TODO: auto generate from charge period, DONE
		// 1 2 3 6 12
		request.AutoRenewPeriod = req.ChargePeriod
	}

	request.InstanceType = "Redis"
	request.EngineVersion = req.Version

	response, err := client.CreateInstance(request)
	if err != nil {
		e := fmt.Errorf("create redis instance failed, error:%v", err)
		logrus.Errorf("%s, request:%+v", e.Error(), req)
		return nil, e
	}
	return response, nil
}
