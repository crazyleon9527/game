package mine

import (
	"math"
	"reflect"
	"rk-api/internal/app/entities"
	"testing"
)

func TestMineGame_gererateMinePosition(t *testing.T) {
	type args struct {
		clientSeed string
		serverSeed string
		count      int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "test1-24",
			args: args{
				clientSeed: "0000000000000000001b34dc6a1e86083f95500b096231436e9b25cbdd0075c4",
				serverSeed: "a7f71d980b02e79a570c164c1c075e164b20cf7f62024021992bd84623ec6cf6",
				count:      24,
			},
			want: []int{9, 15, 8, 18, 22, 5, 19, 21, 10, 16, 2, 20, 23, 0, 17, 12, 13, 14, 6, 1, 24, 3, 11, 7},
		},
		{
			name: "test2-3",
			args: args{
				clientSeed: "0000000000000000001b34dc6a1e86083f95500b096231436e9b25cbdd0075c4",
				serverSeed: "a7f71d980b02e79a570c164c1c075e164b20cf7f62024021992bd84623ec6cf6",
				count:      3,
			},
			want: []int{9, 15, 8},
		},
		{
			name: "test3-1",
			args: args{
				clientSeed: "0000000000000000001b34dc6a1e86083f95500b096231436e9b25cbdd0075c4",
				serverSeed: "a7f71d980b02e79a570c164c1c075e164b20cf7f62024021992bd84623ec6cf6",
				count:      1,
			},
			want: []int{9},
		},
		{
			name: "test4-15",
			args: args{
				clientSeed: "0000000000000000001b34dc6a1e86083f95500b096231436e9b25cbdd0075c4",
				serverSeed: "a7f71d980b02e79a570c164c1c075e164b20cf7f62024021992bd84623ec6cf6",
				count:      15,
			},
			want: []int{9, 15, 8, 18, 22, 5, 19, 21, 10, 16, 2, 20, 23, 0, 17},
		},
		{
			name: "test5-5",
			args: args{
				clientSeed: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				serverSeed: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				count:      5,
			},
			want: []int{7, 10, 3, 23, 12},
		},
	}
	m := &MineGame{
		setting: &MineSetting{
			Add1: ":0:0",
			Add2: ":0:1",
			Add3: ":0:2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.gererateMinePosition(tt.args.clientSeed, tt.args.serverSeed, tt.args.count); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MineGame.gererateMinePosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMineGame_calcMultiple(t *testing.T) {
	type args struct {
		mineCount int
		openCount int
		rate      int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"test1-1", args{1, 1, 10}, 1.03}, {"test1-2", args{1, 2, 10}, 1.08}, {"test1-3", args{1, 3, 10}, 1.13}, {"test1-4", args{1, 4, 10}, 1.18}, {"test1-5", args{1, 5, 10}, 1.24}, {"test1-6", args{1, 6, 10}, 1.30}, {"test1-7", args{1, 7, 10}, 1.38}, {"test1-8", args{1, 8, 10}, 1.46}, {"test1-9", args{1, 9, 10}, 1.55}, {"test1-10", args{1, 10, 10}, 1.65}, {"test1-11", args{1, 11, 10}, 1.77}, {"test1-12", args{1, 12, 10}, 1.90}, {"test1-13", args{1, 13, 10}, 2.06}, {"test1-14", args{1, 14, 10}, 2.25}, {"test1-15", args{1, 15, 10}, 2.48}, {"test1-16", args{1, 16, 10}, 2.75}, {"test1-17", args{1, 17, 10}, 3.09}, {"test1-18", args{1, 18, 10}, 3.54}, {"test1-19", args{1, 19, 10}, 4.13}, {"test1-20", args{1, 20, 10}, 4.95}, {"test1-21", args{1, 21, 10}, 6.19}, {"test1-22", args{1, 22, 10}, 8.25}, {"test1-23", args{1, 23, 10}, 12.38}, {"test1-24", args{1, 24, 10}, 24.75},
		{"test2-1", args{2, 1, 10}, 1.08}, {"test2-2", args{2, 2, 10}, 1.17}, {"test2-3", args{2, 3, 10}, 1.29}, {"test2-4", args{2, 4, 10}, 1.41}, {"test2-5", args{2, 5, 10}, 1.56}, {"test2-6", args{2, 6, 10}, 1.74}, {"test2-7", args{2, 7, 10}, 1.94}, {"test2-8", args{2, 8, 10}, 2.18}, {"test2-9", args{2, 9, 10}, 2.48}, {"test2-10", args{2, 10, 10}, 2.83}, {"test2-11", args{2, 11, 10}, 3.26}, {"test2-12", args{2, 12, 10}, 3.81}, {"test2-13", args{2, 13, 10}, 4.50}, {"test2-14", args{2, 14, 10}, 5.40}, {"test2-15", args{2, 15, 10}, 6.60}, {"test2-16", args{2, 16, 10}, 8.25}, {"test2-17", args{2, 17, 10}, 10.61}, {"test2-18", args{2, 18, 10}, 14.14}, {"test2-19", args{2, 19, 10}, 19.80}, {"test2-20", args{2, 20, 10}, 29.70}, {"test2-21", args{2, 21, 10}, 49.50}, {"test2-22", args{2, 22, 10}, 99.00}, {"test2-23", args{2, 23, 10}, 297.00},
		{"test3-1", args{3, 1, 10}, 1.13}, {"test3-2", args{3, 2, 10}, 1.29}, {"test3-3", args{3, 3, 10}, 1.48}, {"test3-4", args{3, 4, 10}, 1.71}, {"test3-5", args{3, 5, 10}, 2.00}, {"test3-6", args{3, 6, 10}, 2.35}, {"test3-7", args{3, 7, 10}, 2.79}, {"test3-8", args{3, 8, 10}, 3.35}, {"test3-9", args{3, 9, 10}, 4.07}, {"test3-10", args{3, 10, 10}, 5.00}, {"test3-11", args{3, 11, 10}, 6.26}, {"test3-12", args{3, 12, 10}, 7.96}, {"test3-13", args{3, 13, 10}, 10.35}, {"test3-14", args{3, 14, 10}, 13.80}, {"test3-15", args{3, 15, 10}, 18.98}, {"test3-16", args{3, 16, 10}, 27.11}, {"test3-17", args{3, 17, 10}, 40.66}, {"test3-18", args{3, 18, 10}, 65.06}, {"test3-19", args{3, 19, 10}, 113.85}, {"test3-20", args{3, 20, 10}, 227.70}, {"test3-21", args{3, 21, 10}, 569.25}, {"test3-22", args{3, 22, 10}, 2277.00},
		{"test4-1", args{4, 1, 10}, 1.18}, {"test4-2", args{4, 2, 10}, 1.41}, {"test4-3", args{4, 3, 10}, 1.71}, {"test4-4", args{4, 4, 10}, 2.09}, {"test4-5", args{4, 5, 10}, 2.58}, {"test4-6", args{4, 6, 10}, 3.23}, {"test4-7", args{4, 7, 10}, 4.09}, {"test4-8", args{4, 8, 10}, 5.26}, {"test4-9", args{4, 9, 10}, 6.88}, {"test4-10", args{4, 10, 10}, 9.17}, {"test4-11", args{4, 11, 10}, 12.51}, {"test4-12", args{4, 12, 10}, 17.52}, {"test4-13", args{4, 13, 10}, 25.30}, {"test4-14", args{4, 14, 10}, 37.95}, {"test4-15", args{4, 15, 10}, 59.64}, {"test4-16", args{4, 16, 10}, 99.39}, {"test4-17", args{4, 17, 10}, 178.91}, {"test4-18", args{4, 18, 10}, 357.81}, {"test4-19", args{4, 19, 10}, 834.90}, {"test4-20", args{4, 20, 10}, 2504.70}, {"test4-21", args{4, 21, 10}, 12523.50},
		{"test5-1", args{5, 1, 10}, 1.24}, {"test5-2", args{5, 2, 10}, 1.56}, {"test5-3", args{5, 3, 10}, 2.00}, {"test5-4", args{5, 4, 10}, 2.58}, {"test5-5", args{5, 5, 10}, 3.39}, {"test5-6", args{5, 6, 10}, 4.52}, {"test5-7", args{5, 7, 10}, 6.14}, {"test5-8", args{5, 8, 10}, 8.50}, {"test5-9", args{5, 9, 10}, 12.04}, {"test5-10", args{5, 10, 10}, 17.52}, {"test5-11", args{5, 11, 10}, 26.27}, {"test5-12", args{5, 12, 10}, 40.87}, {"test5-13", args{5, 13, 10}, 66.41}, {"test5-14", args{5, 14, 10}, 113.85}, {"test5-15", args{5, 15, 10}, 208.73}, {"test5-16", args{5, 16, 10}, 417.45}, {"test5-17", args{5, 17, 10}, 939.26}, {"test5-18", args{5, 18, 10}, 2504.70}, {"test5-19", args{5, 19, 10}, 8766.45}, {"test5-20", args{5, 20, 10}, 52598.70},
		{"test6-1", args{6, 1, 10}, 1.30}, {"test6-2", args{6, 2, 10}, 1.74}, {"test6-3", args{6, 3, 10}, 2.35}, {"test6-4", args{6, 4, 10}, 3.23}, {"test6-5", args{6, 5, 10}, 4.52}, {"test6-6", args{6, 6, 10}, 6.46}, {"test6-7", args{6, 7, 10}, 9.44}, {"test6-8", args{6, 8, 10}, 14.17}, {"test6-9", args{6, 9, 10}, 21.89}, {"test6-10", args{6, 10, 10}, 35.03}, {"test6-11", args{6, 11, 10}, 58.38}, {"test6-12", args{6, 12, 10}, 102.17}, {"test6-13", args{6, 13, 10}, 189.75}, {"test6-14", args{6, 14, 10}, 379.50}, {"test6-15", args{6, 15, 10}, 834.90}, {"test6-16", args{6, 16, 10}, 2087.25}, {"test6-17", args{6, 17, 10}, 6261.75}, {"test6-18", args{6, 18, 10}, 25047.00}, {"test6-19", args{6, 19, 10}, 175329.00},
		{"test7-1", args{7, 1, 10}, 1.38}, {"test7-2", args{7, 2, 10}, 1.94}, {"test7-3", args{7, 3, 10}, 2.79}, {"test7-4", args{7, 4, 10}, 4.09}, {"test7-5", args{7, 5, 10}, 6.14}, {"test7-6", args{7, 6, 10}, 9.44}, {"test7-7", args{7, 7, 10}, 14.95}, {"test7-8", args{7, 8, 10}, 24.47}, {"test7-9", args{7, 9, 10}, 41.60}, {"test7-10", args{7, 10, 10}, 73.95}, {"test7-11", args{7, 11, 10}, 138.66}, {"test7-12", args{7, 12, 10}, 277.33}, {"test7-13", args{7, 13, 10}, 600.88}, {"test7-14", args{7, 14, 10}, 1442.10}, {"test7-15", args{7, 15, 10}, 3965.78}, {"test7-16", args{7, 16, 10}, 13219.25}, {"test7-17", args{7, 17, 10}, 59486.63}, {"test7-18", args{7, 18, 10}, 475893.00},
		{"test8-1", args{8, 1, 10}, 1.46}, {"test8-2", args{8, 2, 10}, 2.18}, {"test8-3", args{8, 3, 10}, 3.35}, {"test8-4", args{8, 4, 10}, 5.26}, {"test8-5", args{8, 5, 10}, 8.50}, {"test8-6", args{8, 6, 10}, 14.17}, {"test8-7", args{8, 7, 10}, 24.47}, {"test8-8", args{8, 8, 10}, 44.05}, {"test8-9", args{8, 9, 10}, 83.20}, {"test8-10", args{8, 10, 10}, 166.40}, {"test8-11", args{8, 11, 10}, 356.56}, {"test8-12", args{8, 12, 10}, 831.98}, {"test8-13", args{8, 13, 10}, 2163.15}, {"test8-14", args{8, 14, 10}, 6489.45}, {"test8-15", args{8, 15, 10}, 23794.65}, {"test8-16", args{8, 16, 10}, 118973.25}, {"test8-17", args{8, 17, 10}, 1070759.25},
		{"test9-1", args{9, 1, 10}, 1.55}, {"test9-2", args{9, 2, 10}, 2.48}, {"test9-3", args{9, 3, 10}, 4.07}, {"test9-4", args{9, 4, 10}, 6.88}, {"test9-5", args{9, 5, 10}, 12.04}, {"test9-6", args{9, 6, 10}, 21.89}, {"test9-7", args{9, 7, 10}, 41.60}, {"test9-8", args{9, 8, 10}, 83.20}, {"test9-9", args{9, 9, 10}, 176.80}, {"test9-10", args{9, 10, 10}, 404.10}, {"test9-11", args{9, 11, 10}, 1010.26}, {"test9-12", args{9, 12, 10}, 2828.73}, {"test9-13", args{9, 13, 10}, 9193.39}, {"test9-14", args{9, 14, 10}, 36773.55}, {"test9-15", args{9, 15, 10}, 202254.53}, {"test9-16", args{9, 16, 10}, 2022545.25},
		{"test10-1", args{10, 1, 10}, 1.65}, {"test10-2", args{10, 2, 10}, 2.83}, {"test10-3", args{10, 3, 10}, 5.00}, {"test10-4", args{10, 4, 10}, 9.17}, {"test10-5", args{10, 5, 10}, 17.52}, {"test10-6", args{10, 6, 10}, 35.03}, {"test10-7", args{10, 7, 10}, 73.95}, {"test10-8", args{10, 8, 10}, 166.40}, {"test10-9", args{10, 9, 10}, 404.10}, {"test10-10", args{10, 10, 10}, 1077.61}, {"test10-11", args{10, 11, 10}, 3232.84}, {"test10-12", args{10, 12, 10}, 11314.94}, {"test10-13", args{10, 13, 10}, 49031.40}, {"test10-14", args{10, 14, 10}, 294188.40}, {"test10-15", args{10, 15, 10}, 3236072.40},
		{"test11-1", args{11, 1, 10}, 1.77}, {"test11-2", args{11, 2, 10}, 3.26}, {"test11-3", args{11, 3, 10}, 6.26}, {"test11-4", args{11, 4, 10}, 12.51}, {"test11-5", args{11, 5, 10}, 26.27}, {"test11-6", args{11, 6, 10}, 58.38}, {"test11-7", args{11, 7, 10}, 138.66}, {"test11-8", args{11, 8, 10}, 356.56}, {"test11-9", args{11, 9, 10}, 1010.26}, {"test11-10", args{11, 10, 10}, 3232.84}, {"test11-11", args{11, 11, 10}, 12123.15}, {"test11-12", args{11, 12, 10}, 56574.69}, {"test11-13", args{11, 13, 10}, 367735.50}, {"test11-14", args{11, 14, 10}, 4412826.00},
		{"test12-1", args{12, 1, 10}, 1.90}, {"test12-2", args{12, 2, 10}, 3.81}, {"test12-3", args{12, 3, 10}, 7.96}, {"test12-4", args{12, 4, 10}, 17.52}, {"test12-5", args{12, 5, 10}, 40.87}, {"test12-6", args{12, 6, 10}, 102.17}, {"test12-7", args{12, 7, 10}, 277.33}, {"test12-8", args{12, 8, 10}, 831.98}, {"test12-9", args{12, 9, 10}, 2828.73}, {"test12-10", args{12, 10, 10}, 11314.94}, {"test12-11", args{12, 11, 10}, 56574.69}, {"test12-12", args{12, 12, 10}, 396022.85}, {"test12-13", args{12, 13, 10}, 5148297.00},
		{"test13-1", args{13, 1, 10}, 2.06}, {"test13-2", args{13, 2, 10}, 4.50}, {"test13-3", args{13, 3, 10}, 10.35}, {"test13-4", args{13, 4, 10}, 25.30}, {"test13-5", args{13, 5, 10}, 66.41}, {"test13-6", args{13, 6, 10}, 189.75}, {"test13-7", args{13, 7, 10}, 600.88}, {"test13-8", args{13, 8, 10}, 2163.15}, {"test13-9", args{13, 9, 10}, 9193.39}, {"test13-10", args{13, 10, 10}, 49031.40}, {"test13-11", args{13, 11, 10}, 367735.50}, {"test13-12", args{13, 12, 10}, 5148297.00},
		{"test14-1", args{14, 1, 10}, 2.25}, {"test14-2", args{14, 2, 10}, 5.40}, {"test14-3", args{14, 3, 10}, 13.80}, {"test14-4", args{14, 4, 10}, 37.95}, {"test14-5", args{14, 5, 10}, 113.85}, {"test14-6", args{14, 6, 10}, 379.50}, {"test14-7", args{14, 7, 10}, 1442.10}, {"test14-8", args{14, 8, 10}, 6489.45}, {"test14-9", args{14, 9, 10}, 36773.55}, {"test14-10", args{14, 10, 10}, 294188.40}, {"test14-11", args{14, 11, 10}, 4412826.00},
		{"test15-1", args{15, 1, 10}, 2.48}, {"test15-2", args{15, 2, 10}, 6.60}, {"test15-3", args{15, 3, 10}, 18.98}, {"test15-4", args{15, 4, 10}, 59.64}, {"test15-5", args{15, 5, 10}, 208.73}, {"test15-6", args{15, 6, 10}, 834.90}, {"test15-7", args{15, 7, 10}, 3965.78}, {"test15-8", args{15, 8, 10}, 23794.65}, {"test15-9", args{15, 9, 10}, 202254.53}, {"test15-10", args{15, 10, 10}, 3236072.40},
		{"test16-1", args{16, 1, 10}, 2.75}, {"test16-2", args{16, 2, 10}, 8.25}, {"test16-3", args{16, 3, 10}, 27.11}, {"test16-4", args{16, 4, 10}, 99.39}, {"test16-5", args{16, 5, 10}, 417.45}, {"test16-6", args{16, 6, 10}, 2087.25}, {"test16-7", args{16, 7, 10}, 13219.25}, {"test16-8", args{16, 8, 10}, 118973.25}, {"test16-9", args{16, 9, 10}, 2022545.25},
		{"test17-1", args{17, 1, 10}, 3.09}, {"test17-2", args{17, 2, 10}, 10.61}, {"test17-3", args{17, 3, 10}, 40.66}, {"test17-4", args{17, 4, 10}, 178.91}, {"test17-5", args{17, 5, 10}, 939.26}, {"test17-6", args{17, 6, 10}, 6261.75}, {"test17-7", args{17, 7, 10}, 59486.63}, {"test17-8", args{17, 8, 10}, 1070759.25},
		{"test18-1", args{18, 1, 10}, 3.54}, {"test18-2", args{18, 2, 10}, 14.14}, {"test18-3", args{18, 3, 10}, 65.06}, {"test18-4", args{18, 4, 10}, 357.81}, {"test18-5", args{18, 5, 10}, 2504.70}, {"test18-6", args{18, 6, 10}, 25047.00}, {"test18-7", args{18, 7, 10}, 475893.00},
		{"test19-1", args{19, 1, 10}, 4.13}, {"test19-2", args{19, 2, 10}, 19.80}, {"test19-3", args{19, 3, 10}, 113.85}, {"test19-4", args{19, 4, 10}, 834.90}, {"test19-5", args{19, 5, 10}, 8766.45}, {"test19-6", args{19, 6, 10}, 175329.00},
		{"test20-1", args{20, 1, 10}, 4.95}, {"test20-2", args{20, 2, 10}, 29.70}, {"test20-3", args{20, 3, 10}, 227.70}, {"test20-4", args{20, 4, 10}, 2504.70}, {"test20-5", args{20, 5, 10}, 52598.70},
		{"test21-1", args{21, 1, 10}, 6.19}, {"test21-2", args{21, 2, 10}, 49.50}, {"test21-3", args{21, 3, 10}, 569.25}, {"test21-4", args{21, 4, 10}, 12523.50},
		{"test22-1", args{22, 1, 10}, 8.25}, {"test22-2", args{22, 2, 10}, 99.00}, {"test22-3", args{22, 3, 10}, 2277.00},
		{"test23-1", args{23, 1, 10}, 12.38}, {"test23-2", args{23, 2, 10}, 297.00},
		{"test24-1", args{24, 1, 10}, 24.75},
	}
	m := &MineGame{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.calcMultiple(tt.args.mineCount, tt.args.openCount, tt.args.rate)
			got1 := math.Round(got*100) / 100
			if got1 != tt.want {
				t.Errorf("MineGame.calcMultiple() %v, got1 %v, want %v test", got, got1, tt.want)
			}
		})
	}
}

