package orgapis

import (
	"testing"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/stretchr/testify/assert"
)

type mockQueryService struct {
}

func (m *mockQueryService) queryComponentStatus(componentType, clusterName string) (statuses []*statusDTO, err error) {
	switch componentType {
	case "cluster":
		statuses = []*statusDTO{
			{Name: "cluster_status", DisplayName: "", Status: 1},
		}
		return
	case "component":
		statuses = []*statusDTO{
			{Name: "kubernetes", DisplayName: "", Status: 1},
			{Name: "dice_component", DisplayName: "", Status: 0},
			{Name: "machine", DisplayName: "", Status: 1},
		}
		return
	}
	return
}

func newMockProvider() *provider {
	return &provider{
		service: &mockQueryService{},
		L:       logrusx.New(),
	}
}

func TestCreateStatusRespStatus(t *testing.T) {
	p := newMockProvider()
	component, _ := p.getComponentStatus("")
	cluster, _ := p.getClusterStatus("")

	res := createStatusResp(cluster, component)

	assert.Equal(t, "cluster_status", res.Name)
	assert.Equal(t, uint8(1), res.Components["machine"].Status)
}
