package write

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
)

// WriteHCL writes a byte slice to a file,
// In this context, it is used to write the formatted HCL to the source file.
func WriteHCL(src []byte, fileName string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if err := os.WriteFile(fileName, src, 0644); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to write to file: \"%s\"", fileName),
			Detail:   err.Error(),
		})
	}

	return diags
}
