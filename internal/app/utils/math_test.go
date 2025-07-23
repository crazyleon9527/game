package utils

import (
	"math"
	"testing"
)

func TestHexEquation_Solve(t *testing.T) {
	type args struct {
		y float64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "test",
			args: args{y: 1.36},
			want: 7,
		},
		{
			name: "test2",
			args: args{y: 154.65},
			want: 84,
		},
		{
			name: "test3",
			args: args{y: 2.69},
			want: 17,
		},
		{
			name: "test4",
			args: args{y: 16.95},
			want: 49,
		},
		{
			name: "test4",
			args: args{y: 1000},
			want: 118,
		},
		{
			name: "test4",
			args: args{y: 2237.97},
			want: 135,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHexEquation(
				0.00000000035,
				0.000000001,
				0.0000002,
				0.0044,
				0.02,
				1,
			)
			got, err := h.Solve(tt.args.y)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexEquation.Solve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got2 := int64(math.Round(got))
			if got2 != tt.want {
				t.Errorf("HexEquation.Solve() = %v, want %v", got2, tt.want)
			}
			ret := h.Result(got)
			if math.Round(ret) != math.Round(tt.args.y) {
				t.Errorf("HexEquation.Result() = %v, want %v", ret, tt.args.y)
			}
		})
	}
}
