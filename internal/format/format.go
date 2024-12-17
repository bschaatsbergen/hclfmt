package format

import "github.com/hashicorp/hcl/v2/hclwrite"

// FormatHCL takes source code and performs simple whitespace changes to transform
// it to a canonical layout style. It simply wraps hclwrite.Format.
func FormatHCL(b []byte) []byte {
	return hclwrite.Format(b)
}
