package crash

import (
	"rk-api/internal/app/utils"
	"testing"
)

func TestCrashGame_genHash(t *testing.T) {
	type args struct {
		serverSeed string
		blockHash  string
	}
	tests := []struct {
		name string
		g    *CrashGame
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				serverSeed: "a7f71d980b02e79a570c164c1c075e164b20cf7f62024021992bd84623ec6cf6",
				blockHash:  "0000000000000000001b34dc6a1e86083f95500b096231436e9b25cbdd0075c4",
			},
			want: "f50f7df8dfbf504e265de573acc349b4bad657bf339bf01a6f2a0c0cc0a5b67e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.HmacSHA256(tt.args.serverSeed, tt.args.blockHash); got != tt.want {
				t.Errorf("CrashGame.buildHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
