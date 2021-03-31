package k8s

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"testing"
)

func Test_matchNodeLabels(t *testing.T) {
	r1 := matchNodeLabels([]string{"a", "b", "c"},
		[][2][]string{
			{{"b", "c"}, {"a"}}, // false
			{{"a", "b"}, {}},    // true
		})
	if !r1 {
		t.Errorf("matchNodeLabels r1 failed")
	}

	r2 := matchNodeLabels([]string{"a", "b", "c"},
		[][2][]string{
			{{"b", "c"}, {"a"}}, // false
			{{}, {"c"}},         // false
		})
	if r2 {
		t.Errorf("matchNodeLabels r2 failed")
	}

	r3 := matchNodeLabels([]string{"a", "b", "c"},
		[][2][]string{
			{{"b"}, {"a"}}, // false
			{{"a"}, {"c"}}, // false
			{{}, {}},       // true
		})
	if !r3 {
		t.Errorf("matchNodeLabels r3 failed")
	}
}

func Test_extractLabels(t *testing.T) {
	terms := []v1.NodeSelectorTerm{
		{
			MatchExpressions: []v1.NodeSelectorRequirement{
				{Key: "a", Operator: "Exists"},
				{Key: "b", Operator: "Exists"},
				{Key: "c", Operator: "DoesNotExist"},
			},
		},
		{
			MatchExpressions: []v1.NodeSelectorRequirement{
				{Key: "d", Operator: "Exists"},
				{Key: "e", Operator: "DoesNotExist"},
			},
		},
	}

	r := extractLabels(terms)
	assert.True(t, len(r) == 2)
	assert.True(t, len(r[0][0]) == 2)
	assert.True(t, len(r[0][1]) == 1)
	assert.True(t, len(r[1][0]) == 1)
	assert.True(t, len(r[1][1]) == 1)

}
