package template

import (
	pkg1 "bou.ke/statictemplate/example"
	"bou.ke/statictemplate/funcs"
	"io"
)

func Index(w io.Writer, dot []pkg1.Post) (err error) {
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

// header.tmpl(string)
func fun2(w io.Writer, dot string) error {
	_, _ = io.WriteString(w, "<!doctype html>\n<html>\n  <head>\n    ")
	if eval := dot; len(eval) != 0 {
		_, _ = io.WriteString(w, "\n    <title>Bouke's Blog | ")
		_, _ = io.WriteString(w, funcs.Rcdataescaper(dot))
		_, _ = io.WriteString(w, "</title>\n    ")
	} else {
		_, _ = io.WriteString(w, "\n    <title>Bouke's Blog</title>\n    ")
	}
	_, _ = io.WriteString(w, "\n  </head>\n  <body>\n")
	return nil
}

// post.tmpl(pkg1.Post)
func fun3(w io.Writer, dot pkg1.Post) error {
	_, _ = io.WriteString(w, "<article>\n  <h2>")
	_, _ = io.WriteString(w, funcs.Htmlescaper(dot.Title))
	_, _ = io.WriteString(w, "</h2>\n  <p>")
	_, _ = io.WriteString(w, funcs.Htmlescaper(dot.Body))
	_, _ = io.WriteString(w, "</h2>\n</article>\n")
	return nil
}

// footer.tmpl(nil)
func fun4(w io.Writer, dot interface{}) error {
	_, _ = io.WriteString(w, "</body>\n</html>\n")
	return nil
}

// index.tmpl([]pkg1.Post)
func fun0(w io.Writer, dot []pkg1.Post) error {
	if err := fun2(w, "Index"); err != nil {
		return err
	}
	_, _ = io.WriteString(w, "\n\n<section>\n")
	if eval := dot; len(eval) != 0 {
		for _, _Varpost := range eval {
			dot := _Varpost
			_ = dot
			_, _ = io.WriteString(w, "\n")
			if err := fun3(w, _Varpost); err != nil {
				return err
			}
			_, _ = io.WriteString(w, "\n")
		}
	}
	_, _ = io.WriteString(w, "\n</section>\n\n")
	if err := fun4(w, nil); err != nil {
		return err
	}
	_, _ = io.WriteString(w, "\n")
	return nil
}
