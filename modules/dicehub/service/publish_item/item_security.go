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

package publish_item

import (
	"encoding/json"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
	"github.com/erda-project/erda/modules/dicehub/metrics"
	"github.com/erda-project/erda/providers/metrics/query"
)

// GetCertificationlist 获取验证列表
func (i *PublishItem) GetCertificationlist(req *apistructs.PublishItemCertificationListRequest,
	mk *apistructs.MonitorKeys) ([]*apistructs.PublishItemCertificationResponse, error) {
	return GetUserList(mk.AK, mk.AI, req)
}

func GetUserList(ak, ai string, req *apistructs.PublishItemCertificationListRequest) ([]*apistructs.PublishItemCertificationResponse, error) {
	logrus.Infof("start time: %d, end time: %d", req.StartTime, req.EndTime)
	endTime := time.Unix(int64(req.EndTime/1000), 0)
	startTime := time.Unix(int64(req.StartTime/1000), 0)

	// cfg := metrics.NewQueryConfig(metrics.WithSchema("http"), metrics.WithHost(discover.Monitor()))
	deviceReq := query.CreateQueryRequest("ta_metric_mobile_metrics")
	deviceReq = deviceReq.
		StartFrom(startTime).
		EndWith(endTime).
		Filter("tk", ak).
		Filter("ai", ai).
		Match("un", "?*").
		GroupBy([]string{"tags.uid"}).
		Apply("last", "tags.uid").
		Apply("last", "tags.cid").
		Apply("last", "tags.un").
		Apply("max", ".timestamp").
		Apply("sort", "max_.timestamp").
		Apply("limit", "1000")

	// client, err := metrics.NewClient().NewQuery(cfg)
	// if err != nil {
	// 	return nil, err
	// }
	logrus.Infof("monitor request params 获取认证列表数据: %s", deviceReq.ConstructParam().Encode())
	// resp, err := client.QueryMetric(deviceReq)
	// if err != nil {
	// 	return nil, err
	// }
	resp, err := metrics.Client.QueryMetric(deviceReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		logrus.Errorf("请求监控数据失败, statusCode: %d, body: %s", resp.StatusCode, string(resp.Body))
		return nil, errors.Errorf("请求监控数据失败, statusCode: %d, body: %s", resp.StatusCode, string(resp.Body))
	}
	var respBody apistructs.PublishItemMetricsCardinalityInterfaceResp
	logrus.Infof("monitor single resp data: %s", string(resp.Body))
	if err := json.Unmarshal(resp.Body, &respBody); err != nil {
		return nil, err
	}
	if len(respBody.Data.Results[0].Data) == 0 {
		return nil, nil
	}
	logrus.Infof("monitor single resp data respBody.Data.Results data len: %d", len(respBody.Data.Results[0].Data))
	result := make([]*apistructs.PublishItemCertificationResponse, 0, len(respBody.Data.Results[0].Data))
	for _, value := range respBody.Data.Results[0].Data {
		item := &apistructs.PublishItemCertificationResponse{}
		for k, v := range value {
			if strings.Contains(k, "tags.cid") {
				item.DeviceNo = v.Data.(string)
			}
			if strings.Contains(k, "tags.uid") {
				if v.Data == nil {
					item.UserID = ""
				} else {
					item.UserID = v.Data.(string)
				}
			}
			if strings.Contains(k, "tags.un") {
				if v.Data == nil {
					item.UserName = ""
				} else {
					m, err := url.ParseQuery(v.Data.(string))
					if err != nil {
						return nil, err
					}
					for k := range m {
						item.UserName = k
					}

				}
			}
			if strings.Contains(k, "max.timestamp") {
				a := int64(v.Data.(float64))
				item.LastLoginTime = time.Unix(a/int64(time.Second), a%int64(time.Second))
			}
		}
		if item.UserID == "" {
			continue
		}
		result = append(result, item)
	}

	return result, nil
}

// GetBlacklists 获取黑名单列表，分页
func (i *PublishItem) GetBlacklists(req *apistructs.PublishItemUserlistRequest) (*apistructs.PublishItemUserlistData, error) {
	_, err := i.checkPublishItemExsit(int64(req.PublishItemID))
	if err != nil {
		return nil, err
	}
	total, blacklists, err := i.db.GetBlacklists(req.PageNo, req.PageSize, req.PublishItemID)
	if err != nil {
		return nil, err
	}

	var resultData apistructs.PublishItemUserlistData
	if len(*blacklists) == 0 {
		return &resultData, nil
	}

	resultData.Total = total
	for _, v := range *blacklists {
		resultData.List = append(resultData.List, &apistructs.PublishItemUserListResponse{
			ID:        v.ID,
			UserID:    v.UserID,
			UserName:  v.UserName,
			DeviceNo:  v.DeviceNo,
			CreatedAt: v.CreatedAt,
		})
	}
	return &resultData, nil
}

