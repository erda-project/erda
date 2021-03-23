package gittarutil

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

var r *Repo

func init() {
	r = NewRepo("gittar.marathon.l4lb.thisdcos.directory:5566", "terminus-dice-test/pampas-blog")
}

func TestRepo_Branches(t *testing.T) {
	branches, err := r.Branches()
	require.NoError(t, err)
	spew.Dump(branches)
}

func TestRepo_Tags(t *testing.T) {
	tags, err := r.Tags()
	require.NoError(t, err)
	spew.Dump(tags)
}
