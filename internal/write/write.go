package write

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// WriteHCL writes the given HCL file to the given file name.
func WriteHCL(f *hcl.File, fileName string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	outContent := hclwrite.Format(f.Bytes)
	if err := os.WriteFile(fileName, outContent, 0644); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to write to file: \"%s\"", fileName),
			Detail:   err.Error(),
		})
	}

	return diags
}
