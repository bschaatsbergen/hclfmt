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

	// Supported file extensions for formatting
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

	// CLI initialization and argument parsing.
	cli := cli.NewCLI(cliName, version.Version)
	cli.Args = os.Args[1:]
	cli.HelpFunc = Help()

	flags := flag.NewFlagSet(cliName, flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprint(os.Stdout, cli.HelpFunc(cli.Commands))
	}

	// A flag store simply holds the parsed flags for later use.
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

	// Handle version and help flags.
	if cli.IsVersion() {
		fmt.Fprintln(os.Stdout, cli.Version)
		return diags
	}
	if cli.IsHelp() {
		fmt.Fprintln(cli.HelpWriter, cli.HelpFunc(cli.Commands))
		return diags
	}

	// Ensure exactly one target (file or directory) is specified.
	if flags.NArg() != 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Expected exactly one file or directory",
		})
		return diags
	}
	target := flags.Arg(0)

	// Configure diagnostic writer.
	color := term.IsTerminal(int(os.Stderr.Fd()))
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80 // Default terminal width if size retrieval fails.
	}
	diagWriter = hcl.NewDiagnosticTextWriter(os.Stderr, parser.Files(), uint(width), color)

	// Handle recursive formatting.
	if flagStore.Recursive {
		err := filepath.WalkDir(target, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Error accessing path",
					Detail:   fmt.Sprintf("Path: %s, Error: %s", path, walkErr),
				})
				return nil // Collect diagnostics and continue walking.
			}

			// Only process supported files
			if !d.IsDir() && isSupportedFile(path) {
				processDiags := processFile(path)
				diags = append(diags, processDiags...)
			}
			return nil // Collect diagnostics and continue walking.
		})

		// Something else went wrong during the walk.
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to recursively walk directory",
				Detail:   err.Error(),
			})
		}

		return diags
	}

	// By default, we process a single file.
	processDiags := processFile(target)
	diags = append(diags, processDiags...)
	return diags
}

func processFile(fileName string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// Read file contents.
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to read file: \"%s\"", fileName),
			Detail:   err.Error(),
		})
		return diags
	}

	// Format file contents.
	formattedBytes, formatDiags := format(fileName)
	diags = append(diags, formatDiags...)
	if diags.HasErrors() {
		return diags
	}

	// Exit early if file is in canonical form.
	if string(bytes) == string(formattedBytes) {
		return diags
	}

	// Write to stdout or perform diffs based on flags.
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

	// Write the produced result to the source file.
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
		tw := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)
		defer tw.Flush()

		fmt.Fprintln(tw, "Usage: hclfmt [options] <file or directory>")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "Description:")
		fmt.Fprintln(tw, "  Formats all HCL configuration files to a canonical format. Supported")
		fmt.Fprintln(tw, "  configuration files (.hcl) are updated in place unless otherwise specified.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "  By default, hclfmt scans the current directory for HCL configuration files.")
		fmt.Fprintln(tw, "  If you provide a directory as the target argument, hclfmt will scan that")
		fmt.Fprintln(tw, "  directory recursively when the -recursive flag is set. If you provide a file,")
		fmt.Fprintln(tw, "  hclfmt will process only that file.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "Options:")
		fmt.Fprintln(tw, "  -write=true")
		fmt.Fprintln(tw, "      Write formatted output back to the source file (default: true).")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "  -diff")
		fmt.Fprintln(tw, "      Display diffs of formatting changes without modifying files.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "  -recursive")
		fmt.Fprintln(tw, "      Recursively format HCL files in the specified directory.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "  -help")
		fmt.Fprintln(tw, "      Show this help message.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "  -version")
		fmt.Fprintln(tw, "      Display the version of hclfmt.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "Examples:")
		fmt.Fprintln(tw, "  hclfmt example.hcl")
		fmt.Fprintln(tw, "      Formats the specified file.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "  hclfmt -recursive ./directory")
		fmt.Fprintln(tw, "      Formats all supported HCL files in the specified directory and its subdirectories.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "  hclfmt -diff example.hcl")
		fmt.Fprintln(tw, "      Displays the formatting changes for the specified file without modifying it.")
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "Supported file extensions:")
		fmt.Fprintln(tw, "  .hcl")

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

	// Perform the necessary surgical changes.
	return hclwrite.Format(f.Bytes), diags
}

func isSupportedFile(path string) bool {
	return fmtSupportedExts[filepath.Ext(path)]
}
