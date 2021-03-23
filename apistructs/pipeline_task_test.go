package apistructs

import (
	"fmt"
	"testing"
)

func TestPipelineTaskLoop_Duplicate(t *testing.T) {
	var l *PipelineTaskLoop
	fmt.Println(l.Duplicate())
}
