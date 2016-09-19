package main

import (
	"fmt"
	"io"
)

func writeDevTemplate(w io.Writer, targets compilationTargets, templateFiles []string, html bool, funcMapImport, funcMapName string, pkg string) error {
	fmt.Fprintf(w, `// +build dev

package %s

import (
  "io"
`, pkg)
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
	io.WriteString(w, ")\n")
	for i, target := range targets {
		var dot string
		if target.dot.packagePath == "" {
			dot = fmt.Sprintf("%s%s", target.dot.prefix, target.dot.typeName)
		} else {
			dot = fmt.Sprintf("%spkg%d.%s", target.dot.prefix, i, target.dot.typeName)
		}
		fmt.Fprintf(w, `func %s(w io.Writer, dot %s) error {
  temp, err := template.New("")`, target.functionName, dot)
		if funcMapName != "" {
			fmt.Fprintf(w, ".Funcs(%s)", funcMapName)
		}
		io.WriteString(w, ".ParseFiles(\n")
		for _, templateFile := range templateFiles {
			fmt.Fprintf(w, "%q,\n", templateFile)
		}
		io.WriteString(w, `)
        if err != nil {
          return err
        }
        return temp.Execute(w, dot)
}
`)
	}
	return nil
}
