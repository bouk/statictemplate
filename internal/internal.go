package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strconv"

	"golang.org/x/tools/go/loader"
)

var valueReferenceRe = regexp.MustCompile(`^(?:(.+)\.)?([A-Za-z][A-Za-z0-9]*)$`)

func ImportFuncMap(funcMap string) (string, string, map[string]*types.Func, error) {
	if funcMap == "" {
		return "", "", nil, nil
	}
	var funcMapImport, funcMapName string
	values := valueReferenceRe.FindStringSubmatch(funcMap)
	if values == nil || values[1] == "" {
		return "", "", nil, fmt.Errorf("invalid funcs value %q, expected <import>.<name>", funcMap)
	}
	funcMapImport = values[1]
	funcMapName = values[2]

	var conf loader.Config
	conf.Import(funcMapImport)
	prog, err := conf.Load()
	if err != nil {
		return "", "", nil, err
	}
	pack := prog.Package(funcMapImport)
	var obj *ast.Object
	for _, f := range pack.Files {
		if obj = f.Scope.Lookup(funcMapName); obj != nil {
			break
		}
	}
	if obj == nil {
		return "", "", nil, fmt.Errorf("Can't find function map %q", funcMap)
	}

	funcs := make(map[string]*types.Func)
	for _, el := range obj.Decl.(*ast.ValueSpec).Values[0].(*ast.CompositeLit).Elts {
		ex, ok := el.(*ast.KeyValueExpr)
		if !ok {
			return "", "", nil, fmt.Errorf("invalid function map format")
		}
		lit, ok := ex.Key.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return "", "", nil, fmt.Errorf("invalid function map format")
		}
		name, err := strconv.Unquote(lit.Value)
		if err != nil {
			return "", "", nil, fmt.Errorf("invalid function map format: %v", err)
		}
		ident, ok := ex.Value.(*ast.Ident)
		if !ok {
			return "", "", nil, fmt.Errorf("invalid function map format")
		}
		if f, ok := pack.Pkg.Scope().Lookup(ident.Name).(*types.Func); ok {
			funcs[name] = f
		} else {
			return "", "", nil, fmt.Errorf("invalid function map format")
		}
	}
	return funcMapImport, funcMapName, funcs, nil
}
