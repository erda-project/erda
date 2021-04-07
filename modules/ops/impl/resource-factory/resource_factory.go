package resource_factory

import (
	"encoding/json"
	"runtime/debug"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/dbclient"
	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
	"github.com/erda-project/erda/modules/ops/impl/aliyun-resources/vpc"
)

type BaseResourceMaterial interface {
	GetVendor() string
	SetVendor(string)
	GetRegion() string
	SetRegion(string)
	GetVpcID() string
	SetVpcID(string)
	GetVSwitchID() string
	SetVSwitchID(string)
	GetZoneID() string
	SetZoneID(string)

	GetOrgID() string
	GetUserID() string
	GetClusterName() string
	GetProjectID() string
	GetSource() string
	GetClientToken() string

	GetInstanceName() string
	GetAddonID() string
}

type ResourceCreator func(aliyun_resources.Context, BaseResourceMaterial, *dbclient.Record, *apistructs.CreateCloudResourceRecord, apistructs.CloudResourceVpcBaseInfo) (*apistructs.AddonConfigCallBackResponse, *dbclient.ResourceRouting, error)

type ResourceFactory interface {
	GetDbClient() *dbclient.DBClient
	SetDbClient(*dbclient.DBClient)
	GetRecordType() dbclient.RecordType
	GetCreator() ResourceCreator
	CreateResource(aliyun_resources.Context, BaseResourceMaterial) (*dbclient.Record, error)
}

type BaseResourceFactory struct {
	RecordType dbclient.RecordType
	DBClient   *dbclient.DBClient
	Creator    ResourceCreator
}

var resourceFactories = make(map[dbclient.ResourceType]ResourceFactory)

func Register(resourceType dbclient.ResourceType, factory ResourceFactory) error {
	if _, exist := resourceFactories[resourceType]; exist {
		return errors.Errorf("factory already exists, resourceType: %s", resourceType)
	}
	resourceFactories[resourceType] = factory
	logrus.Infof("register resource factory for resource type: %s, factory:%+v", resourceType, resourceFactories)
	return nil
}

func GetResourceFactory(dbClient *dbclient.DBClient, resourceType dbclient.ResourceType) (ResourceFactory, error) {
	if dbClient == nil {
		return nil, errors.New("invalid dbClient")
	}

	factory, ok := resourceFactories[resourceType]
	if !ok {
		return nil, errors.Errorf("resource factory not found, type: %s", resourceType)
	}
	factory.SetDbClient(dbClient)
	return factory, nil
}

func (obj BaseResourceFactory) GetDbClient() *dbclient.DBClient {
	return obj.DBClient
}

func (obj *BaseResourceFactory) SetDbClient(dbclient *dbclient.DBClient) {
	obj.DBClient = dbclient
}

func (obj BaseResourceFactory) GetRecordType() dbclient.RecordType {
	return obj.RecordType
}

func (obj BaseResourceFactory) GetCreator() ResourceCreator {
	return obj.Creator
}

func (obj *BaseResourceFactory) SetRecordType(t dbclient.RecordType) {
	obj.RecordType = t
}

func (obj BaseResourceFactory) sourceCheck(m BaseResourceMaterial) error {
	source := m.GetSource()
	switch source {
	case apistructs.CloudResourceSourceResource:
		if m.GetRegion() == "" {
			return errors.New("request come from [resource] failed, missing param: [region]")
		}
	case apistructs.CloudResourceSourceAddon:
		if m.GetRegion() == "" && m.GetClusterName() == "" {
			return errors.New("request come from [addon] failed, both region and clusterName is empty")
		}
	default:
		return errors.Errorf("request failed, invalide param, source: %s, only support:[addon, resource)] ", source)
	}
	return nil
}

func (obj BaseResourceFactory) initRecord(m BaseResourceMaterial) (*dbclient.Record, error) {
	record := &dbclient.Record{
		RecordType:  obj.GetRecordType(),
		UserID:      m.GetUserID(),
		OrgID:       m.GetOrgID(),
		ClusterName: m.GetClusterName(),
		Status:      dbclient.StatusTypeProcessing,
	}
	recordID, err := obj.GetDbClient().RecordsWriter().Create(record)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write record")
	}
	records, err := obj.GetDbClient().RecordsReader().ByIDs(strconv.FormatUint(recordID, 10)).Do()
	if err != nil {
		return nil, errors.Wrap(err, "failed to query records")
	}
	if len(records) == 0 {
		return nil, errors.Wrapf(err, "failed to query records, empty record, record id: %v", recordID)
	}
	return &records[0], nil
}

