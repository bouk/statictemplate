package statictemplate

import (
	"go/types"
	"gopkg.in/stretchr/testify.v1/assert"
	"strings"
	"testing"
	"text/template"
)

func equalish(t *testing.T, a string, b []byte, input string) {
	assert.Equal(t, strings.Replace(strings.TrimSpace(a), "\t", "  ", -1), strings.Replace(strings.TrimSpace(string(b)), "\t", "  ", -1), input)
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

// template.tmpl(string)
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

// template.tmpl(string)
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "hi")
  return nil
}`},
		{`{{ print ( "hi" | print ) }}`, `
package main

import (
  "bou.ke/statictemplate/funcs"
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Print(funcs.Print("hi")))
  return nil
}`},
		{`{{ printf "%d" (or 0 1) }}`, `
package main

import (
  "bou.ke/statictemplate/funcs"
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

// template.tmpl(string)
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

// template.tmpl(string)
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

// template.tmpl(string)
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

// template.tmpl(string)
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

// template.tmpl(string)
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _Vara := 1
	_ = _Vara
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _Vara := "hey"
	_ = _Vara
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _Vara := 1
	_ = _Vara
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _Vara := 1
	_ = _Vara
  if eval := dot; len(eval) != 0 {
    _Vara := 2
		_ = _Vara
  }
  _Vara = 3
  return nil
}`},
		{`{{ "hi" | print }}`, `
package main

import (
  "bou.ke/statictemplate/funcs"
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Print("hi"))
  return nil
}`},
		{`{{ ( "hi" | printf "%v" ) | print }}`, `
package main

import (
  "bou.ke/statictemplate/funcs"
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Print(funcs.Printf("%v", "hi")))
  return nil
}`},
		{`{{ ( "hi" | print ) | printf "%v" }}`, `
package main

import (
  "bou.ke/statictemplate/funcs"
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Printf("%v", funcs.Print("hi")))
  return nil
}`},
		{`{{ "hi" | print | print }}`, `
package main

import (
  "bou.ke/statictemplate/funcs"
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, funcs.Print(funcs.Print("hi")))
  return nil
}`},
		{`{{ "<wow>" | html }}`, `
package main

import (
  "bou.ke/statictemplate/funcs"
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

// template.tmpl(string)
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

// template.tmpl(string)
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

// template.tmpl(string)
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

// template.tmpl(string)
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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "\n")
  _, _ = io.WriteString(w, "\n")
  if err := fun1(w, nil); err != nil {
    return err
  }
  return nil
}`},
	} {
		temp := template.Must(template.New("template.tmpl").Parse(c.input))
		actual, err := Translate(temp, "main", []TranslateInstruction{
			{"Name", "template.tmpl", types.Typ[types.String]},
		})
		if assert.NoError(t, err, c.input) {
			equalish(t, c.expected, actual, c.input)
		}
	}
}
