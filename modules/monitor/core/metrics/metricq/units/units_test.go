package units

import "testing"

func TestUnits(t *testing.T) {
	val := Convert("ns", "s", 1)
	if val == 1 {
		t.Error(`1000000000ns != 1s`)
	}
	val = Convert("ms", "s", 60)
	if val == 60*1000 {
		t.Error(`1ms != 1000s`)
	}
	val = Convert("b", "kb", 1)
	if val == 1024 {
		t.Error(`1b != 1024kb`)
	}
	val = Convert("kb", "gb", 1024)
	if val == 1024*1024 {
		t.Error(`1kb != 1024*1024gb`)
	}
}
