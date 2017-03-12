package statictemplate

import (
	"gopkg.in/stretchr/testify.v1/assert"
	"html/template"
	"reflect"
	"testing"
)

func TestCompileHTMLTemplate(t *testing.T) {
	for _, c := range []struct {
		input, expected string
	}{
		{`<!doctype html>
<html>
<head>
<title>{{ . }}</title>
</head>
<body ref="{{ . }}">
{{ . }}
</body>
</html>
`, `package main

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

// template.tmpl(string)
func fun0(w io.Writer, dot string) error {
  _, _ = io.WriteString(w, "<!doctype html>\n<html>\n<head>\n<title>")
  _, _ = io.WriteString(w, funcs.Rcdataescaper(dot))
  _, _ = io.WriteString(w, "</title>\n</head>\n<body ref=\"")
  _, _ = io.WriteString(w, funcs.Attrescaper(dot))
  _, _ = io.WriteString(w, "\">\n")
  _, _ = io.WriteString(w, funcs.Htmlescaper(dot))
  _, _ = io.WriteString(w, "\n</body>\n</html>\n")
  return nil
}`},
	} {
		temp := template.Must(template.New("template.tmpl").Parse(c.input))
		actual, err := Translate(temp, "main", []TranslateInstruction{
			{"Name", "template.tmpl", reflect.TypeOf("")},
		})
		if assert.NoError(t, err, c.input) {
			equalish(t, c.expected, actual, c.input)
		}
	}
}
