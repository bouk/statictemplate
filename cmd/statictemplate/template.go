package main

import (
	"fmt"
	"io"
)

func writeTemplate(w io.Writer, targets compilationTargets, templateFiles []string, html bool, funcMapImport, funcMapName string) error {
	io.WriteString(w, `package main

  import (
    "github.com/bouk/statictemplate"
    "log"
    "os"
    "reflect"
`)
	if html {
		io.WriteString(w, `"html/template"
`)
	} else {
		io.WriteString(w, `"text/template"
`)
	}
	if funcMapImport != "" {
		io.WriteString(w, funcMapImport)
	}
	for i, target := range targets {
		if target.dot.packagePath != "" {
			fmt.Fprintf(w, "pkg%d %q\n", i, target.dot.packagePath)
		}
	}
	io.WriteString(w, `)

  func main() {
    var (
  `)
	for i, target := range targets {
		if target.dot.packagePath == "" {
			fmt.Fprintf(w, "dot%d %s%s\n", i, target.dot.prefix, target.dot.typeName)
		} else {
			fmt.Fprintf(w, "dot%d %spkg%d.%s\n", i, target.dot.prefix, i, target.dot.typeName)
		}
	}
	io.WriteString(w, `  )
    tmpl, err := template.New("")`)
	if funcMapName != "" {
		fmt.Fprintf(w, ".Funcs(%s)", funcMapName)
	}
	io.WriteString(w, ".ParseFiles(")
	for _, templateFile := range templateFiles {
		fmt.Fprintf(w, "%q,\n", templateFile)
	}
	io.WriteString(w, `)
    if err != nil {
      log.Fatal(err)
    }
		translator := statictemplate.New(tmpl)
`)
	if funcMapName != "" {
		fmt.Fprintf(w, "translator.Funcs = %s\n", funcMapName)
	}
	io.WriteString(w, `code, err := translator.Translate("template", []statictemplate.TranslateInstruction{
  `)

	for i, target := range targets {
		fmt.Fprintf(w, "{%q, %q, reflect.TypeOf(dot%d)},\n", target.functionName, target.templateName, i)
	}
	io.WriteString(w, `})

    if err != nil {
      log.Fatal(err)
    }

    os.Stdout.Write(code)
  }`)
	return nil
}
