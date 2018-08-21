package statictemplate

import (
	"go/types"
	"testing"
	"text/template"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestComplexInput(t *testing.T) {
	stringSlice := types.NewSlice(types.Typ[types.String])
	structA := types.NewStruct([]*types.Var{types.NewVar(0, nil, "A", types.Typ[types.String])}, nil)
	structASlice := types.NewStruct([]*types.Var{types.NewVar(0, nil, "A", types.NewSlice(types.Typ[types.Int]))}, nil)
	structABool := types.NewStruct([]*types.Var{types.NewVar(0, nil, "A", types.Typ[types.Bool])}, nil)
	p := types.NewPackage("bou.ke/statictemplate/statictemplate", "statictemplate")
	emptyStruct := types.NewStruct(nil, nil)
	testStruct := types.NewNamed(types.NewTypeName(0, p, "testStruct", emptyStruct), emptyStruct, []*types.Func{
		types.NewFunc(0, p, "Hello", types.NewSignature(types.NewVar(0, p, "t", emptyStruct), types.NewTuple(), types.NewTuple(
			types.NewVar(0, p, "", types.Typ[types.String]),
		), false)),
		types.NewFunc(0, p, "Upcase", types.NewSignature(types.NewVar(0, p, "t", emptyStruct), types.NewTuple(
			types.NewVar(0, p, "input", types.Typ[types.String]),
		), types.NewTuple(
			types.NewVar(0, p, "", types.Typ[types.String]),
		), false)),
		types.NewFunc(0, p, "Bla", types.NewSignature(types.NewVar(0, p, "t", emptyStruct), types.NewTuple(), types.NewTuple(
			types.NewVar(0, p, "", types.Typ[types.Int]),
			types.NewVar(0, p, "", types.Typ[types.String]),
		), false)),
	})

	testStruct.AddMethod(
		types.NewFunc(0, p, "Recursive", types.NewSignature(types.NewVar(0, p, "t", emptyStruct), types.NewTuple(), types.NewTuple(
			types.NewVar(0, p, "", types.NewPointer(testStruct)),
		), false)),
	)

	for _, c := range []struct {
		input, expected string
		typ             types.Type
	}{
		{"{{ .A }}", `
package main

import (
  "io"
)

func Name(w io.Writer, dot struct{ A string }) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(struct{A string})
func fun0(w io.Writer, dot struct{ A string }) error {
  _, _ = io.WriteString(w, dot.A)
  return nil
}`, structA},
		{"{{ range . }}Hello{{ end }}", `
package main

import (
  "io"
)

func Name(w io.Writer, dot []string) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl([]string)
func fun0(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
		for _, dot := range eval {
			_ = dot
      _, _ = io.WriteString(w, "Hello")
    }
  }
  return nil
}`, stringSlice},
		{"{{ range $a := . }}{{ $a }}{{ end }}", `
package main

import (
  "io"
)

func Name(w io.Writer, dot []string) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl([]string)
func fun0(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
    for _, _Vara := range eval {
			dot := _Vara
			_ = dot
      _, _ = io.WriteString(w, _Vara)
    }
  }
  return nil
}`, stringSlice},
		{"{{ range $i, $a := . }}{{ $i }}{{ $a }}{{ end }}", `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot []string) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl([]string)
func fun0(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
    for _Vari, _Vara := range eval {
			_ = _Vari
			dot := _Vara
			_ = dot
      _, _ = fmt.Fprint(w, _Vari)
      _, _ = io.WriteString(w, _Vara)
    }
  }
  return nil
}`, stringSlice},
		{"{{ print .A }}", `
package main

import (
  "bou.ke/statictemplate/funcs"
  "io"
)

func Name(w io.Writer, dot struct{ A string }) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(struct{A string})
func fun0(w io.Writer, dot struct{ A string }) error {
  _, _ = io.WriteString(w, funcs.Print(dot.A))
  return nil
}`, structA},
		{"{{ (.).A }}", `
package main

import (
  "io"
)

func Name(w io.Writer, dot struct{ A string }) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(struct{A string})
func fun0(w io.Writer, dot struct{ A string }) error {
  _, _ = io.WriteString(w, dot.A)
  return nil
}`, structA},
		{"{{ (.A) }}", `
package main

import (
  "io"
)

func Name(w io.Writer, dot struct{ A string }) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(struct{A string})
func fun0(w io.Writer, dot struct{ A string }) error {
  _, _ = io.WriteString(w, dot.A)
  return nil
}`, structA},
		{"{{ with .A }} {{ . }} {{else}} {{ .A }} {{end}}", `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot struct{ A string }) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(struct{A string})
func fun0(w io.Writer, dot struct{ A string }) error {
  if eval := dot.A; len(eval) != 0 {
    dot := eval
		_ = dot
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot)
    _, _ = io.WriteString(w, " ")
  } else {
    _, _ = io.WriteString(w, " ")
    _, _ = io.WriteString(w, dot.A)
    _, _ = io.WriteString(w, " ")
  }
  return nil
}`, structA},
		{"{{ with .A }} {{ . }} {{else}} {{ .A }} {{end}}", `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot struct{ A bool }) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(struct{A bool})
func fun0(w io.Writer, dot struct{ A bool }) error {
  if eval := dot.A; eval {
    dot := eval
		_ = dot
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot)
    _, _ = io.WriteString(w, " ")
  } else {
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot.A)
    _, _ = io.WriteString(w, " ")
  }
  return nil
}`, structABool},
		{"{{ with .A }} {{ . }} {{else}} {{ .A }} {{end}}", `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot struct{ A []int }) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(struct{A []int})
func fun0(w io.Writer, dot struct{ A []int }) error {
  if eval := dot.A; len(eval) != 0 {
    dot := eval
		_ = dot
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot)
    _, _ = io.WriteString(w, " ")
  } else {
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot.A)
    _, _ = io.WriteString(w, " ")
  }
  return nil
}`, structASlice},
		{"{{ if $b := .A }}{{ $b }}{{end}}", `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot struct{ A []int }) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(struct{A []int})
func fun0(w io.Writer, dot struct{ A []int }) error {
  if eval := dot.A; len(eval) != 0 {
    _Varb := eval
		_ = _Varb
    _, _ = fmt.Fprint(w, _Varb)
  }
  return nil
}`, structASlice},
		{`{{ .Hello }}`, `
package main

import (
  pkg1 "bou.ke/statictemplate/statictemplate"
  "io"
)

func Name(w io.Writer, dot pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(pkg1.testStruct)
func fun0(w io.Writer, dot pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Hello())
  return nil
}`, testStruct},
		{`{{ .Hello }}`, `
package main

import (
  pkg1 "bou.ke/statictemplate/statictemplate"
  "io"
)

func Name(w io.Writer, dot *pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Hello())
  return nil
}`, types.NewPointer(testStruct)},
		{`{{ .Recursive.Recursive.Recursive.Upcase "whatup" }}`, `
package main

import (
  pkg1 "bou.ke/statictemplate/statictemplate"
  "io"
)

func Name(w io.Writer, dot pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(pkg1.testStruct)
func fun0(w io.Writer, dot pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Recursive().Recursive().Recursive().Upcase("whatup"))
  return nil
}`, testStruct},
		{`{{ ( .Recursive.Recursive ).Recursive.Upcase "whatup" }}`, `
package main

import (
  pkg1 "bou.ke/statictemplate/statictemplate"
  "io"
)

func Name(w io.Writer, dot pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(pkg1.testStruct)
func fun0(w io.Writer, dot pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Recursive().Recursive().Recursive().Upcase("whatup"))
  return nil
}`, testStruct},
		{`{{ .Hello | printf "%q" }}`, `
package main

import (
  "bou.ke/statictemplate/funcs"
  pkg1 "bou.ke/statictemplate/statictemplate"
  "io"
)

func Name(w io.Writer, dot pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(pkg1.testStruct)
func fun0(w io.Writer, dot pkg1.testStruct) error {
  _, _ = io.WriteString(w, funcs.Printf("%q", dot.Hello()))
  return nil
}`, testStruct},
		{`{{ .Upcase "whatup" }}`, `
package main

import (
  pkg1 "bou.ke/statictemplate/statictemplate"
  "io"
)

func Name(w io.Writer, dot pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(pkg1.testStruct)
func fun0(w io.Writer, dot pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Upcase("whatup"))
  return nil
}`, testStruct},
		{`{{ "whatup" | .Upcase  }}`, `
package main

import (
  pkg1 "bou.ke/statictemplate/statictemplate"
  "io"
)

func Name(w io.Writer, dot pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// template.tmpl(pkg1.testStruct)
func fun0(w io.Writer, dot pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Upcase("whatup"))
  return nil
}`, testStruct},
		{`{{ .Bla }}`, `
package main

import (
  pkg1 "bou.ke/statictemplate/statictemplate"
  "fmt"
  "io"
)

func Name(w io.Writer, dot pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

func fun2(value int, err error) int {
  if err != nil {
    panic(err)
  }
  return value
}

// template.tmpl(pkg1.testStruct)
func fun0(w io.Writer, dot pkg1.testStruct) error {
  _, _ = fmt.Fprint(w, fun2(dot.Bla()))
  return nil
}
`, testStruct},
		{`{{define "T1"}}{{ . }}{{end}}
{{define "T2"}}TWO {{template "T1" .Hello}}{{end}}
{{define "T3"}}{{template "T1" .}} {{template "T2" .}}{{end}}
{{template "T3" .}}`, `
package main

import (
  pkg1 "bou.ke/statictemplate/statictemplate"
  "fmt"
  "io"
)

func Name(w io.Writer, dot *pkg1.testStruct) (err error) {
  defer func() {
    if recovered := recover(); recovered != nil {
      var ok bool
      if err, ok = recovered.(error); !ok {
        panic(recovered)
      }
    }
  }()
  return fun0(w, dot)
}

// T1(*pkg1.testStruct)
func fun3(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = fmt.Fprint(w, dot)
  return nil
}

// T1(string)
func fun5(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, dot)
  return nil
}

// T2(*pkg1.testStruct)
func fun4(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, "TWO ")
  if err := fun5(w, dot.Hello()); err != nil {
    return err
  }
  return nil
}

// T3(*pkg1.testStruct)
func fun2(w io.Writer, dot *pkg1.testStruct) error {
  if err := fun3(w, dot); err != nil {
    return err
  }
  _, _ = io.WriteString(w, " ")
  if err := fun4(w, dot); err != nil {
    return err
  }
  return nil
}

// template.tmpl(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, "\n")
  _, _ = io.WriteString(w, "\n")
  _, _ = io.WriteString(w, "\n")
  if err := fun2(w, dot); err != nil {
    return err
  }
  return nil
}`, types.NewPointer(testStruct)},
	} {
		temp := template.Must(template.New("template.tmpl").Parse(c.input))
		actual, err := Translate(temp, "main", []TranslateInstruction{
			{"Name", "template.tmpl", c.typ},
		})
		if assert.NoError(t, err, c.input) {
			equalish(t, c.expected, actual, c.input)
		}
	}
}
