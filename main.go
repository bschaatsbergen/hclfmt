package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	diagWriter hcl.DiagnosticWriter

	fmtSupportedExts = map[string]bool{
		".hcl": true,
	}
)

func main() {
	diags := run()
	if diags.HasErrors() {
		diagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() hcl.Diagnostics {
	var diags hcl.Diagnostics

	cli := cli.NewCLI(cliName, version.Version)
	cli.Args = os.Args[1:]
	cli.HelpFunc = Help()

	flags := flag.NewFlagSet(cliName, flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprint(os.Stdout, cli.HelpFunc(cli.Commands))
	}

	flagStore = NewFlagStore()
	flags.BoolVar(&flagStore.Overwrite, "write", true, "write result to source file instead of stdout")
	flags.BoolVar(&flagStore.Diff, "diff", false, "display diffs of formatting changes")
	flags.BoolVar(&flagStore.Recursive, "recursive", false, "recursively format HCL files in a directory")

	if err := flags.Parse(cli.Args); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to parse flags",
			Detail:   err.Error(),
		})
		return diags
	}

	if cli.IsVersion() {
		fmt.Fprintln(os.Stdout, cli.Version)
		return diags
	}

	if cli.IsHelp() {
		fmt.Fprintln(cli.HelpWriter, cli.HelpFunc(cli.Commands))
		return diags
	}

	if flags.NArg() != 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Expected exactly one file or directory",
		})
		return diags
	}
	target := flags.Arg(0)

	color := term.IsTerminal(int(os.Stderr.Fd()))
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
	}
	diagWriter = hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(width), color)

	if flagStore.Recursive {
		err := filepath.WalkDir(target, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Error accessing path",
					Detail:   fmt.Sprintf("Path: %s, Error: %s", path, walkErr),
				})
				return nil // Continue walking, collect diagnostics
			}

			// Process files only if they are supported
			if !d.IsDir() && isSupportedFile(path) {
				processDiags := processFile(path)
				diags = append(diags, processDiags...)
			}
			return nil // Collect processing diagnostics, continue walking
		})

		// If WalkDir itself failed, append the error to diagnostics
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to recursively walk directory",
				Detail:   err.Error(),
			})
		}

		// Return all collected diagnostics
		return diags
	}

	// By default, process a single given file
	processDiags := processFile(target)
	diags = append(diags, processDiags...)
	return diags
}

func processFile(fileName string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	bytes, err := os.ReadFile(fileName)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to read file: \"%s\"", fileName),
			Detail:   err.Error(),
		})
		return diags
	}

	formattedBytes, formatDiags := format(fileName)
	diags = append(diags, formatDiags...)
	if diags.HasErrors() {
		return diags
	}

	// If the file is already formatted, we simply return
	if string(bytes) == string(formattedBytes) {
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
		return diags
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
		return nil, diags
	}

	f, parseDiags := parser.ParseHCL(fileName)
	diags = append(diags, parseDiags...)
	if diags.HasErrors() {
		return nil, diags
	}

	return hclwrite.Format(f.Bytes), diags
}

func isSupportedFile(path string) bool {
	return fmtSupportedExts[filepath.Ext(path)]
}
