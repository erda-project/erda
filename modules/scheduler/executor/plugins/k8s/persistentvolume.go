package k8s

import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

func (k *Kubernetes) DeletePV(sg *apistructs.ServiceGroup) error {
	if !IsGroupStateful(sg) {
		return nil
	}
	// todo:
	for _, service := range sg.Services {
		for _, bind := range service.Binds {
			hostPath := bind.HostPath
			// 找到本地盘
			if strings.HasPrefix(hostPath, "/") || len(hostPath) == 0 {
				continue
			}
			// todo: pv名字规则由某个函数统一生产
			pvName := strutil.Concat("lp-", sg.ID, "-")
			if len(hostPath) > 8 {
				pvName = strutil.Concat(pvName, hostPath[:8])
			} else {
				pvName = strutil.Concat(pvName, hostPath)
			}

			// todo: 确认该pv与该runtime下的service的对应的pvc绑定
			list, err := k.pv.List(pvName)
			if err != nil {
				logrus.Errorf("failed to list pv, runtime: %s, pv: %s, (%v)", sg.ID, pvName, err)
				continue
			}
			for i := range list.Items {
				if !strings.HasPrefix(list.Items[i].Name, pvName) {
					continue
				}
				logrus.Infof("succeed to got pvName: %s, phase: %v", list.Items[i].Name, list.Items[i].Status.Phase)
				if err := k.pv.Delete(list.Items[i].Name); err != nil {
					logrus.Errorf("failed to delete pv name: %s, (%v)", list.Items[i].Name, err)
				}
			}
		}
	}
	return nil
}
