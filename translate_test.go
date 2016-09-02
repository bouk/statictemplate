package statictemplate

import (
	"gopkg.in/stretchr/testify.v1/assert"
	"reflect"
	"strings"
	"testing"
)

func equalish(t *testing.T, a, b, input string) {
	assert.Equal(t, strings.Replace(strings.TrimSpace(a), "\t", "  ", -1), strings.Replace(strings.TrimSpace(b), "\t", "  ", -1), input)
}

func TestTranslate(t *testing.T) {
	for _, c := range []struct {
		input, expected string
	}{
		{"hello", `
func Name(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "hello")
  return nil
}`},
		{"hi{{/* comment*/}}there", `
func Name(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "hi")
  _, _ = io.WriteString(w, "there")
  return nil
}`},
		{`{{ "hi" }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, "hi")
  return nil
}`},
		{`{{ print ( "hi" | print ) }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, fmt.Sprint(fmt.Sprint("hi")))
  return nil
}`},
		{`{{ 1 }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, 1)
  return nil
}`},
		{`{{ . }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, dot)
  return nil
}`},
		{`{{ true }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, true)
  return nil
}`},
		{`{{ false }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, false)
  return nil
}`},
		{`{{ $a := 1 }}{{ $a }}`, `
func Name(w io.Writer, dot string) error {
  _Vara := 1
  _, _ = fmt.Fprint(w, _Vara)
  return nil
}`},
		{`{{ $a := 1 }}{{ $a := 2 }}`, `
func Name(w io.Writer, dot string) error {
  _Vara := 1
  _Vara = 2
  return nil
}`},
		{`{{ $a := 1 }}{{ if . }}{{ $a := 2 }}{{ end }}{{ $a := 3 }}`, `
func Name(w io.Writer, dot string) error {
  _Vara := 1
  if eval := dot; len(eval) != 0 {
    _Vara := 2
  }
  _Vara = 3
  return nil
}`},
		{`{{ "hi" | print }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, fmt.Sprint("hi"))
  return nil
}`},
		{`{{ ( "hi" | printf "%v" ) | print }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, fmt.Sprint(fmt.Sprintf("%v", "hi")))
  return nil
}`},
		{`{{ ( "hi" | print ) | printf "%v" }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, fmt.Sprintf("%v", fmt.Sprint("hi")))
  return nil
}`},
		{`{{ "hi" | print | print }}`, `
func Name(w io.Writer, dot string) error {
  _, _ = fmt.Fprint(w, fmt.Sprint(fmt.Sprint("hi")))
  return nil
}`},
		{`{{ if true }}a{{end}}`, `
func Name(w io.Writer, dot string) error {
  if eval := true; eval {
    _, _ = io.WriteString(w, "a")
  }
  return nil
}`},
		{`{{ if true }}a{{else}}b{{end}}`, `
func Name(w io.Writer, dot string) error {
  if eval := true; eval {
    _, _ = io.WriteString(w, "a")
  } else {
    _, _ = io.WriteString(w, "b")
  }
  return nil
}`},
	} {
		actual, err := Translate("Name", c.input, reflect.TypeOf(""))
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

func TestComplexInput(t *testing.T) {
	for _, c := range []struct {
		input, expected string
		typ             interface{}
	}{
		{"{{ .A }}", `
func Name(w io.Writer, dot struct{ A string }) error {
  _, _ = fmt.Fprint(w, dot.A)
  return nil
}`, struct{ A string }{""}},
		{"{{ range . }}Hello{{ end }}", `
func Name(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
    for range eval {
      _, _ = io.WriteString(w, "Hello")
    }
  }
  return nil
}`, []string{"hi"}},
		{"{{ range $a := . }}{{ $a }}{{ end }}", `
func Name(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
    for _, _Vara := range eval {
      _, _ = fmt.Fprint(w, _Vara)
    }
  }
  return nil
}`, []string{"hi"}},
		{"{{ range $i, $a := . }}{{ $i }}{{ $a }}{{ end }}", `
func Name(w io.Writer, dot []string) error {
  if eval := dot; len(eval) != 0 {
    for _Vari, _Vara := range eval {
      _, _ = fmt.Fprint(w, _Vari)
      _, _ = fmt.Fprint(w, _Vara)
    }
  }
  return nil
}`, []string{"hi"}},
		{"{{ print .A }}", `
func Name(w io.Writer, dot struct{ A string }) error {
  _, _ = fmt.Fprint(w, fmt.Sprint(dot.A))
  return nil
}`, struct{ A string }{""}},
		{"{{ (.).A }}", `
func Name(w io.Writer, dot struct{ A string }) error {
  _, _ = fmt.Fprint(w, dot.A)
  return nil
}`, struct{ A string }{""}},
		{"{{ (.A) }}", `
func Name(w io.Writer, dot struct{ A string }) error {
  _, _ = fmt.Fprint(w, dot.A)
  return nil
}`, struct{ A string }{""}},
		{"{{ with .A }} {{ . }} {{else}} {{ .A }} {{end}}", `
func Name(w io.Writer, dot struct{ A string }) error {
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
}`, struct{ A string }{""}},
		{"{{ with .A }} {{ . }} {{else}} {{ .A }} {{end}}", `
func Name(w io.Writer, dot struct{ A bool }) error {
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
func Name(w io.Writer, dot struct{ A []int }) error {
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
func Name(w io.Writer, dot struct{ A []int }) error {
  if eval := dot.A; len(eval) != 0 {
    _Varb := eval
    _, _ = fmt.Fprint(w, _Varb)
  }
  return nil
}`, struct{ A []int }{nil}},
		{`{{ .Hello }}`, `
func Name(w io.Writer, dot *statictemplate.testStruct) error {
  _, _ = fmt.Fprint(w, dot.Hello())
  return nil
}`, &testStruct{}},
		{`{{ .Recursive.Recursive.Recursive.Upcase "whatup" }}`, `
func Name(w io.Writer, dot *statictemplate.testStruct) error {
  _, _ = fmt.Fprint(w, dot.Recursive().Recursive().Recursive().Upcase("whatup"))
  return nil
}`, &testStruct{}},
		{`{{ ( .Recursive.Recursive ).Recursive.Upcase "whatup" }}`, `
func Name(w io.Writer, dot *statictemplate.testStruct) error {
  _, _ = fmt.Fprint(w, dot.Recursive().Recursive().Recursive().Upcase("whatup"))
  return nil
}`, &testStruct{}},
		{`{{ .Hello | printf "%q" }}`, `
func Name(w io.Writer, dot *statictemplate.testStruct) error {
  _, _ = fmt.Fprint(w, fmt.Sprintf("%q", dot.Hello()))
  return nil
}`, &testStruct{}},
		{`{{ .Upcase "whatup" }}`, `
func Name(w io.Writer, dot *statictemplate.testStruct) error {
  _, _ = fmt.Fprint(w, dot.Upcase("whatup"))
  return nil
}`, &testStruct{}},
		{`{{ "whatup" | .Upcase  }}`, `
func Name(w io.Writer, dot *statictemplate.testStruct) error {
  _, _ = fmt.Fprint(w, dot.Upcase("whatup"))
  return nil
}`, &testStruct{}},
	} {
		actual, err := Translate("Name", c.input, reflect.TypeOf(c.typ))
		if assert.NoError(t, err, c.input) {
			equalish(t, c.expected, actual, c.input)
		}
	}
}
