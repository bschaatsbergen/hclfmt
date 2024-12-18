package parse

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Parser knows how to parse HCL files.
type Parser struct {
	*hclparse.Parser
}

// NewParser creates a new instance of Parser, wrapping around hclparse.Parser.
func NewParser() *Parser {
	return &Parser{
		Parser: hclparse.NewParser(),
	}
}

// ParseConfig is a simple wrapper around hclparse.Parser.ParseConfig.
// It reads the content of the given file and parses it into an *hclwrite.File.
// The function returns the parsed file along with any diagnostics produced.
func (p *Parser) ParseConfig(fileName string) (*hclwrite.File, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	inContent, err := os.ReadFile(fileName)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to read file: \"%s\"", fileName),
			Detail:   err.Error(),
		})
		return nil, diags
	}

	p.Parser.AddFile(fileName, &hcl.File{
		Body:  hcl.EmptyBody(),
		Bytes: inContent,
	})

	f, hclDiags := hclwrite.ParseConfig(inContent, fileName, hcl.InitialPos)
	if hclDiags.HasErrors() {
		diags = append(diags, hclDiags...)
		return nil, diags
	}

	return f, diags
}
