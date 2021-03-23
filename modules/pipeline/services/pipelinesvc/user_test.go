package pipelinesvc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpclient"
)

func TestPipelineSvc_tryGetUser(t *testing.T) {
	s := &PipelineSvc{bdl: bundle.New(bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))))}
	invalidUserID := "invalid user id"
	user := s.tryGetUser(invalidUserID)
	assert.Equal(t, invalidUserID, user.ID)
	assert.Empty(t, user.Name)
}
