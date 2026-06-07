package service

import "testing"

func TestRewriteDescribeSQL(t *testing.T) {
	tests := []struct {
		sql      string
		expected string
		rewrote  bool
	}{
		{"DESC orders", "PRAGMA table_info(orders)", true},
		{"describe `orders`", "PRAGMA table_info(`orders`)", true},
		{"DESC orders;", "PRAGMA table_info(orders)", true},
		{"desc SELECT * from orders limit 10;", "desc SELECT * from orders limit 10;", false},
	}
	for _, tc := range tests {
		got, ok := rewriteDescribeSQL(tc.sql)
		if ok != tc.rewrote {
			t.Fatalf("%q: rewrote=%v want %v", tc.sql, ok, tc.rewrote)
		}
		if got != tc.expected {
			t.Fatalf("%q: got %q want %q", tc.sql, got, tc.expected)
		}
	}
}
