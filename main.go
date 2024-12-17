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
	"github.com/mitchellh/cli"
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
	f, parseDiags := parser.ParseHCL(fileName)
	diags = append(diags, parseDiags...)
	if diags.HasErrors() {
		parser.DiagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}

	writeDiags := write.WriteHCL(f, fileName)
	diags = append(diags, writeDiags...)
	if diags.HasErrors() {
		parser.DiagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, fileName)

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