func TestMineGame_checkOrderStatus(t *testing.T) {
	m := &MineGame{}
	order := &entities.MineGameOrder{Status: "preparing", Settled: 0}
	err := m.checkOrderStatus(order, "preparing")
	if err != nil {
		t.Errorf("checkOrderStatus() error = %v, want nil", err)
	}
	order2 := &entities.MineGameOrder{Status: "gameover", Settled: 0}
	if err := m.checkOrderStatus(order2, "preparing"); err == nil {
		t.Errorf("checkOrderStatus() want error for status mismatch")
	}
	order3 := &entities.MineGameOrder{Status: "preparing", Settled: 1}
	if err := m.checkOrderStatus(order3, "preparing"); err == nil {
		t.Errorf("checkOrderStatus() want error for settled")
	}
}

func TestMineGame_checkOpenPositionReq(t *testing.T) {
	m := &MineGame{}
	err := m.checkOpenPositionReq(&entities.MineGameOpenPositionReq{OpenPosition: 10})
	if err != nil {
		t.Errorf("checkOpenPositionReq() error = %v, want nil", err)
	}
	err = m.checkOpenPositionReq(&entities.MineGameOpenPositionReq{OpenPosition: 25})
	if err == nil {
		t.Errorf("checkOpenPositionReq() want error for out of range")
	}
}

