package client

import "testing"

func TestEncodeCredential(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "admin", want: "YWRtaW4AAAAA"},
		{raw: "testpass9", want: "dGVzdHBhc3M5"},
		{raw: "1234567890", want: "MTIzNDU2Nzg5MA"},
	}

	for _, tt := range tests {
		if got := encodeCredential(tt.raw); got != tt.want {
			t.Fatalf("encodeCredential(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}
