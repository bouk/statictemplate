package funcs

import (
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
}

//go:linkname And text/template.and
func And(arg0 interface{}, args ...interface{}) interface{}

//go:linkname Or text/template.or
func Or(arg0 interface{}, args ...interface{}) interface{}

//go:linkname Not text/template.not
func Not(arg interface{}) bool

//go:linkname Eq text/template.eq
func Eq(arg1 interface{}, arg2 ...interface{}) (bool, error)

//go:linkname Ne text/template.ne
func Ne(arg1, arg2 interface{}) (bool, error)

//go:linkname Lt text/template.lt
func Lt(arg1, arg2 interface{}) (bool, error)

//go:linkname Le text/template.le
func Le(arg1, arg2 interface{}) (bool, error)

//go:linkname Gt text/template.gt
func Gt(arg1, arg2 interface{}) (bool, error)

//go:linkname Ge text/template.ge
func Ge(arg1, arg2 interface{}) (bool, error)

//go:linkname Index text/template.index
func Index(item interface{}, indices ...interface{}) (interface{}, error)

//go:linkname Length text/template.length
func Length(item interface{}) (int, error)

//go:linkname Call text/template.call
func Call(fn interface{}, args ...interface{}) (interface{}, error)

//go:linkname Html text/template.HTMLEscaper
func Html(args ...interface{}) string

//go:linkname Js text/template.JSEscaper
func Js(args ...interface{}) string

//go:linkname Urlquery text/template.URLQueryEscaper
func Urlquery(args ...interface{}) string

//go:linkname Print fmt.Sprint
func Print(a ...interface{}) string

//go:linkname Printf fmt.Sprintf
func Printf(format string, a ...interface{}) string

//go:linkname Println fmt.Sprintln
func Println(a ...interface{}) string
