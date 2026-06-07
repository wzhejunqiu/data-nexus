package postgres

import "testing"

func TestRebindPlaceholders(t *testing.T) {
	got := rebindPlaceholders("SELECT * FROM t WHERE a = ? AND b IN (?, ?) LIMIT ? OFFSET ?")
	want := "SELECT * FROM t WHERE a = $1 AND b IN ($2, $3) LIMIT $4 OFFSET $5"
	if got != want {
		t.Fatalf("rebindPlaceholders() = %q, want %q", got, want)
	}
}
