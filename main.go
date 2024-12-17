package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/bschaatsbergen/hclfmt/internal/parse"
	"github.com/bschaatsbergen/hclfmt/internal/write"
	"github.com/bschaatsbergen/hclfmt/version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/cli"
	"golang.org/x/term"
)

const (
	cliName = "hclfmt"
)

func main() {
	var diags hcl.Diagnostics

	cli := cli.NewCLI(cliName, version.Version)
	cli.Args = os.Args[1:]
	cli.HelpFunc = Help()

	flags := flag.NewFlagSet(cliName, flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprint(os.Stdout, cli.HelpFunc(cli.Commands))
		os.Exit(0)
	}

	flagStore := NewFlagStore()
	flags.BoolVar(&flagStore.Overwrite, "write", true, "write result to source file instead of stdout")
	flags.BoolVar(&flagStore.Diff, "diff", false, "display diffs of formatting changes")

	if err := flags.Parse(cli.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if cli.IsVersion() {
		fmt.Fprintln(os.Stdout, cli.Version)
		os.Exit(0)
	}

	if cli.IsHelp() {
		fmt.Fprintln(cli.HelpWriter, cli.HelpFunc(cli.Commands))
		os.Exit(0)
	}

	if flags.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "You must specify exactly one file to format\n")
		os.Exit(1)
	}
	fileName := flags.Arg(0)

	parser := parse.NewParser()

	color := term.IsTerminal(int(os.Stderr.Fd()))
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
	}
	diagWriter := hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(width), color)

	_, err = os.Stat(fileName)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("No file or directory at \"%s\"", fileName),
			Detail:   err.Error(),
		})
		diagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}

	f, parseDiags := parser.ParseHCL(fileName)
	diags = append(diags, parseDiags...)
	if diags.HasErrors() {
		diagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}

	formattedBytes := hclwrite.Format(f.Bytes)

	if !flagStore.Overwrite {
		_, err := fmt.Fprintln(os.Stdout, string(formattedBytes))
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to write to stdout",
				Detail:   err.Error(),
			})
			diagWriter.WriteDiagnostics(diags)
			os.Exit(1)
		}

		// We're done, so exit successfully
		os.Exit(0)
	}

	if flagStore.Diff {
		diff, err := bytesDiff(formattedBytes, f.Bytes, fileName)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to diff",
				Detail:   err.Error(),
			})
			diagWriter.WriteDiagnostics(diags)
			os.Exit(1)
		}
		if len(diff) > 0 {
			fmt.Fprintln(os.Stdout, string(diff))
		}

		// We're done, so exit successfully
		os.Exit(0)
	}

	writeDiags := write.WriteHCL(f, fileName)
	diags = append(diags, writeDiags...)
	if diags.HasErrors() {
		diagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}

	if flagStore.Overwrite {
		fmt.Fprintln(os.Stdout, fileName)
	}
}

func Help() cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var b bytes.Buffer
		tw := tabwriter.NewWriter(&b, 0, 8, 1, '\t', 0)
		defer tw.Flush()

		fmt.Fprintln(tw, "Usage: hclfmt [-version] [-help] [args]")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "Examples:")
		fmt.Fprintln(tw, "    hclfmt example.hcl")

		return b.String()
	}
}
