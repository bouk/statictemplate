package funcs

import (
	_ "html/template"
	_ "unsafe"
)

//go:linkname attrescaper html/template.attrEscaper
func attrescaper(args ...interface{}) string
func Attrescaper(args ...interface{}) string {
	return attrescaper(args...)
}

//go:linkname commentescaper html/template.commentEscaper
func commentescaper(args ...interface{}) string
func Commentescaper(args ...interface{}) string {
	return commentescaper(args...)
}

//go:linkname cssescaper html/template.cssEscaper
func cssescaper(args ...interface{}) string
func Cssescaper(args ...interface{}) string {
	return cssescaper(args...)
}

//go:linkname cssvaluefilter html/template.cssValueFilter
func cssvaluefilter(args ...interface{}) string
func Cssvaluefilter(args ...interface{}) string {
	return cssvaluefilter(args...)
}

//go:linkname htmlnamefilter html/template.htmlNameFilter
func htmlnamefilter(args ...interface{}) string
func Htmlnamefilter(args ...interface{}) string {
	return htmlnamefilter(args...)
}

//go:linkname htmlescaper html/template.htmlEscaper
func htmlescaper(args ...interface{}) string
func Htmlescaper(args ...interface{}) string {
	return htmlescaper(args...)
}

//go:linkname jsregexpescaper html/template.jsRegexpEscaper
func jsregexpescaper(args ...interface{}) string
func Jsregexpescaper(args ...interface{}) string {
	return jsregexpescaper(args...)
}

//go:linkname jsstrescaper html/template.jsStrEscaper
func jsstrescaper(args ...interface{}) string
func Jsstrescaper(args ...interface{}) string {
	return jsstrescaper(args...)
}

//go:linkname jsvalescaper html/template.jsValEscaper
func jsvalescaper(args ...interface{}) string
func Jsvalescaper(args ...interface{}) string {
	return jsvalescaper(args...)
}

//go:linkname htmlnospaceescaper html/template.htmlNospaceEscaper
func htmlnospaceescaper(args ...interface{}) string
func Htmlnospaceescaper(args ...interface{}) string {
	return htmlnospaceescaper(args...)
}

//go:linkname rcdataescaper html/template.rcdataEscaper
func rcdataescaper(args ...interface{}) string
func Rcdataescaper(args ...interface{}) string {
	return rcdataescaper(args...)
}

//go:linkname urlescaper html/template.urlEscaper
func urlescaper(args ...interface{}) string
func Urlescaper(args ...interface{}) string {
	return urlescaper(args...)
}

//go:linkname urlfilter html/template.urlFilter
func urlfilter(args ...interface{}) string
func Urlfilter(args ...interface{}) string {
	return urlfilter(args...)
}

//go:linkname urlnormalizer html/template.urlNormalizer
func urlnormalizer(args ...interface{}) string
func Urlnormalizer(args ...interface{}) string {
	return urlnormalizer(args...)
}
