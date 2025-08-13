package database

import "testing"

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "valid utf8",
			arg:  string([]byte{0xeb, 0xdf, 0x74}),
			want: "��t",
		},
		{
			name: "valid utf8",
			arg:  "AG¡",
			want: "AG¡",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeString(tt.arg); got != tt.want {
				t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
			}
		})
	}
}
