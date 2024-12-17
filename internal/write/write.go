package write

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func WriteHCL(f *hcl.File, fileName string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	outContent := hclwrite.Format(f.Bytes)
	if err := os.WriteFile(fileName, outContent, 0644); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Failed to write file",
			Detail:  err.Error(),
		})
	}

	return diags
}