// GetEraselists 获取数据擦除列表，分页
func (i *PublishItem) GetEraselists(req *apistructs.PublishItemUserlistRequest) (*apistructs.PublishItemUserlistData, error) {
	_, err := i.checkPublishItemExsit(int64(req.PublishItemID))
	if err != nil {
		return nil, err
	}
	total, blacklists, err := i.db.GetErases(req.PageNo, req.PageSize, req.PublishItemID)
	if err != nil {
		return nil, err
	}

	var resultData apistructs.PublishItemUserlistData
	if len(*blacklists) == 0 {
		return &resultData, nil
	}

	resultData.Total = total
	for _, v := range *blacklists {
		resultData.List = append(resultData.List, &apistructs.PublishItemUserListResponse{
			EraseStatus: v.EraseStatus,
			DeviceNo:    v.DeviceNo,
			CreatedAt:   v.UpdatedAt,
		})
	}
	return &resultData, nil
}

// AddBlacklist 添加黑名单
func (i *PublishItem) AddBlacklist(req *apistructs.PublishItemUserlistRequest) (error, *dbclient.PublishItem) {
	artifact, err := i.checkPublishItemExsit(int64(req.PublishItemID))
	if err != nil {
		return err, nil
	}
	if req.UserID != "" {
		blacklist, err := i.db.GetBlacklistByUserID(req.UserID, req.PublishItemID)
		if err != nil {
			return err, nil
		}
		if blacklist != nil {
			for _, v := range blacklist {
				if v.DeviceNo == req.DeviceNo {
					return errors.New("请勿重复添加"), nil
				}
			}
		}
	}
	if req.DeviceNo != "" {
		blacklist, err := i.db.GetBlacklistByDeviceNo(req.PublishItemID, req.DeviceNo)
		if err != nil {
			return err, nil
		}
		if blacklist != nil {
			for _, v := range blacklist {
				if v.UserID == req.UserID {
					return errors.New("请勿重复添加"), nil
				}
			}
		}
	}
	if err := i.db.CreateBlacklist(&dbclient.PublishItemBlackList{
		PublishItemID:  req.PublishItemID,
		PublishItemKey: artifact.AK,
		DeviceNo:       req.DeviceNo,
		UserID:         req.UserID,
		UserName:       req.UserName,
		Operator:       req.Operator,
	}); err != nil {
		return err, nil
	}
	return nil, artifact
}

// RemoveBlacklist 移除黑名单
func (i *PublishItem) RemoveBlacklist(blacklistID, publishItemID uint64) (error, *dbclient.PublishItemBlackList, *dbclient.PublishItem) {
	publishItem, err := i.checkPublishItemExsit(int64(publishItemID))
	if err != nil {
		return err, nil, nil
	}
	blacklist, err := i.db.GetBlacklistByID(blacklistID)
	if err != nil {
		return err, nil, nil
	}
	if blacklist != nil {
		if err := i.db.DeleteBlacklist(blacklist); err != nil {
			return err, nil, nil
		}
	}
	return nil, blacklist, publishItem
}

// AddErase 添加数据擦除
func (i *PublishItem) AddErase(req *apistructs.PublishItemUserlistRequest) (error, *dbclient.PublishItem) {
	artifact, err := i.checkPublishItemExsit(int64(req.PublishItemID))
	if err != nil {
		return err, nil

	}
	erase, err := i.db.GetEraseByDeviceNo(req.PublishItemID, req.DeviceNo)
	if err != nil {
		return err, nil
	}
	if erase == nil {
		if err := i.db.CreateErase(&dbclient.PublishItemErase{
			PublishItemID:  req.PublishItemID,
			PublishItemKey: artifact.AK,
			DeviceNo:       req.DeviceNo,
			EraseStatus:    apistructs.Erasing,
			Operator:       req.Operator,
		}); err != nil {
			return err, nil
		}
		return nil, artifact
	}
	if erase.EraseStatus == apistructs.Erasing {
		return errors.Errorf("请勿重复添加"), nil
	}

	erase.EraseStatus = apistructs.Erasing
	if err := i.db.UpdateErase(erase); err != nil {
		return err, nil
	}
	return nil, artifact
}

