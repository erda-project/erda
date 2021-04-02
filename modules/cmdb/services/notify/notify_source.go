package notify

import "github.com/erda-project/erda/apistructs"

func (o *NotifyGroup) DeleteNotifySource(request *apistructs.DeleteNotifySourceRequest) error {
	return o.db.DeleteNotifySource(request)
}
