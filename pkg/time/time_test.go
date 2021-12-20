package time

import "testing"

func TestAutomaticConversionUnit(t *testing.T) {
	type args struct {
		v float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case-ns", args{v: 0.2}, "ns"},
		{"case-ns", args{v: 1}, "ns"},
		{"case-µs", args{v: 1000}, "µs"},
		{"case-ms", args{v: 1000000}, "ms"},
		{"case-s", args{v: 1000000000}, "s"},
		{"case-m", args{v: 60000000000}, "m"},
		{"case-h", args{v: 14400000000000}, "h"},
		{"case-h", args{v: 14400000000000}, "h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, unit := AutomaticConversionUnit(tt.args.v)
			if unit != tt.want {
				t.Errorf("AutomaticConversionUnit() got1 = %v, want %v", unit, tt.want)
			}
		})
	}
}
