package statictemplate

import (
	"gopkg.in/stretchr/testify.v1/assert"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

func equalish(t *testing.T, a, b, input string) {
	assert.Equal(t, strings.Replace(strings.TrimSpace(a), "\t", "  ", -1), strings.Replace(strings.TrimSpace(b), "\t", "  ", -1), input)
}

func TestTranslate(t *testing.T) {
	for _, c := range []struct {
		input, expected string
	}{
		{"hello", `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "hello")
  return nil
}`},
		{"hi{{/* comment*/}}there", `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "hi")
  _, _ = io.WriteString(w, "there")
  return nil
}`},
		{`{{ "hi" }}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "hi")
  return nil
}`},
		{`{{ print ( "hi" | print ) }}`, `
package main

import (
  "github.com/bouk/statictemplate/funcs"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Print(funcs.Print("hi")))
  return nil
}`},
		{`{{ printf "%d" (or 0 1) }}`, `
package main

import (
  "github.com/bouk/statictemplate/funcs"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Printf("%d", funcs.Or(0, 1)))
  return nil
}`},
		{`{{ 1 }}`, `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, 1)
  return nil
}`},
		{`{{ . }}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, dot)
  return nil
}`},
		{`{{ true }}`, `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, true)
  return nil
}`},
		{`{{ false }}`, `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, false)
  return nil
}`},
		{`{{ $a := 1 }}{{ $a }}`, `
package main

import (
  "fmt"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _Vara := 1
  _, _ = fmt.Fprint(w, _Vara)
  return nil
}`},
		{`{{ $a := "hey" }}{{ $a }}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _Vara := "hey"
  _, _ = io.WriteString(w, _Vara)
  return nil
}`},
		{`{{ $a := 1 }}{{ $a := 2 }}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _Vara := 1
  _Vara = 2
  return nil
}`},
		{`{{ $a := 1 }}{{ if . }}{{ $a := 2 }}{{ end }}{{ $a := 3 }}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _Vara := 1
  if eval := dot; len(eval) != 0 {
    _Vara := 2
  }
  _Vara = 3
  return nil
}`},
		{`{{ "hi" | print }}`, `
package main

import (
  "github.com/bouk/statictemplate/funcs"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Print("hi"))
  return nil
}`},
		{`{{ ( "hi" | printf "%v" ) | print }}`, `
package main

import (
  "github.com/bouk/statictemplate/funcs"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Print(funcs.Printf("%v", "hi")))
  return nil
}`},
		{`{{ ( "hi" | print ) | printf "%v" }}`, `
package main

import (
  "github.com/bouk/statictemplate/funcs"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Printf("%v", funcs.Print("hi")))
  return nil
}`},
		{`{{ "hi" | print | print }}`, `
package main

import (
  "github.com/bouk/statictemplate/funcs"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Print(funcs.Print("hi")))
  return nil
}`},
		{`{{ "<wow>" | html }}`, `
package main

import (
  "github.com/bouk/statictemplate/funcs"
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Html("<wow>"))
  return nil
}`},
		{`{{ if true }}a{{end}}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  if eval := true; eval {
    _, _ = io.WriteString(w, "a")
  }
  return nil
}`},
		{`{{ if true }}a{{else}}b{{end}}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// Name(string)
func fun0(w io.Writer, dot string) error {
  if eval := true; eval {
    _, _ = io.WriteString(w, "a")
  } else {
    _, _ = io.WriteString(w, "b")
  }
  return nil
}`},
		{`{{define "T1"}}ONE{{end}}
{{define "T2"}}TWO {{template "T1"}}{{end}}
{{define "T3"}}{{template "T1"}} {{template "T2"}}{{end}}
{{template "T3"}}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// T1(nil)
func fun2(w io.Writer, dot interface{}) error {
  _, _ = io.WriteString(w, "ONE")
  return nil
}

// T2(nil)
func fun3(w io.Writer, dot interface{}) error {
  _, _ = io.WriteString(w, "TWO ")
  if err := fun2(w, nil); err != nil {
    return err
  }
  return nil
}

// T3(nil)
func fun1(w io.Writer, dot interface{}) error {
  if err := fun2(w, nil); err != nil {
    return err
  }
  _, _ = io.WriteString(w, " ")
  if err := fun3(w, nil); err != nil {
    return err
  }
  return nil
}

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "\n")
  _, _ = io.WriteString(w, "\n")
  _, _ = io.WriteString(w, "\n")
  if err := fun1(w, nil); err != nil {
    return err
  }
  return nil
}`},
		{`
{{define "T1"}}{{if .}}TWO{{else}}ONE{{template "T1" true}}{{end}}{{end}}
{{template "T1"}}`, `
package main

import (
  "io"
)

func Name(w io.Writer, dot string) (err error) {
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

// T1(bool)
func fun2(w io.Writer, dot bool) error {
  if eval := dot; eval {
    _, _ = io.WriteString(w, "TWO")
  } else {
    _, _ = io.WriteString(w, "ONE")
    if err := fun2(w, true); err != nil {
      return err
    }
  }
  return nil
}

// T1(nil)
func fun1(w io.Writer, dot interface{}) error {
  if eval := dot; eval != nil {
    _, _ = io.WriteString(w, "TWO")
  } else {
    _, _ = io.WriteString(w, "ONE")
    if err := fun2(w, true); err != nil {
      return err
    }
  }
  return nil
}

// Name(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "\n")
  _, _ = io.WriteString(w, "\n")
  if err := fun1(w, nil); err != nil {
    return err
  }
  return nil
}`},
	} {
		actual, err := Translate("main", "Name", c.input, reflect.TypeOf(""))
		if assert.NoError(t, err, c.input) {
			equalish(t, c.expected, actual, c.input)
		}
	}
}

type testStruct struct{}

func (t *testStruct) Hello() string {
	return "hi"
}

func (t *testStruct) Upcase(input string) string {
	return strings.ToUpper(input)
}

func (t *testStruct) Recursive() *testStruct {
	return t
}

func (t *testStruct) Bla() (int, error) {
	return 1, nil
}

func TestComplexInput(t *testing.T) {
	for _, c := range []struct {
		input, expected string
		typ             interface{}
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

// Name(struct { A string })
func fun0(w io.Writer, dot struct{ A string }) error {
  _, _ = io.WriteString(w, dot.A)
  return nil
}`, struct{ A string }{""}},
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

// Name([]string)
func fun0(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
    for range eval {
      _, _ = io.WriteString(w, "Hello")
    }
  }
  return nil
}`, []string{"hi"}},
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

// Name([]string)
func fun0(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
    for _, _Vara := range eval {
      _, _ = io.WriteString(w, _Vara)
    }
  }
  return nil
}`, []string{"hi"}},
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

// Name([]string)
func fun0(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
    for _Vari, _Vara := range eval {
      _, _ = fmt.Fprint(w, _Vari)
      _, _ = io.WriteString(w, _Vara)
    }
  }
  return nil
}`, []string{"hi"}},
		{"{{ print .A }}", `
package main

import (
  "github.com/bouk/statictemplate/funcs"
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

// Name(struct { A string })
func fun0(w io.Writer, dot struct{ A string }) error {
  _, _ = io.WriteString(w, funcs.Print(dot.A))
  return nil
}`, struct{ A string }{""}},
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

