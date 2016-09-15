package template

import (
	"github.com/bouk/statictemplate/funcs"
	"io"
	"text/template"
)

func Hi(w io.Writer, dot string) (err error) {
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

func Hello(w io.Writer, dot *template.Template) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			var ok bool
			if err, ok = recovered.(error); !ok {
				panic(recovered)
			}
		}
	}()
	return fun2(w, dot)
}

// notice.tmpl(nil)
func fun1(w io.Writer, dot interface{}) error {
	_, _ = io.WriteString(w, "Hello\n")
	return nil
}

// hi.tmpl(string)
func fun0(w io.Writer, dot string) error {
	if err := fun1(w, nil); err != nil {
		return err
	}
	_, _ = io.WriteString(w, " ")
	_, _ = io.WriteString(w, funcs.Htmlescaper(dot))
	_, _ = io.WriteString(w, "!\n")
	return nil
}

// hi.tmpl(*template.Template)
func fun2(w io.Writer, dot *template.Template) error {
	if err := fun1(w, nil); err != nil {
		return err
	}
	_, _ = io.WriteString(w, " ")
	_, _ = io.WriteString(w, funcs.Htmlescaper(dot))
	_, _ = io.WriteString(w, "!\n")
	return nil
}
