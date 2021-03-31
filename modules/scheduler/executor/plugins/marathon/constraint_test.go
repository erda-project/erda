package marathon

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstraints(t *testing.T) {
	cons := NewConstraints()

	r1 := cons.NewLikeRule("TATATAT")
	r1.OR(AND("xxx", "yyy"), AND("platform"))

	reg := regexp.MustCompile(r1.generate())

	assert.Zero(t, len(reg.FindStringIndex("xxx")))
	assert.NotZero(t, len(reg.FindStringIndex("xxx,yyy")))
	assert.NotZero(t, len(reg.FindStringIndex("xxx,platform")))

	r2 := cons.NewLikeRule("dice_tags")
	r2.OR(AND("non"), AND("org-terminus", "workspace-dev", "workspace-test"))
	reg2 := regexp.MustCompile(r2.generate())

	assert.NotZero(t,
		len(reg2.FindStringIndex("any,org-terminus,workspace-dev,workspace-test,workspace-staging,workspace-produn")))
}
