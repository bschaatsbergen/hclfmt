package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

var (
	flagStore *FlagStore

	parser     = parse.NewParser()
	diagWriter = hcl.NewDiagnosticTextWriter(os.Stderr, nil, 80, true)

	fmtSupportedExts = []string{
		".hcl",
	}
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

	flagStore = NewFlagStore()
	flags.BoolVar(&flagStore.Overwrite, "write", true, "write result to source file instead of stdout")
	flags.BoolVar(&flagStore.Diff, "diff", false, "display diffs of formatting changes")
	flags.BoolVar(&flagStore.Recursive, "recursive", false, "recursively format HCL files in a directory")

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
		fmt.Fprintf(os.Stderr, "You must specify exactly one file or directory to format\n")
		os.Exit(1)
	}
	target := flags.Arg(0)

	color := term.IsTerminal(int(os.Stderr.Fd()))
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
	}
	diagWriter = hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(width), color)

	if flagStore.Recursive {
		err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// If it's a symbolic link, skip it to avoid infinite recursion
			if info.Mode()&os.ModeSymlink != 0 {
				return nil
			}

			if !info.IsDir() && isSupportedFile(info.Name()) {
				processDiags := processFile(path)
				diags = append(diags, processDiags...)
				if diags.HasErrors() {
					diagWriter.WriteDiagnostics(diags)
					os.Exit(1)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing directory: %v\n", err)
			os.Exit(1)
		}
		// No errors, exit
		os.Exit(0)
	}

	// By default, process a single given file
	processDiags := processFile(target)
	diags = append(diags, processDiags...)
	if diags.HasErrors() {
		diagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}
}

func processFile(fileName string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	formattedBytes, formatDiags := format(fileName)
	diags = append(diags, formatDiags...)
	if diags.HasErrors() {
		return diags
	}

	if !flagStore.Overwrite {
		_, err := fmt.Fprintln(os.Stdout, string(formattedBytes))
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to write to stdout",
				Detail:   err.Error(),
			})
			return diags
		}
		return diags
	}

	if flagStore.Diff {
		bytes, err := os.ReadFile(fileName)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to read file: \"%s\"", fileName),
				Detail:   err.Error(),
			})
			return diags
		}

		diff, err := bytesDiff(formattedBytes, bytes, fileName)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to diff",
				Detail:   err.Error(),
			})
			return diags
		}
		if len(diff) > 0 {
			fmt.Fprintln(os.Stdout, string(diff))
		}
		return diags
	}

	writeDiags := write.WriteHCL(formattedBytes, fileName)
	diags = append(diags, writeDiags...)
	if diags.HasErrors() {
		diagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}

	if flagStore.Overwrite {
		fmt.Fprintln(os.Stdout, fileName)
	}

	return diags
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
		fmt.Fprintln(tw, "    hclfmt -recursive ./directory")

		return b.String()
	}
}

func format(fileName string) ([]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	_, err := os.Stat(fileName)
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

	return hclwrite.Format(f.Bytes), diags
}

func isSupportedFile(path string) bool {
	for _, ext := range fmtSupportedExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
