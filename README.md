# hclfmt

An HCL (HashiCorp Configuration Language) formatter.

### You won't quickly need this
At HashiCorp we recommend that products using HCL implement their own formatter, such as `terraform fmt` or `nomad fmt`. Applications typically extend the generic formatting rules to include domain-specific logic, ensuring that formatting aligns with the idiomatic structures they recommend in their documentation. Think of this repository as a reference implementation for a generic HCL formatter, that you can use as a starting point for your own implementation—or if you just need a quick way to format a plain HCL file.

Applications that do implement their own formatter typically build on top of the generic HCL formatting process but extend it with logic to handle domain-specific constructs and idiomatic conventions. Here’s how that process generally works:

1. Applications use [hclwrite](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hclwrite) parser to generate a hybrid structure that combines an abstract syntax tree (AST) with a physical syntax tree. This hybrid structure allows for both logical analysis and any surgical changes to normalize the configuration.
2. The application adds custom logic to normalize HCL configurations by rewriting constructs into the idiomatic form preferred by the application. This may involve reordering attributes, adjusting whitespace, modifying indentation, or enforcing specific conventions for expressions and relationships within the configuration.
3. Once the normalization is complete, the formatter serializes the updated AST back into HCL syntax using [hclwrite](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hclwrite) package's formatter.

Bear in mind that any normalization process must be idempotent, meaning that running the formatter multiple times on the same input should produce the same result. If it does not, consider it a bug that needs to be fixed.

Whatever you do, don’t use this tool as a replacement for a proper application-specific formatter.

### "Isn't it all just HCL?"
[@apparentlymart](https://github.com/apparentlymart) does a much better job explaining this than I ever could:
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
Usage: hclfmt [-help] [options] [-version] [args]
```

There are currently two options:

- `-write`: By default, `hclfmt` will overwrite the input file with the formatted output. You can set the `-write` option to `false` to print the formatted output to stdout.
- `-diff`: If you want to see the differences between the input and formatted output, you can set the `-diff` option to `true`.

