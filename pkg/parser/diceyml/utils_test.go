package diceyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssignWithoutEmpty(t *testing.T) {
	{
		a, b := 1, 2
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 2, a)
	}
	{
		a, b := 1, 0
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 1, a)
	}

	{
		a := []int{1, 1}
		b := []int{2, 2}
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 2, len(a))
		assert.Equal(t, 2, a[0])
		assert.Equal(t, 2, a[1])
	}
	{
		a := []int{1, 1}
		b := []int{}
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 2, len(a))
		assert.Equal(t, 1, a[0])
		assert.Equal(t, 1, a[1])
	}

	{
		a := []int64{1, 1}
		b := []int64{2, 2}
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 2, len(a))
		assert.Equal(t, int64(2), a[0])
		assert.Equal(t, int64(2), a[1])
	}
	{
		a := []int64{1, 1}
		b := []int64{}
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 2, len(a))
		assert.Equal(t, int64(1), a[0])
		assert.Equal(t, int64(1), a[1])
	}

	{
		a := []float64{1, 1}
		b := []float64{2, 2}
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 2, len(a))
		assert.Equal(t, float64(2), a[0])
		assert.Equal(t, float64(2), a[1])
	}
	{
		a := []float64{1, 1}
		b := []float64{}
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 2, len(a))
		assert.Equal(t, float64(1), a[0])
		assert.Equal(t, float64(1), a[1])
	}

	{
		a, b := "1", "2"
		assignWithoutEmpty(&a, b)
		assert.Equal(t, "2", a)
	}
	{
		a, b := []string{"1"}, []string{"2"}
		assignWithoutEmpty(&a, b)
		assert.Equal(t, 1, len(a))
		assert.Equal(t, "2", a[0])
	}
}

//func TestCopyObj(t *testing.T) {
//	src := new(Object)
//	src.Services = map[string]*Service{
//		"aa": {
//			Image: "bb",
//		},
//	}
//	dst := CopyObj(src)
//	assert.True(t, dst.Services["aa"].Image == "bb")
//}
