package dingding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DingPrint(t *testing.T) {
	var contentArr []rune

	maxContentSize = 10

	for i := 0; i < maxContentSize; i++ {
		contentArr = append(contentArr, 'A')
	}
	content := string(contentArr)

	t.Log("content count", len(content))

	test := []struct {
		name    string
		want    string
		t       string
		isError bool
	}{
		{
			name:    "no_over",
			want:    content,
			t:       content,
			isError: true,
		},
		{
			name:    "over",
			want:    content,
			t:       content + "1",
			isError: true,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			got := DingPrint(tt.t)
			require.Equal(t, got, tt.want)
		})
	}
}
