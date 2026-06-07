package version

import (
	_ "embed"
	"strings"
)

//go:embed product.txt
var productVersion []byte

// Version returns the product version from wails.json (via generated product.txt).
func Version() string {
	return strings.TrimSpace(string(productVersion))
}
