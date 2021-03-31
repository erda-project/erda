package clusterinfo

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
)

// ParseJobHostBindTemplate 对 hostPath 进行模版解析，替换成 cluster info 值
func ParseJobHostBindTemplate(hostPath string, clusterInfo apistructs.ClusterInfoData) (string, error) {
	var b bytes.Buffer

	if hostPath == "" {
		return "", errors.New("hostPath is empty")
	}

	t, err := template.New("jobBind").
		Option("missingkey=error").
		Parse(hostPath)
	if err != nil {
		return "", errors.Errorf("failed to parse bind, hostPath: %s, (%v)", hostPath, err)
	}

	err = t.Execute(&b, &clusterInfo)
	if err != nil {
		return "", errors.Errorf("failed to execute bind, hostPath: %s, (%v)", hostPath, err)
	}

	return b.String(), nil
}
