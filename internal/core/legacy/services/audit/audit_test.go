package audit

// Import the necessary packages
import (
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/internal/core/legacy/model"
)

func TestGetAllProjectIdInOrg(t *testing.T) {
	// Create a new instance of Audit
	audit := &Audit{}

	mockProjectMap := make(map[uint64][]model.Project)

	mockProjectMap[1] = []model.Project{
		{BaseModel: model.BaseModel{ID: 1}},
		{BaseModel: model.BaseModel{ID: 2}},
		{BaseModel: model.BaseModel{ID: 3}},
		{BaseModel: model.BaseModel{ID: 4}},
	}

	mockProjectMap[2] = []model.Project{
		{BaseModel: model.BaseModel{ID: 5}},
		{BaseModel: model.BaseModel{ID: 6}},
		{BaseModel: model.BaseModel{ID: 7}},
	}

	// Replace the ListProjectByOrgID method with a mock implementation
	monkey.PatchInstanceMethod(reflect.TypeOf(audit.db), "ListProjectByOrgID", func(_ *dao.DBClient, orgId uint64) ([]model.Project, error) {
		return mockProjectMap[orgId], nil
	})
	defer monkey.UnpatchAll()

	// Call the GetAllProjectIdInOrg method
	projectIds, err := audit.GetAllProjectIdInOrg(1)

	// Perform assertions on the returned values
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedProjectIds := []uint64{1, 2, 3, 4}
	if !reflect.DeepEqual(projectIds, expectedProjectIds) {
		t.Errorf("Expected project IDs %v, but got %v", expectedProjectIds, projectIds)
	}
}
