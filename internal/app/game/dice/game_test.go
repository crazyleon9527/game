package dice

import (
	"testing"
)

func TestDiceGame_generateDiceResult(t *testing.T) {
	type args struct {
		clientSeed string
		serverSeed string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test1-24",
			args: args{
				clientSeed: "0000000000000000001b34dc6a1e86083f95500b096231436e9b25cbdd0075c4",
				serverSeed: "a7f71d980b02e79a570c164c1c075e164b20cf7f62024021992bd84623ec6cf6",
			},
			want: 36.73,
		},
	}
	m := &DiceGame{
		setting: &DiceSetting{
			Add1: ":0:0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.generateDiceResult(tt.args.clientSeed, tt.args.serverSeed); got != tt.want {
				t.Errorf("DiceGame.generateDiceResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiceGame_calcMultiple(t *testing.T) {
	type args struct {
		target  float64
		result  float64
		isAbove int
		rate    int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"test1", args{24.69, 24.69, 0, 10}, 4.0097},
		{"test2", args{24.69, 23.69, 0, 10}, 0},
		{"test3", args{24.69, 25.69, 0, 10}, 4.0097},
		{"test4", args{24.69, 24.69, 1, 10}, 1.3146},
		{"test5", args{24.69, 25.69, 1, 10}, 0},
		{"test6", args{24.69, 23.69, 1, 10}, 1.3146},
		{"test7", args{50, 50, 0, 10}, 1.9800},
		{"test8", args{50, 50, 1, 10}, 1.9800},
		{"test9", args{90, 90, 0, 10}, 1.1000},
		{"test10", args{90, 90, 1, 10}, 9.9000},
		{"test11", args{10, 10, 0, 10}, 9.9000},
		{"test12", args{10, 10, 1, 10}, 1.1000},
		{"test13", args{2, 2, 0, 10}, 49.5000},
		{"test14", args{2, 2, 1, 10}, 1.0102},
		{"test15", args{98, 98, 0, 10}, 1.0102},
		{"test16", args{98, 98, 1, 10}, 49.5000},
		{"test17", args{0.01, 0.01, 0, 10}, 9900.0000},
		{"test18", args{99.99, 99.99, 1, 10}, 9900.0000},
	}
	m := &DiceGame{
		setting: &DiceSetting{
			Add1: ":0:0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.calcMultiple(tt.args.target, tt.args.result, tt.args.isAbove, tt.args.rate); got != tt.want {
				t.Errorf("DiceGame.calcMultiple() = %v, want %v", got, tt.want)
			}
		})
	}
}
