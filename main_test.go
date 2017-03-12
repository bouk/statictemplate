package main

import (
	"gopkg.in/stretchr/testify.v1/assert"
	"testing"
)

func TestParseCompilationTargets(t *testing.T) {
	var ct compilationTargets
	assert.NoError(t, ct.Set("Hi:hi.tmpl:string"))
	assert.NoError(t, ct.Set("Hello:hi.tmpl:*text/template.Template"))
	assert.NoError(t, ct.Set("Cool:hi.tmpl:text/template.Template"))
	assert.NoError(t, ct.Set("Neat:hi.tmpl:*github.com/bouk/whatever.Template"))
	expected := compilationTargets{
		compilationTarget{
			functionName: "Hi",
			templateName: "hi.tmpl",
			dot: dotType{
				packagePath: "",
				typeName:    "string",
				prefix:      "",
			},
		},
		compilationTarget{
			functionName: "Hello",
			templateName: "hi.tmpl",
			dot: dotType{
				packagePath: "text/template",
				typeName:    "Template",
				prefix:      "*",
			},
		},
		compilationTarget{
			functionName: "Cool",
			templateName: "hi.tmpl",
			dot: dotType{
				packagePath: "text/template",
				typeName:    "Template",
				prefix:      "",
			},
		},
		compilationTarget{
			functionName: "Neat",
			templateName: "hi.tmpl",
			dot: dotType{
				packagePath: "github.com/bouk/whatever",
				typeName:    "Template",
				prefix:      "*",
			},
		},
	}
	assert.Equal(t, expected, ct)
}

func TestParseCompilationTargetsError(t *testing.T) {
	var ct compilationTargets
	assert.Error(t, ct.Set("lol whatever man"), `expect compilation target in functionName:templateName:typeName format, got "lol whatever man"`)
}
