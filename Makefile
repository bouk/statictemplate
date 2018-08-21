build: example/template/template.go

test:
	go test ./...

example/template/template.go: example/template/*.tmpl
	statictemplate -html -o $@ -t "Index:index.tmpl:[]bou.ke/statictemplate/example.Post" $^

.PHONY: test build
