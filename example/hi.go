package main

import (
	"github.com/bouk/statictemplate/example/template"
	"os"
)

func main() {
	template.Hi(os.Stdout, "Bouke")
}
