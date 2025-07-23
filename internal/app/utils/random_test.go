package utils

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"testing"
)

func TestGenerateSecureHex(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSecureHex()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecureHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GenerateSecureHex() = %v \n", got)

			hash := fmt.Sprintf("%x", sha256.Sum256([]byte("round.Hash")))
			t.Logf("sha256 %s", hash)

			k, _ := strconv.ParseInt(hash[:8], 16, 64)
			t.Logf("sha256 k %s %d", hash[:8], k)

			k, _ = strconv.ParseInt("ffffffff", 16, 64)
			t.Logf("sha256 k %s %d", "ffffffff", k)
		})
	}
}
