package parse

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Parser struct {
	*hclparse.Parser
	DiagWriter hcl.DiagnosticWriter
}

func NewParser() *Parser {
	p := hclparse.NewParser()
	return &Parser{
		Parser:     p,
		DiagWriter: hcl.NewDiagnosticTextWriter(os.Stderr, p.Files(), 80, true),
	}
}

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
