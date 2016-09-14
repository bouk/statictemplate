example/template/template.go: example/template/*.tmpl
	statictemplate -o $@ -t "Hi:hi.tmpl:string" -t "Hello:hi.tmpl:*text/template.Template" $^
