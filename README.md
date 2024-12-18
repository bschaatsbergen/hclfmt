# hclfmt

An HCL (HashiCorp Configuration Language) formatter.

### You won't quickly need this
Our recommendation is that products implementing HCL provide their own formatters, as seen in tools like Terraform, Packer, and others. This allows applications to extend the generic formatting rules with domain-specific logic, ensuring that formatting not only adheres to general HCL conventions but also aligns with the idiomatic patterns and best practices outlined in their documentation. Sometimes this is something simple as reordering attributes or adjusting indentation, but it also extends to enforcing specific conventions for expressions and relationships within the configuration.

Applications that do implement their own formatter often build on the generic HCL formatting process, extending it with additional logic to support domain-specific constructs and enforce idiomatic conventions based on their configuration standards.

1. **Parse the HCL configuration**  
   Use the [`hclwrite`](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hclwrite) package to parse the HCL configuration. This generates a hybrid syntax tree that combines an abstract / physical syntax tree (AST). This allows the application to make any surgical changes where necessary.

2. **Normalize the configuration**  
   Add custom logic to normalize the HCL configuration according to the application's idiomatic preferences. This process may include:
   - Reordering attributes to align with conventions.
   - Adjusting whitespace and indentation for readability and consistency.
   - Enforcing domain-specific conventions for expressions, object relationships, and configuration constructs.

3. **Serialize back to HCL**
   After normalization, serialize the updated syntax tree back into HCL syntax using the [`hclwrite`](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hclwrite) package. This ensures the final configuration is clean, consistent, and adheres to both HCL conventions and the application's formatting standards.

Bear in mind that any normalization process must be **idempotent**, meaning running the formatter multiple times on the same input should always produce the same result. If it doesn’t, treat it as a bug that needs to be addressed.

Above all, don’t rely on this tool as a substitute for a proper, application-specific formatter.

### "Isn't it all just HCL?"
[@apparentlymart](https://github.com/apparentlymart) explains this far better than I ever could:
> Unlike some other formats like JSON and YAML, a HCL file is more like a program to be executed than a data structure to be parsed, and so there's considerably more application-level interpretation to be done than you might be accustomed to with other grammars.

> HCL is designed as a toolkit for building languages rather than as a language in its own right, but it's true that a bunch of the existing HCL-based languages aren't doing that much above what HCL itself offers, aside from defining their expected block types and attributes.

> The languages that allow for e.g. creating relationships between declared objects via expressions, or writing "libraries" like Terraform's modules, will tend to bend HCL in more complicated ways than where HCL is being used mainly just as a serialization of a flat data structure. To be specific, I would expect the Terraform Language, the Packer Language and the Waypoint Language to all eventually benefit from application-specific extensions with their own formatters, but something like Vault's policy language or Consul's agent configuration files would probably suffice with a generic HCL extension and generic formatter.

## Installing

```sh
brew install <coming-soon>
```

From source:
```sh
git clone git@github.com:bschaatsbergen/hclfmt
cd hclfmt
make
```

Pre-built packages for Darwin and Linux are also available on the [Releases page](https://github.com/bschaatsbergen/hclfmt/releases).

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
