package elasticsearch

import (
	"fmt"
	"github.com/erda-project/erda/apistructs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidate(t *testing.T) {
	esOperator := ElasticsearchOperator{
		k8s:         nil,
		statefulset: nil,
		ns:          nil,
		service:     nil,
		overcommit:  nil,
		secret:      nil,
		configmap:   nil,
		imageSecret: nil,
		client:      nil,
	}

	testcases := []struct {
		name  string
		input *apistructs.ServiceGroup
		want  error
	}{
		{
			name: "valid USE_OPERATOR",
			input: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Labels: map[string]string{},
				},
			},
			want: fmt.Errorf("[BUG] sg need USE_OPERATOR label"),
		},
		{
			name: "USE_OPERATOR is not elasticsearch",
			input: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Labels: map[string]string{
						"USE_OPERATOR": "test",
					},
				},
			},
			want: fmt.Errorf("[BUG] value of label USE_OPERATOR should be 'elasticsearch'"),
		},
		{
			name: "not VERSION",
			input: &apistructs.ServiceGroup{
				Dice: apistructs.Dice{
					Labels: map[string]string{
						"USE_OPERATOR": "elasticsearch",
					},
				},
			},
			want: fmt.Errorf("[BUG] sg need VERSION label"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := esOperator.Validate(tc.input)
			if err != nil && tc.want == nil {
				t.Errorf("expected no error, got %v", err)
			}
			assert.EqualError(t, err, tc.want.Error())
		})
	}
}
