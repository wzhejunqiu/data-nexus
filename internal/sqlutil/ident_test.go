package sqlutil_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

func TestIsSafeQuotedIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"dest", true},
		{"发送分", true},
		{"bad-name", true},
		{"col_1", true},
		{"", false},
		{"   ", false},
		{`has"quote`, false},
		{"a\x00b", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sqlutil.IsSafeQuotedIdentifier(tc.name)
			if got != tc.valid {
				t.Fatalf("IsSafeQuotedIdentifier(%q) = %v, want %v", tc.name, got, tc.valid)
			}
		})
	}
}
