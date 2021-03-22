package pipelineymlv1

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_timeout(t *testing.T) {
	// 1
	_, err := parseTimeout(1)
	fmt.Println(err)
	require.Error(t, err)

	// -1
	forever, err := parseTimeout(-1)
	require.NoError(t, err)
	require.True(t, forever == -1)

	// int/int32/int64/float32/float64 -1
	forever, err = parseTimeout(int(-1))
	require.NoError(t, err)
	require.True(t, forever == -1)
	forever, err = parseTimeout(int32(-1))
	require.NoError(t, err)
	require.True(t, forever == -1)
	forever, err = parseTimeout(int64(-1))
	require.NoError(t, err)
	require.True(t, forever == -1)
	forever, err = parseTimeout(float32(-1))
	require.NoError(t, err)
	require.True(t, forever == -1)
	forever, err = parseTimeout(float64(-1))
	require.NoError(t, err)
	require.True(t, forever == -1)

	// a
	_, err = parseTimeout("a")
	fmt.Println(err)
	require.Error(t, err)

	// 1h
	_, err = parseTimeout("1h")
	require.NoError(t, err)

	// 1m1h
	d, err := parseTimeout("1m1h")
	require.NoError(t, err)
	require.Equal(t, time.Hour+time.Minute, d)
	dd, err := parseTimeout("1h1m")
	require.NoError(t, err)
	require.True(t, dd == d)

	// invalid type
	_, err = parseTimeout(map[string]string{"a": "b"})
	fmt.Println(err)
	require.Error(t, err)

	// -3 invalid
	d, err = parseTimeout(-3)
	require.Error(t, err)

	// -1h1s invalid
	d, err = parseTimeout("-1h1s")
	require.Error(t, err)
	fmt.Println(err)

	// 0 invalid
	d, err = parseTimeout(0)
	require.Error(t, err)
	fmt.Println(err)

	// "0" invalid
	d, err = parseTimeout("0")
	require.Error(t, err)
	fmt.Println(err)
}

func TestPipelineYml_checkTimeout(t *testing.T) {
	var getTask GetTask
	getTask.Timeout = -1

	var putTask PutTask
	putTask.Timeout = "1h1m1s"

	var customTask CustomTask

	y := PipelineYml{
		obj: &Pipeline{
			Stages: []*Stage{
				{
					Name:  "bigdata",
					Tasks: []StepTask{&getTask, &putTask, &customTask},
				},
			},
		},
	}
	err := y.checkTimeout()
	require.NoError(t, err)
	d, err := y.obj.Stages[0].Tasks[2].GetTimeout()
	require.NoError(t, err)
	fmt.Println(d)
	require.True(t, d == -2)

	getTask.Timeout = -2
	putTask.Timeout = "1d"
	customTask.Timeout = "20s"

	y = PipelineYml{
		obj: &Pipeline{
			Stages: []*Stage{
				{
					Name:  "bigdata",
					Tasks: []StepTask{&getTask, &putTask, &customTask},
				},
			},
		},
	}
	err = y.checkTimeout()
	require.Error(t, err)
}
