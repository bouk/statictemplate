package example_test

import (
	"bytes"
	"fmt"
	"github.com/bouk/statictemplate/example"
	staticTemplate "github.com/bouk/statictemplate/example/template"
	"testing"
	"text/template"
)

var testData []example.Post

func init() {
	for i := 0; i < 100; i++ {
		testData = append(testData, example.Post{
			Title: fmt.Sprintf("Post %d", i),
			Body:  "Very good post",
		})
	}
}

func BenchmarkStaticTemplate(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var b bytes.Buffer
		if err := staticTemplate.Index(&b, testData); err != nil {
			panic(err)
		}
	}
}

func BenchmarkDynamicTemplate(b *testing.B) {
	t := template.Must(template.ParseGlob("./template/*.tmpl"))
	for n := 0; n < b.N; n++ {
		var b bytes.Buffer
		if err := t.ExecuteTemplate(&b, "index.tmpl", testData); err != nil {
			panic(err)
		}
	}
}
