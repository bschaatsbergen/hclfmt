package parse

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
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

// ParseHCL is a simple wrapper around hclparse.Parser.ParseHCL.
// It reads the content of the given file and parses it into an *hcl.File.
// The function returns the parsed file along with any diagnostics produced.
func (p *Parser) ParseHCL(fileName string) (*hcl.File, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	inContent, err := os.ReadFile(fileName)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Failed to read file",
			Detail:  err.Error(),
		})
		return nil, diags
	}

	f, diags := p.Parser.ParseHCL(inContent, fileName)
	if diags.HasErrors() {
		return nil, diags
	}

	return f, diags
}
