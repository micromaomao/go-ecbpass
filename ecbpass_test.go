package main

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func Test1(t *testing.T) {
	tests := []struct {
		password     string
		salt         string
		wantPBKDF    string
		wantHashhint string
	}{
		{"test", "ecbpass", "BiDY;^Du^s", "     4 F        2"},
		{"aaaaa", "google.com", "2djGw;xa7d6", "5         F      3"},
		{"12345", "averylonglonglonglonglonglongdomain.com", "2rP>D?I5QKt", "3 2     EF   BE2"},
		{"averyveryveryveryveryveryverylongpassword", "domain.com", "37Jnqbd`EtO", "3     F4           C"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("#%d: %v %v", i, tt.password, tt.salt), func(t *testing.T) {
			if got := PBKDF2([]byte(tt.password), []byte(tt.salt)); got != tt.wantPBKDF {
				t.Errorf("PBKDF2() = %v, want %v", got, tt.wantPBKDF)
			}
		})
		t.Run(fmt.Sprintf("#%v hashhint", i), func(t *testing.T) {
			if got := strings.TrimRight(Hashhint([]byte(tt.password)), " "); got != tt.wantHashhint {
				t.Errorf("Hashhint() = %v, want %v", got, tt.wantHashhint)
			}
		})
	}
}

func Test_baseStringEnc(t *testing.T) {
	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte{}, "0"},
		{[]byte{0x00}, "0"},
		{[]byte{0x01}, "1"},
		{[]byte{0x02}, "2"},
		{[]byte{0x0a}, ":"},
		{[]byte{0x0b}, ";"},
		// TODO: add more.
	}
	for _, tt := range tests {
		t.Run(hex.EncodeToString(tt.input), func(t *testing.T) {
			if got := baseStringEnc(tt.input); got != tt.want {
				t.Errorf("baseStringEnc() = %v, want %v", got, tt.want)
			}
		})
	}
}
