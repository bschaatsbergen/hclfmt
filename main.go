package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/bschaatsbergen/hclfmt/version"
	"github.com/mitchellh/cli"
)

func main() {
	cli := cli.NewCLI("hclfmt", version.Version)
	cli.Args = os.Args[1:]
	cli.HelpFunc = Help()

	flags := flag.NewFlagSet("hclfmt", flag.ExitOnError)
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

	// Default behavior if no version or help flag is provided
	fmt.Fprintln(os.Stderr, "Invalid usage. Use -help for usage information.")
	os.Exit(1)
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

		return strings.TrimSpace(b.String())
	}
}
