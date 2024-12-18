# hclfmt

An HCL (HashiCorp Configuration Language) formatter.

### You probably won't need this
Sometimes, people ask why there isn’t a generic formatter for HCL (HashiCorp Configuration Language). The short answer is that HCL was designed as a framework for building languages, not as a standalone language, so it’s up to each application to define how formatting should work. Tools like Terraform and Packer include their own formatters, which extend basic HCL conventions with domain-specific rules. This ensures formatting aligns not only with general HCL syntax but also with the specific idiomatic patterns and best practices of the application.

For example, an application might reorder attributes, adjust indentation, or enforce conventions for expressions and relationships between blocks. These details go beyond generic formatting, reflecting the unique ways each tool uses HCL to model its domain.

Applications that provide their own formatters typically start with generic HCL formatting logic and layer on additional rules to handle their specific use cases. This lets them produce configurations that feel natural within their ecosystem, rather than forcing users into a one-size-fits-all approach.

In our own applications, we use the hclwrite package to parse HCL into a hybrid syntax tree. This allows for precise, targeted edits—whether reordering blocks or rewriting expressions into an idiomatic form. The result is then serialized back into HCL, ensuring the output is clean and consistent.

Bear in mind that any normalization process must be **idempotent**, meaning running the formatter multiple times on the same input should always produce the same result. If it doesn’t, treat it as a bug that needs to be addressed.

### "Isn't it all just HCL?"
[@apparentlymart](https://github.com/apparentlymart) explains this far better than I ever could:
> Unlike some other formats like JSON and YAML, a HCL file is more like a program to be executed than a data structure to be parsed, and so there's considerably more application-level interpretation to be done than you might be accustomed to with other grammars.

> HCL is designed as a toolkit for building languages rather than as a language in its own right, but it's true that a bunch of the existing HCL-based languages aren't doing that much above what HCL itself offers, aside from defining their expected block types and attributes.

> The languages that allow for e.g. creating relationships between declared objects via expressions, or writing "libraries" like Terraform's modules, will tend to bend HCL in more complicated ways than where HCL is being used mainly just as a serialization of a flat data structure. To be specific, I would expect the Terraform Language, the Packer Language and the Waypoint Language to all eventually benefit from application-specific extensions with their own formatters, but something like Vault's policy language or Consul's agent configuration files would probably suffice with a generic HCL extension and generic formatter.

Above all, consider this repository a learning resource and a reference for creating your own HCL formatter.

## Installing

From source:
```sh
git clone git@github.com:bschaatsbergen/hclfmt
cd hclfmt
go build
```

## Usage

```sh
$ hclfmt -help
Usage: hclfmt [options] <file or directory>

Description:
  Formats all HCL configuration files to a canonical format. Supported
  configuration files (.hcl) are updated in place unless otherwise specified.

  By default, hclfmt scans the current directory for HCL configuration files.
  If you provide a directory as the target argument, hclfmt will scan that
  directory recursively when the -recursive flag is set. If you provide a file,
  hclfmt will process only that file.

Options:
  -write=true
      Write formatted output back to the source file (default: true).

  -diff
      Display diffs of formatting changes without modifying files.

  -recursive
      Recursively rewrite HCL configuration files from the specified directory.

  -help
      Show this help message.

  -version
      Display the version of hclfmt.

Examples:
  hclfmt example.hcl
      Formats the specified file.

  hclfmt -recursive ./directory
      Formats all supported HCL files in the specified directory and its subdirectories.

  hclfmt -diff example.hcl
      Displays the formatting changes for the specified file without modifying it.

Supported file extensions:
  .hcl
```

There are currently two options:

- `-write`: By default, `hclfmt` will overwrite the input file with the formatted output. You can set the `-write` option to `false` to print the formatted output to stdout.
- `-diff`: If you want to see the differences between the input and formatted output, you can set the `-diff` option to `true`.
- `-recursive`: If you want to walk the directory recursively and format all `.hcl` files, you can set the `-recursive` option to `true`.
