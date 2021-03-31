package marathon

import (
	"fmt"

	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
)

func (m *Marathon) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return fmt.Errorf("SetNodeLabels not implemented in marathon")
}

// // SetNodeLabels set the labels of the dcos node
// func (m *Marathon) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
// 	u, err := url.Parse(setting.SoldierURL) // http://colony-soldier.marathon.l4lb.thisdcos.directory:9028
// 	if err != nil {
// 		return err
// 	}
// 	host := u.Host
// 	// {"force":false,
// 	//  "tag":"any,workspace-dev,workspace-test,workspace-staging,workspace-prod,job,org-terminus",
// 	//  "hosts":["10.168.0.100"]}
// 	var req struct {
// 		Force bool     `json:"force"`
// 		Tag   string   `json:"tag"`
// 		Hosts []string `json:"hosts"`
// 	}
// 	tags := []string{}
// 	// For marathon, `label value` is meaningless because we only use the `label key`
// 	for k := range labels {
// 		tags = append(tags, convertLabels(k))
// 	}
// 	req.Tag = strutil.Join(tags, ",")
// 	req.Hosts = hosts
// 	req.Force = true

// 	logrus.Infof("set dcos nodelabels, hosts: %v, req: %v", req)

// 	var body bytes.Buffer
// 	c, err := httpclient.New().Post(host).Path("/api/nodes/tag").JSONBody(req).Do().Body(&body)
// 	if err != nil {
// 		return err
// 	}
// 	if !c.IsOK() {
// 		return fmt.Errorf("failed to set nodelabel: %v", body.String())
// 	}
// 	return nil
// }

func convertLabels(key string) string {
	switch key {
	case "pack-job":
		return "pack"
	case "bigdata-job":
		return "bigdata"
	case "stateful-service":
		return "service-stateful"
	case "stateless-service":
		return "service-stateless"
	default:
		return key
	}
}