func TestMineGame_checkPosition(t *testing.T) {
	m := &MineGame{}
	open := []*entities.MineGamePosition{
		{Position: 1}, {Position: 2},
	}
	err := m.checkPosition(open, 3)
	if err != nil {
		t.Errorf("checkPosition() error = %v, want nil", err)
	}
	err = m.checkPosition(open, 2)
	if err == nil {
		t.Errorf("checkPosition() want error for duplicate")
	}
}

func TestMineGame_checkIsOpenMine(t *testing.T) {
	m := &MineGame{}
	mine := []int{2, 5, 8}
	if !m.checkIsOpenMine(mine, 5) {
		t.Errorf("checkIsOpenMine() want true")
	}
	if m.checkIsOpenMine(mine, 7) {
		t.Errorf("checkIsOpenMine() want false")
	}
}

func TestMineGame_checkChangeSeedReq(t *testing.T) {
	m := &MineGame{}
	err := m.checkChangeSeedReq(&entities.MineGameChangeSeedReq{ClientSeed: "abc"})
	if err != nil {
		t.Errorf("checkChangeSeedReq() error = %v, want nil", err)
	}
	err = m.checkChangeSeedReq(&entities.MineGameChangeSeedReq{ClientSeed: ""})
	if err == nil {
		t.Errorf("checkChangeSeedReq() want error for empty seed")
	}
}
