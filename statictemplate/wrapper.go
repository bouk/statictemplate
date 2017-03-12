package statictemplate

import (
	htmlTemplate "html/template"
	textTemplate "text/template"
	"text/template/parse"
	_ "unsafe"
)

type Template interface {
	Tree() *parse.Tree
	Lookup(name string) (Template, error)
	Name() string
}

type textTemplateWrapper struct {
	*textTemplate.Template
}

var _ Template = textTemplateWrapper{}

func (t textTemplateWrapper) Tree() *parse.Tree {
	return t.Template.Tree
}

func (t textTemplateWrapper) Lookup(name string) (Template, error) {
	return textTemplateWrapper{t.Template.Lookup(name)}, nil
}

type htmlTemplateWrapper struct {
	*htmlTemplate.Template
}

var _ Template = htmlTemplateWrapper{}

func (t htmlTemplateWrapper) Tree() *parse.Tree {
	return t.Template.Tree
}

func (t htmlTemplateWrapper) Lookup(name string) (Template, error) {
	temp, err := lookupAndEscapeTemplate(t.Template, name)
	return htmlTemplateWrapper{temp}, err
}

func wrap(template interface{}) Template {
	switch template := template.(type) {
	case *htmlTemplate.Template:
		return htmlTemplateWrapper{template}
	case *textTemplate.Template:
		return textTemplateWrapper{template}
	default:
		panic("invalid template passed in")
	}
}

//go:linkname lookupAndEscapeTemplate html/template.(*Template).lookupAndEscapeTemplate
func lookupAndEscapeTemplate(t *htmlTemplate.Template, name string) (*htmlTemplate.Template, error)
