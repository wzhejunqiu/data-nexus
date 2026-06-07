package sqlutil

import "strings"

const MaxIdentifierLen = 128

// IsSafeQuotedIdentifier reports whether name is safe to embed via fmt.Sprintf("%q", name).
// Dynamic SQL in this project quotes identifiers with %q; rejects empty names and characters
// that break quoting (NUL, double quote).
func IsSafeQuotedIdentifier(name string) bool {
	if name == "" || len(name) > MaxIdentifierLen {
		return false
	}
	if strings.TrimSpace(name) == "" {
		return false
	}
	for _, r := range name {
		if r == 0 || r == '"' {
			return false
		}
	}
	return true
}
