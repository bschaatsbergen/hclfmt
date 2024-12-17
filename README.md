# hclfmt

hclfmt is a simple CLI tool to format HashiCorp Configuration Language (HCL) files. 

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