// UpdateErase 更新数据擦除
func (i *PublishItem) UpdateErase(request *apistructs.PublishItemEraseRequest) error {
	publishItem, err := i.GetPublishItemByAKAI(request.Ak, request.Ai)
	if err != nil {
		return err
	}
	if publishItem == nil {
		return nil
	}
	erase, err := i.db.GetEraseByDeviceNo(publishItem.ID, request.DeviceNo)
	if err != nil {
		return err
	}
	if erase == nil {
		return nil
	}
	if erase.EraseStatus != apistructs.EraseSuccess {
		erase.EraseStatus = request.EraseStatus
		if err := i.db.UpdateErase(erase); err != nil {
			logrus.Errorf("update erase status failure, deviceNo: %s, publishItemKey key: %s", request.DeviceNo, request.Ai)
			return err
		}
	}

	return nil
}

// GetSecurityStatus 获取客户安全信息状态
func (i *PublishItem) GetSecurityStatus(request apistructs.PublishItemSecurityStatusRequest) (*apistructs.PublishItemSecurityStatusResponse, error) {
	publishItem, err := i.GetPublishItemByAKAI(request.Ak, request.Ai)
	if err != nil {
		return nil, err
	}
	if publishItem == nil {
		return nil, errors.Errorf("not exist publish item")
	}
	inBlacklist := false
	if request.UserID != "" {
		blacklist, err := i.db.GetBlacklistByUserID(request.UserID, publishItem.ID)
		if err != nil {
			return nil, err
		}

		if len(blacklist) > 0 {
			for _, i := range blacklist {
				if i.DeviceNo == "" {
					inBlacklist = true
					break
				}
				if i.DeviceNo == request.DeviceNo {
					inBlacklist = true
					break
				}
			}
		}
	}

	if !inBlacklist {
		blacklist, err := i.db.GetBlacklistByDeviceNo(publishItem.ID, request.DeviceNo)
		if err != nil {
			return nil, err
		}
		if len(blacklist) > 0 {
			for _, i := range blacklist {
				if i.UserID == "" {
					inBlacklist = true
					break
				}
				if i.UserID == request.UserID {
					inBlacklist = true
				}
			}
		}
	}

	erase, err := i.db.GetEraseByDeviceNo(publishItem.ID, request.DeviceNo)
	if err != nil {
		return nil, err
	}
	eraseStatus := ""
	if erase != nil {
		eraseStatus = erase.EraseStatus
	}
	result := apistructs.PublishItemSecurityStatusResponse{
		NoJailbreak: publishItem.NoJailbreak,
		InBlacklist: inBlacklist,
		InEraseList: erase != nil,
		EraseStatus: eraseStatus,
	}
	if publishItem.GeofenceLon == 0.0 || publishItem.GeofenceLat == 0.0 {
		result.WithinGeofence = true
	} else {
		if request.Lon == 0.0 || request.Lat == 0.0 {
			result.WithinGeofence = false
		} else {
			dist := geoDistance(publishItem.GeofenceLon, publishItem.GeofenceLat, request.Lon, request.Lat)
			result.WithinGeofence = Smaller(dist*1000.0, publishItem.GeofenceRadius)
		}
	}

	return &result, nil
}

// checkPublishItemExsit 查看artifact是否存在
func (i *PublishItem) checkPublishItemExsit(publishItemID int64) (*dbclient.PublishItem, error) {
	publishItem, err := i.db.GetPublishItem(publishItemID)
	if err != nil {
		return nil, err
	}
	if publishItem == nil {
		return nil, errors.Errorf("non-existing publishItem information，publishItem id: %d", publishItemID)
	}
	return publishItem, nil
}

// GeoDistance 计算地理距离，依次为两个坐标的纬度、经度、单位（默认：英里，K => 公里，N => 海里）
func geoDistance(lng1 float64, lat1 float64, lng2 float64, lat2 float64) float64 {
	const PI float64 = 3.141592653589793

	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(PI * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515
	dist = dist * 1.609344

	return dist
}

func Smaller(a, b float64) bool {
	return math.Max(a, b) == b && math.Abs(a-b) > 0.00001
}
