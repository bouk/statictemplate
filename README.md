# statictemplate

Statictemplate is a code generator for Go's text/template and html/template packages. It works by reading in the template files, and generating the needed functions based on the combination of requested function names and type signatures.

Please read [my blogpost](http://bouk.co/blog/code-generating-code/) about this project for some background.

## Installation

To install the commandline tool, run `go get github.com/bouk/statictemplate`.

## Usage

These are the supported flags:

```
Usage of statictemplate:
  -dev string
        Name of the dev output file
  -funcs string
        A reference to a custom Funcs map to include
  -html
        Interpret templates as HTML, to enable Go's automatic HTML escaping
  -o string
        Name of the output file (default "template.go")
  -package string
        Name of the package of the result file. Defaults to name of the folder of the output file
  -t value
        Target to process, supports multiple. The format is <function name>:<template name>:<type of the template argument>
```

After the flags you pass in one or more globs to specify the templates.


The example in this project uses the following command

```
statictemplate -html -o example/template/template.go -t "Index:index.tmpl:[]github.com/bouk/statictemplate/example.Post" example/template/*.tmpl
```

## Docs

[Check out the docs](https://godoc.org/github.com/bouk/statictemplate/statictemplate).