func doRecover() {
	if r := recover(); r != nil {
		logrus.Errorf("recovered from: %+v ", r)
		debug.PrintStack()
	}
}

func fillRoutingInfo(routing *dbclient.ResourceRouting, m BaseResourceMaterial, r *dbclient.Record, addonID string) {
	routing.Vendor = m.GetVendor()
	routing.OrgID = m.GetOrgID()
	routing.ClusterName = m.GetClusterName()
	routing.ProjectID = m.GetProjectID()
	routing.AddonID = addonID
	routing.Status = dbclient.ResourceStatusAttached
	routing.RecordID = r.ID
}

// CreateResource 创建资源
func (obj BaseResourceFactory) CreateResource(ctx aliyun_resources.Context, m BaseResourceMaterial) (*dbclient.Record, error) {
	err := obj.sourceCheck(m)
	if err != nil {
		return nil, err
	}
	r, err := obj.initRecord(m)
	if err != nil {
		return nil, err
	}
	d := apistructs.CreateCloudResourceRecord{
		ClientToken:  m.GetClientToken(),
		InstanceName: m.GetInstanceName(),
	}
	createInstanceStep := apistructs.CreateCloudResourceStep{
		Step:   string(obj.GetRecordType()),
		Status: string(dbclient.StatusTypeSuccess),
	}
	d.Steps = append(d.Steps, createInstanceStep)
	d.Steps[len(d.Steps)-1].Name = m.GetInstanceName()

	creator := obj.GetCreator()
	if creator == nil {
		aliyun_resources.ProcessFailedRecord(ctx, m.GetSource(), m.GetClientToken(), r, &d, err)
		return nil, errors.New("invalid creator")
	}
	go func() {
		var (
			err     error
			vpcinfo apistructs.CloudResourceVpcBaseInfo
		)

		defer doRecover()
		defer func() {
			if err != nil {
				aliyun_resources.ProcessFailedRecord(ctx, m.GetSource(), m.GetClientToken(), r, &d, err)
				logrus.Error(err)
			}
		}()

		// get vpc info
		// 来自addon的请求，不带region，通过cluster name，查找vpc，得到region、cidr等信息
		// 来自云管的请求
		//	有些带 region和vpc id，据此查询更详细的cidr，zoneID等信息
		//	有些如oss，ons，则只带有region信息，只保留region就可以了
		if m.GetZoneID() == "" {
			ctx.Region = m.GetRegion()
			ctx.VpcID = m.GetVpcID()
			// 对于需要vpc信息的资源，获取vpc信息
			v, e := vpc.GetVpcBaseInfo(ctx, m.GetClusterName(), m.GetVpcID())
			if e != nil {
				err = e
				return
			}
			// 有些资源并不需要详细的vpc信息（返回v为空），只需要region信息
			if m.GetRegion() == "" {
				m.SetRegion(v.Region)
			}
			m.SetVpcID(v.VpcID)
			m.SetVSwitchID(v.VSwitchID)
			m.SetZoneID(v.ZoneID)
			vpcinfo = v
		}
		ctx.Region = m.GetRegion()

		// create resource
		config, routing, err := creator(ctx, m, r, &d, vpcinfo)
		if err != nil {
			return
		}

		if config != nil {
			// resource config callback
			addonID := m.GetAddonID()
			if addonID == "" {
				addonID = m.GetClientToken()
			}
			logrus.Infof("start addon config callback, addonID: %s, request: %+v", addonID, m)
			_, err = ctx.Bdl.AddonConfigCallback(addonID, *config)
			if err != nil {
				err = errors.Wrapf(err, "addon call back config failed, addonID: %s", addonID)
				return
			}
			_, err = ctx.Bdl.AddonConfigCallbackProvison(addonID, apistructs.AddonCreateCallBackResponse{IsSuccess: true})
			if err != nil {
				err = errors.Wrapf(err, "addon call back provision failed, addonId: %s", addonID)
				return
			}

			// resource routing record
			fillRoutingInfo(routing, m, r, addonID)
			_, err = ctx.DB.ResourceRoutingWriter().Create(routing)
			if err != nil {
				err = errors.Wrap(err, "update resource routing failed")
				return
			}
		}

		// update ops record [success]
		content, err := json.Marshal(d)
		if err != nil {
			err = errors.Wrapf(err, "marshal record detail, failed, detail: %+v", d)
		}
		r.Status = dbclient.StatusTypeSuccess
		r.Detail = string(content)
		err = ctx.DB.RecordsWriter().Update(*r)
		if err != nil {
			err = errors.Wrap(err, "failed to update record")
		}
	}()
	return r, nil
}
