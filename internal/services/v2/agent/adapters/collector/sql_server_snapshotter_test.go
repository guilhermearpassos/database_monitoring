package collector

import (
	"encoding/base64"
	"errors"
	"testing"
)

func TestPlanHandleValidation(t *testing.T) {
	validBytes := make([]byte, 44)
	validBytes[0] = 0x06
	validBytes[10] = 0xAB
	validB64 := base64.StdEncoding.EncodeToString(validBytes)

	tests := []struct {
		name string
		raw  []byte
		want string
	}{
		{name: "nil", raw: nil, want: ""},
		{name: "empty", raw: []byte{}, want: ""},
		{name: "single_zero_byte_0x0", raw: []byte{0x00}, want: ""},
		{name: "all_zero_44_bytes", raw: make([]byte, 44), want: ""},
		{name: "too_long", raw: append(make([]byte, 64), 0x01), want: ""},
		{name: "valid_handle", raw: validBytes, want: validB64},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := encodeValidPlanHandle(tc.raw); got != tc.want {
				t.Fatalf("encodeValidPlanHandle() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsInvalidPlanHandleErr(t *testing.T) {
	if !isInvalidPlanHandleErr(errors.New("mssql: The handle that was passed to dm_exec_query_plan was invalid.")) {
		t.Fatalf("expected invalid plan handle error to be recognized")
	}
	if isInvalidPlanHandleErr(errors.New("connection reset by peer")) {
		t.Fatalf("did not expect unrelated error to match")
	}
}
