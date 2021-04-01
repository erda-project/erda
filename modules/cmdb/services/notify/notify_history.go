package notify

import "github.com/erda-project/erda/apistructs"

func (o *NotifyGroup) CreateNotifyHistory(request *apistructs.CreateNotifyHistoryRequest) (int64, error) {
	return o.db.CreateNotifyHistory(request)
}

func (o *NotifyGroup) QueryNotifyHistories(request *apistructs.QueryNotifyHistoryRequest) (*apistructs.QueryNotifyHistoryData, error) {
	return o.db.QueryNotifyHistories(request)
}
