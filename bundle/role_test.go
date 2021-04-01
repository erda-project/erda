package bundle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundle_CheckIfRoleIsManager(t *testing.T) {
	bdl := New()
	isManager := bdl.CheckIfRoleIsManager(RoleSysManager)
	assert.True(t, isManager)
}