// Name(struct { A string })
func fun0(w io.Writer, dot struct{ A string }) error {
  _, _ = io.WriteString(w, dot.A)
  return nil
}`, struct{ A string }{""}},
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

// Name(struct { A string })
func fun0(w io.Writer, dot struct{ A string }) error {
  _, _ = io.WriteString(w, dot.A)
  return nil
}`, struct{ A string }{""}},
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

// Name(struct { A string })
func fun0(w io.Writer, dot struct{ A string }) error {
  if eval := dot.A; len(eval) != 0 {
    dot := eval
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot)
    _, _ = io.WriteString(w, " ")
  } else {
    _, _ = io.WriteString(w, " ")
    _, _ = io.WriteString(w, dot.A)
    _, _ = io.WriteString(w, " ")
  }
  return nil
}`, struct{ A string }{""}},
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

// Name(struct { A bool })
func fun0(w io.Writer, dot struct{ A bool }) error {
  if eval := dot.A; eval {
    dot := eval
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot)
    _, _ = io.WriteString(w, " ")
  } else {
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot.A)
    _, _ = io.WriteString(w, " ")
  }
  return nil
}`, struct{ A bool }{true}},
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

// Name(struct { A []int })
func fun0(w io.Writer, dot struct{ A []int }) error {
  if eval := dot.A; len(eval) != 0 {
    dot := eval
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot)
    _, _ = io.WriteString(w, " ")
  } else {
    _, _ = io.WriteString(w, " ")
    _, _ = fmt.Fprint(w, dot.A)
    _, _ = io.WriteString(w, " ")
  }
  return nil
}`, struct{ A []int }{nil}},
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

// Name(struct { A []int })
func fun0(w io.Writer, dot struct{ A []int }) error {
  if eval := dot.A; len(eval) != 0 {
    _Varb := eval
    _, _ = fmt.Fprint(w, _Varb)
  }
  return nil
}`, struct{ A []int }{nil}},
		{`{{ .Hello }}`, `
package main

import (
  pkg1 "github.com/bouk/statictemplate"
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

// Name(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Hello())
  return nil
}`, &testStruct{}},
		{`{{ .Recursive.Recursive.Recursive.Upcase "whatup" }}`, `
package main

import (
  pkg1 "github.com/bouk/statictemplate"
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

// Name(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Recursive().Recursive().Recursive().Upcase("whatup"))
  return nil
}`, &testStruct{}},
		{`{{ ( .Recursive.Recursive ).Recursive.Upcase "whatup" }}`, `
package main

import (
  pkg1 "github.com/bouk/statictemplate"
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

// Name(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Recursive().Recursive().Recursive().Upcase("whatup"))
  return nil
}`, &testStruct{}},
		{`{{ .Hello | printf "%q" }}`, `
package main

import (
  pkg1 "github.com/bouk/statictemplate"
  "github.com/bouk/statictemplate/funcs"
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

// Name(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, funcs.Printf("%q", dot.Hello()))
  return nil
}`, &testStruct{}},
		{`{{ .Upcase "whatup" }}`, `
package main

import (
  pkg1 "github.com/bouk/statictemplate"
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

// Name(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Upcase("whatup"))
  return nil
}`, &testStruct{}},
		{`{{ "whatup" | .Upcase  }}`, `
package main

import (
  pkg1 "github.com/bouk/statictemplate"
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

// Name(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, dot.Upcase("whatup"))
  return nil
}`, &testStruct{}},
		{`{{ .Bla }}`, `
package main

import (
  "fmt"
  pkg1 "github.com/bouk/statictemplate"
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

func fun2(value int, err error) int {
  if err != nil {
    panic(err)
  }
  return value
}

// Name(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = fmt.Fprint(w, fun2(dot.Bla()))
  return nil
}
`, &testStruct{}},
		{`{{define "T1"}}{{ . }}{{end}}
{{define "T2"}}TWO {{template "T1" .Hello}}{{end}}
{{define "T3"}}{{template "T1" .}} {{template "T2" .}}{{end}}
{{template "T3" .}}`, `
package main

import (
  "fmt"
  pkg1 "github.com/bouk/statictemplate"
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

// Name(*pkg1.testStruct)
func fun0(w io.Writer, dot *pkg1.testStruct) error {
  _, _ = io.WriteString(w, "\n")
  _, _ = io.WriteString(w, "\n")
  _, _ = io.WriteString(w, "\n")
  if err := fun2(w, dot); err != nil {
    return err
  }
  return nil
}`, &testStruct{}},
		{`{{ (.Parse "hey").Name }}`, `
package main

import (
  "io"
  "text/template"
)

func Name(w io.Writer, dot *template.Template) (err error) {
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

func fun1(value *template.Template, err error) *template.Template {
  if err != nil {
    panic(err)
  }
  return value
}

// Name(*template.Template)
func fun0(w io.Writer, dot *template.Template) error {
  _, _ = io.WriteString(w, fun1(dot.Parse("hey")).Name())
  return nil
}`, template.New("hi")},
	} {
		actual, err := Translate("main", "Name", c.input, reflect.TypeOf(c.typ))
		if assert.NoError(t, err, c.input) {
			equalish(t, c.expected, actual, c.input)
		}
	}
}
