package write

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// WriteHCL writes the given HCL file to the given file name or stdout, depending on the overwrite flag.
func WriteHCL(f *hcl.File, fileName string, overwrite bool) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if !overwrite {
		_, err := fmt.Fprintln(os.Stdout, string(f.Bytes))
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Summary: "Failed to write to stdout",
				Detail:  err.Error(),
			})
		}
		return diags
	}

	outContent := hclwrite.Format(f.Bytes)
	if err := os.WriteFile(fileName, outContent, 0644); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Failed to write file",
			Detail:  err.Error(),
		})
	}

	return diags
}
