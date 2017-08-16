package funcs

import (
	"fmt"
	"text/template"
	_ "unsafe"
)

var Funcs = map[string]interface{}{
	"and":    And,
	"or":     Or,
	"not":    Not,
	"eq":     Eq,
	"ne":     Ne,
	"lt":     Lt,
	"le":     Le,
	"gt":     Gt,
	"ge":     Ge,
	"index":  Index,
	"length": Length,
	"call":   Call,

	"html":     Html,
	"js":       Js,
	"urlquery": Urlquery,
	"print":    Print,
	"printf":   Printf,
	"println":  Println,

	"_html_template_attrescaper":     Attrescaper,
	"_html_template_commentescaper":  Commentescaper,
	"_html_template_cssescaper":      Cssescaper,
	"_html_template_cssvaluefilter":  Cssvaluefilter,
	"_html_template_htmlnamefilter":  Htmlnamefilter,
	"_html_template_htmlescaper":     Htmlescaper,
	"_html_template_jsregexpescaper": Jsregexpescaper,
	"_html_template_jsstrescaper":    Jsstrescaper,
	"_html_template_jsvalescaper":    Jsvalescaper,
	"_html_template_nospaceescaper":  Htmlnospaceescaper,
	"_html_template_rcdataescaper":   Rcdataescaper,
	"_html_template_urlescaper":      Urlescaper,
	"_html_template_urlfilter":       Urlfilter,
	"_html_template_urlnormalizer":   Urlnormalizer,
}

//go:linkname and text/template.and
func and(arg0 interface{}, args ...interface{}) interface{}
func And(arg0 interface{}, args ...interface{}) interface{} {
	return and(arg0, args...)
}

//go:linkname or text/template.or
func or(arg0 interface{}, args ...interface{}) interface{}
func Or(arg0 interface{}, args ...interface{}) interface{} {
	return or(arg0, args...)
}

//go:linkname not text/template.not
func not(arg interface{}) bool
func Not(arg interface{}) bool {
	return not(arg)
}

//go:linkname eq text/template.eq
func eq(arg1 interface{}, arg2 ...interface{}) (bool, error)
func Eq(arg1 interface{}, arg2 ...interface{}) (bool, error) {
	return eq(arg1, arg2...)
}

//go:linkname ne text/template.ne
func ne(arg1, arg2 interface{}) (bool, error)
func Ne(arg1, arg2 interface{}) (bool, error) {
	return ne(arg1, arg2)
}

//go:linkname lt text/template.lt
func lt(arg1, arg2 interface{}) (bool, error)
func Lt(arg1, arg2 interface{}) (bool, error) {
	return lt(arg1, arg2)
}

//go:linkname le text/template.le
func le(arg1, arg2 interface{}) (bool, error)
func Le(arg1, arg2 interface{}) (bool, error) {
	return le(arg1, arg2)
}

//go:linkname gt text/template.gt
func gt(arg1, arg2 interface{}) (bool, error)
func Gt(arg1, arg2 interface{}) (bool, error) {
	return gt(arg1, arg2)
}

//go:linkname ge text/template.ge
func ge(arg1, arg2 interface{}) (bool, error)
func Ge(arg1, arg2 interface{}) (bool, error) {
	return ge(arg1, arg2)
}

//go:linkname index text/template.index
func index(item interface{}, indices ...interface{}) (interface{}, error)
func Index(item interface{}, indices ...interface{}) (interface{}, error) {
	return index(item, indices...)
}

//go:linkname length text/template.length
func length(item interface{}) (int, error)
func Length(item interface{}) (int, error) {
	return length(item)
}

//go:linkname call text/template.call
func call(fn interface{}, args ...interface{}) (interface{}, error)
func Call(fn interface{}, args ...interface{}) (interface{}, error) {
	return call(fn, args...)
}

func Html(args ...interface{}) string {
	return template.HTMLEscaper(args...)
}

func Js(args ...interface{}) string {
	return template.JSEscaper(args...)
}

func Urlquery(args ...interface{}) string {
	return template.URLQueryEscaper(args...)
}

func Print(a ...interface{}) string {
	return fmt.Sprint(a...)
}

func Printf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func Println(a ...interface{}) string {
	return fmt.Sprintln(a...)
}
