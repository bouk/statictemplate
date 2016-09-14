package statictemplate

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"path"
	"reflect"
	"runtime"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/bouk/statictemplate/funcs"
)

const VarPrefix = "_Var"

type scope map[string]reflect.Type
type TranslateInstruction struct {
	FunctionName string
	TemplateName string
	Dot          reflect.Type
}

func Translate(temp *template.Template, pkg string, instructions []TranslateInstruction) ([]byte, error) {
	translator := &translator{
		funcs: map[string]interface{}{},
		scopes: []scope{
			make(scope),
		},
		specializedFunctions: make(map[*template.Template]map[reflect.Type]string),
		errorFunctions:       make(map[reflect.Type]string),
		imports:              make(map[string]string),
		template:             temp,
	}
	return translator.translate(pkg, instructions)
}

type translator struct {
	funcs                template.FuncMap
	scopes               []scope
	template             *template.Template
	id                   int
	specializedFunctions map[*template.Template]map[reflect.Type]string
	errorFunctions       map[reflect.Type]string
	generatedFunctions   []string
	imports              map[string]string
}

func (t *translator) importPackage(name string) string {
	if pkg, ok := t.imports[name]; ok {
		return pkg
	}

	var pkg string
	switch name {
	case "fmt", "io":
		pkg = name
	case "text/template":
		pkg = "template"
	case "github.com/bouk/statictemplate/funcs":
		pkg = "funcs"
	default:
		pkg = fmt.Sprintf("pkg%d", t.id)
		t.id++
	}

	t.imports[name] = pkg
	return pkg
}

func (t *translator) generateFunctionName() string {
	name := fmt.Sprintf("fun%d", t.id)
	t.id++
	return name
}

func (t *translator) pushScope() {
	t.scopes = append(t.scopes, make(scope))
}

func (t *translator) popScope() {
	t.scopes = t.scopes[:len(t.scopes)-1]
}

// Checks whether identifier is in scope
func (t *translator) inScope(name string) bool {
	_, ok := t.scopes[len(t.scopes)-1][name]
	return ok
}

// Checks whether identifier is in scope, or add it otherwise
func (t *translator) addToScope(name string, typ reflect.Type) {
	t.scopes[len(t.scopes)-1][name] = typ
}

func (t *translator) findVariable(name string) (reflect.Type, error) {
	for i := len(t.scopes) - 1; i >= 0; i-- {
		if typ, ok := t.scopes[i][name]; ok {
			return typ, nil
		}
	}
	return nil, fmt.Errorf("Can't find variable %s in scope", name)
}

type sortedTypes []reflect.Type

func (a sortedTypes) Len() int      { return len(a) }
func (a sortedTypes) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortedTypes) Less(i, j int) bool {
	if a[i] == nil {
		return true
	} else if a[j] == nil {
		return false
	} else {
		return a[i].String() < a[j].String()
	}
}

type resultEntry struct {
	name, typeName, functionName string
}

func (t *translator) translate(pkg string, instructions []TranslateInstruction) ([]byte, error) {
	var result []resultEntry

	for _, instruction := range instructions {
		functionName, err := t.generateTemplate(t.template.Lookup(instruction.TemplateName), instruction.Dot)
		if err != nil {
			return nil, err
		}
		result = append(result, resultEntry{
			name:         instruction.FunctionName,
			typeName:     t.typeName(instruction.Dot),
			functionName: functionName,
		})
	}

	t.importPackage("io")

	var buf bytes.Buffer

	fmt.Fprintf(&buf, `package %s
import (
`, pkg)
	for pkgPath, alias := range t.imports {
		if path.Base(pkgPath) == alias {
			fmt.Fprintf(&buf, "%q\n", pkgPath)
		} else {
			fmt.Fprintf(&buf, "%s %q\n", alias, pkgPath)
		}
	}
	io.WriteString(&buf, ")")

	for _, entry := range result {
		fmt.Fprintf(&buf, `
func %s(w io.Writer, dot %s) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			var ok bool
			if err, ok = recovered.(error); !ok {
				panic(recovered)
			}
		}
	}()
	return %s(w, dot)
}
`, entry.name, entry.typeName, entry.functionName)
	}

	for _, code := range t.generatedFunctions {
		io.WriteString(&buf, "\n")
		io.WriteString(&buf, code)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s: %v", buf.String(), err)
	}
	return formatted, nil
}

func (t *translator) translateNode(w io.Writer, node parse.Node, dot reflect.Type) error {
	switch node := node.(type) {
	case *parse.ListNode:
		for _, item := range node.Nodes {
			if err := t.translateNode(w, item, dot); err != nil {
				return err
			}
		}
		return nil
	case *parse.TextNode:
		t.importPackage("io")
		_, err := fmt.Fprintf(w, "_, _ = io.WriteString(w, %q)\n", node.Text)
		return err
	case *parse.ActionNode:
		pipe := node.Pipe
		writer := w
		if len(pipe.Decl) == 0 {
			writer = new(bytes.Buffer)
		} else if len(pipe.Decl) == 1 {
			ident := pipe.Decl[0].Ident[0][1:]
			if t.inScope(ident) {
				fmt.Fprintf(writer, "%s%s = ", VarPrefix, ident)
			} else {
				fmt.Fprintf(writer, "%s%s := ", VarPrefix, ident)
			}
		} else {
			return fmt.Errorf("Only support single variable for assignment")
		}

		typ, err := t.translatePipe(writer, dot, pipe)
		if err != nil {
			return err
		}
		if len(pipe.Decl) == 1 {
			ident := pipe.Decl[0].Ident[0][1:]
			t.addToScope(ident, typ)
		}

		if len(node.Pipe.Decl) == 0 {
			if typ == reflect.TypeOf("") {
				t.importPackage("io")
				io.WriteString(w, "_, _ = io.WriteString(w, ")
			} else {
				t.importPackage("fmt")
				io.WriteString(w, "_, _ = fmt.Fprint(w, ")
			}
			writer.(*bytes.Buffer).WriteTo(w)
			io.WriteString(w, ")")
		}
		_, err = io.WriteString(w, "\n")

		return err
	case *parse.WithNode:
		return t.translateScoped(w, dot, node.Type(), node.Pipe, node.List, node.ElseList)
	case *parse.IfNode:
		return t.translateScoped(w, dot, node.Type(), node.Pipe, node.List, node.ElseList)
	case *parse.RangeNode:
		return t.translateScoped(w, dot, node.Type(), node.Pipe, node.List, node.ElseList)
	case *parse.TemplateNode:
		return t.translateTemplate(w, dot, node)
	default:
		return fmt.Errorf("Unknown Node %s", node.Type())
	}
}

func writeTruthiness(w io.Writer, typ reflect.Type) error {
	if typ == nil {
		_, err := io.WriteString(w, "eval != nil")
		return err
	}
	switch typ.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		_, err := io.WriteString(w, "len(eval) != 0")
		return err
	case reflect.Bool:
		_, err := io.WriteString(w, "eval")
		return err
	case reflect.Ptr, reflect.Chan:
		_, err := io.WriteString(w, "eval != nil")
		return err
	case reflect.Struct:
		_, err := io.WriteString(w, "true")
		return err
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Float32, reflect.Float64:
		_, err := io.WriteString(w, "eval != 0")
		return err
	default:
		return fmt.Errorf("Don't know how to evaluate %s", typ)
	}
}

func (t *translator) generateTemplate(temp *template.Template, typ reflect.Type) (string, error) {
	funcs, ok := t.specializedFunctions[temp]
	if !ok {
		funcs = make(map[reflect.Type]string)
		t.specializedFunctions[temp] = funcs
	}
	functionName, ok := funcs[typ]
	if !ok {
		functionName = t.generateFunctionName()
		funcs[typ] = functionName

		var buf bytes.Buffer
		typeName := "interface{}"
		if typ != nil {
			typeName = t.typeName(typ)
		}

		fmt.Fprintf(&buf, "// %s(", temp.Name())
		if typ == nil {
			buf.WriteString("nil")
		} else {
			buf.WriteString(typeName)
		}
		t.importPackage("io")
		fmt.Fprintf(&buf, ")\nfunc %s(w io.Writer, dot %s) error {\n", functionName, typeName)
		oldScopes := t.scopes
		t.scopes = []scope{make(scope)}
		if err := t.translateNode(&buf, temp.Root, typ); err != nil {
			return "", err
		}
		t.scopes = oldScopes
		buf.WriteString("return nil\n}\n")

		t.generatedFunctions = append(t.generatedFunctions, buf.String())
	}

	return functionName, nil
}

func (t *translator) translateTemplate(w io.Writer, dot reflect.Type, node *parse.TemplateNode) error {
	var buf bytes.Buffer
	typ, err := t.translatePipe(&buf, dot, node.Pipe)
	if err != nil {
		return err
	}
	name, err := t.generateTemplate(t.template.Lookup(node.Name), typ)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "if err := %s(w, ", name)
	buf.WriteTo(w)
	_, err = io.WriteString(w, "); err != nil {\nreturn err\n}\n")
	return err
}

func (t *translator) translateScoped(w io.Writer, dot reflect.Type, nodeType parse.NodeType, pipe *parse.PipeNode, list, elseList *parse.ListNode) error {
	io.WriteString(w, "if eval := ")
	typ, err := t.translatePipe(w, dot, pipe)
	if err != nil {
		return err
	}
	io.WriteString(w, "; ")
	if err := writeTruthiness(w, typ); err != nil {
		return err
	}
	io.WriteString(w, "{\n")
	t.pushScope()

	if nodeType == parse.NodeWith {
		io.WriteString(w, "dot := eval\n")
	}

	if nodeType == parse.NodeRange {
		switch len(pipe.Decl) {
		case 0:
			io.WriteString(w, "for range eval {\n")
		case 1:
			ident := pipe.Decl[0].Ident[0][1:]
			fmt.Fprintf(w, "for _, %s%s := range eval {\n", VarPrefix, ident)
			t.addToScope(ident, typ.Elem())
		case 2:
			index := pipe.Decl[0].Ident[0][1:]
			ident := pipe.Decl[1].Ident[0][1:]
			t.addToScope(index, reflect.TypeOf(int64(0)))
			t.addToScope(ident, typ.Elem())
			fmt.Fprintf(w, "for %s%s, %s%s := range eval {\n", VarPrefix, index, VarPrefix, ident)
		default:
			return fmt.Errorf("Too many declarations for range")
		}
	} else {
		switch len(pipe.Decl) {
		case 0:
		case 1:
			ident := pipe.Decl[0].Ident[0][1:]
			fmt.Fprintf(w, "%s%s := eval\n", VarPrefix, ident)
			t.addToScope(ident, typ)
		default:
			return fmt.Errorf("Too many declarations")
		}
	}

	if err := t.translateNode(w, list, dot); err != nil {
		return err
	}

	if nodeType == parse.NodeRange {
		io.WriteString(w, "}\n")
	}

	t.popScope()
	io.WriteString(w, "}")
	if elseList != nil {
		io.WriteString(w, " else {\n")
		if err := t.translateNode(w, elseList, dot); err != nil {
			return err
		}
		io.WriteString(w, "}")
	}
	io.WriteString(w, "\n")
	return nil
}

func (t *translator) translatePipe(w io.Writer, dot reflect.Type, pipe *parse.PipeNode) (reflect.Type, error) {
	if pipe == nil {
		io.WriteString(w, "nil")
		return nil, nil
	} else {
		return t.translateCommand(w, dot, pipe.Cmds[len(pipe.Cmds)-1], pipe.Cmds[:len(pipe.Cmds)-1])
	}
}

func (t *translator) translateCall(w io.Writer, dot reflect.Type, args []parse.Node, nextCommands []*parse.CommandNode) error {
	io.WriteString(w, "(")
	for i, arg := range args {
		if i != 0 {
			io.WriteString(w, ", ")
		}
		if _, err := t.translateArg(w, dot, arg); err != nil {
			return err
		}
	}
	if len(nextCommands) != 0 {
		if len(args) != 0 {
			io.WriteString(w, ", ")
		}
		if _, err := t.translateCommand(w, dot, nextCommands[len(nextCommands)-1], nextCommands[:len(nextCommands)-1]); err != nil {
			return err
		}
	}
	io.WriteString(w, ")")
	return nil
}

func (t *translator) translateCommand(w io.Writer, dot reflect.Type, cmd *parse.CommandNode, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	action := cmd.Args[0]
	args := cmd.Args[1:]

	switch action := action.(type) {
	case *parse.FieldNode:
		return t.translateField(w, dot, action, args, nextCommands)
	case *parse.ChainNode:
		return t.translateChain(w, dot, action, args, nextCommands)
	case *parse.IdentifierNode:
		return t.translateFunction(w, dot, action, args, nextCommands)
	case *parse.PipeNode:
		// We ignore args, nextCommands in pipes
		return t.translatePipe(w, dot, action)
	case *parse.VariableNode:
		return t.translateVariable(w, dot, action, args, nextCommands)
	}

	if len(args) > 0 || len(nextCommands) > 0 {
		return nil, fmt.Errorf("Dunno what to do with args %v %s %v", cmd.Args, action.Type(), nextCommands)
	}

	switch action := action.(type) {
	case *parse.BoolNode:
		_, err := fmt.Fprint(w, action.True)
		return reflect.TypeOf(true), err
	case *parse.DotNode:
		_, err := io.WriteString(w, "dot")
		return dot, err
	case *parse.NilNode:
		return nil, fmt.Errorf("nil is not a command")
	case *parse.NumberNode:
		if action.IsInt {
			_, err := fmt.Fprint(w, action.Int64)
			return reflect.TypeOf(int64(0)), err
		} else {
			return nil, fmt.Errorf("Unknown number node %v", action)
		}
	case *parse.StringNode:
		_, err := fmt.Fprintf(w, "%q", action.Text)
		return reflect.TypeOf(""), err
	default:
		return nil, fmt.Errorf("Unknown pipe node %s, %s", action.String(), action.Type())
	}
}

func (t *translator) translateArg(w io.Writer, dot reflect.Type, arg parse.Node) (reflect.Type, error) {
	switch arg := arg.(type) {
	case *parse.FieldNode:
		return t.translateField(w, dot, arg, nil, nil)
	case *parse.ChainNode:
		return t.translateChain(w, dot, arg, nil, nil)
	case *parse.IdentifierNode:
		return t.translateFunction(w, dot, arg, nil, nil)
	case *parse.PipeNode:
		if len(arg.Decl) > 0 {
			// TODO(bouk): do (is it even possible?)
			return nil, fmt.Errorf("Can't process inline variable assignment right now")
		}
		return t.translatePipe(w, dot, arg)
	case *parse.VariableNode:
		return t.translateVariable(w, dot, arg, nil, nil)
	case *parse.BoolNode:
		_, err := fmt.Fprint(w, arg.True)
		return reflect.TypeOf(true), err
	case *parse.DotNode:
		_, err := io.WriteString(w, "dot")
		return dot, err
	case *parse.NilNode:
		_, err := io.WriteString(w, "nil")
		return reflect.TypeOf(nil), err
	case *parse.NumberNode:
		if arg.IsInt {
			_, err := fmt.Fprint(w, arg.Int64)
			return reflect.TypeOf(int64(0)), err
		} else {
			return nil, fmt.Errorf("Unknown number node %v", arg)
		}
	case *parse.StringNode:
		_, err := fmt.Fprintf(w, "%q", arg.Text)
		return reflect.TypeOf(""), err
	default:
		return nil, fmt.Errorf("Unknown arg %s, %s", arg.String(), arg.Type())
	}
}

func (t *translator) translateChain(w io.Writer, dot reflect.Type, node *parse.ChainNode, args []parse.Node, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	var buf bytes.Buffer
	typ, err := t.translateArg(&buf, dot, node.Node)
	if err != nil {
		return nil, err
	}
	return t.translateFieldChain(w, dot, &buf, typ, node.Field, args, nextCommands)
}

func (t *translator) translateVariable(w io.Writer, dot reflect.Type, node *parse.VariableNode, args []parse.Node, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	ident := node.Ident[0][1:]
	if len(node.Ident) > 1 && (len(args) != 0 || len(nextCommands) != 0) {
		return nil, fmt.Errorf("Can't call variable %s", node.Ident[0])
	}
	typ, err := t.findVariable(ident)
	if err != nil {
		return nil, err
	}

	return t.translateFieldChain(w, dot, constantWriterTo(VarPrefix+ident), typ, node.Ident[1:], args, nextCommands)
}

func (t *translator) generateErrorFunction(typ reflect.Type) string {
	name, ok := t.errorFunctions[typ]
	if !ok {
		name = t.generateFunctionName()
		typeName := t.typeName(typ)

		t.generatedFunctions = append(t.generatedFunctions, fmt.Sprintf(`
func %s(value %s, err error) %s {
	if err != nil {
		panic(err)
	}
	return value
}`, name, typeName, typeName))
	}
	return name
}

func (t *translator) getFunction(ident string) (reflect.Type, string, error) {
	if f, ok := t.funcs[ident]; ok {
		fName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		strs := strings.SplitN(fName, ".", 2)
		pkgName := t.importPackage(strs[0])
		return reflect.TypeOf(f), fmt.Sprintf("%s.%s", pkgName, strs[1]), nil
	} else if fType, ok := funcs.Funcs[ident]; ok {
		packageName := t.importPackage("github.com/bouk/statictemplate/funcs")
		return reflect.TypeOf(fType), fmt.Sprintf("%s.%s", packageName, strings.Title(ident)), nil
	} else {
		return nil, "", fmt.Errorf("Unknown function %s", ident)
	}
}

func (t *translator) translateFunction(w io.Writer, dot reflect.Type, ident *parse.IdentifierNode, args []parse.Node, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	typ, fName, err := t.getFunction(ident.Ident)
	if err != nil {
		return nil, err
	}

	numOut := typ.NumOut()

	if numOut == 2 {
		fmt.Fprintf(w, "%s(", t.generateErrorFunction(typ))
	} else if numOut != 1 {
		return nil, fmt.Errorf("Only support 1, 2 output variable %s", ident.Ident)
	}

	io.WriteString(w, fName)

	if err := t.translateCall(w, dot, args, nextCommands); err != nil {
		return nil, err
	}

	if numOut == 2 {
		io.WriteString(w, ")")
	}

	return typ.Out(0), nil
}

func (t *translator) translateField(w io.Writer, dot reflect.Type, field *parse.FieldNode, args []parse.Node, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	return t.translateFieldChain(w, dot, constantWriterTo("dot"), dot, field.Ident, args, nextCommands)
}

func (t *translator) translateFieldChain(w io.Writer, dot reflect.Type, dotCode io.WriterTo, typ reflect.Type, fields []string, args []parse.Node, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	var buf bytes.Buffer
	guards := []string{}
	for i, name := range fields {
		if method, ok := typ.MethodByName(name); ok {
			typ = method.Type.Out(0)
			numOut := method.Type.NumOut()
			if numOut == 2 {
				guards = append(guards, fmt.Sprintf("%s(", t.generateErrorFunction(typ)))
			} else if numOut != 1 {
				return nil, fmt.Errorf("Only support 1, 2 output variable %s.%s", t.typeName(typ), method.Name)
			}
			fmt.Fprintf(&buf, ".%s", name)

			var err error
			if i == len(fields)-1 {
				err = t.translateCall(&buf, dot, args, nextCommands)
			} else {
				err = t.translateCall(&buf, dot, nil, nil)
			}
			if err != nil {
				return nil, err
			}
			if numOut == 2 {
				io.WriteString(&buf, ")")
			}
		} else if field, ok := typ.FieldByName(name); ok {
			fmt.Fprintf(&buf, ".%s", name)
			typ = field.Type
		} else {
			return nil, fmt.Errorf("Unknown field %s for type %s", name, typ.Name())
		}
	}
	for i := len(guards) - 1; i >= 0; i-- {
		io.WriteString(w, guards[i])
	}
	_, err := dotCode.WriteTo(w)
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteTo(w)
	return typ, err
}

func (t *translator) typeName(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Ptr:
		return fmt.Sprintf("*%s", t.typeName(typ.Elem()))
	case reflect.Slice:
		return fmt.Sprintf("[]%s", t.typeName(typ.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", t.typeName(typ.Key()), t.typeName(typ.Elem()))
	case reflect.Chan:
		return fmt.Sprintf("chan %s", t.typeName(typ.Elem()))
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", typ.Len(), t.typeName(typ.Elem()))
	}
	pkg := typ.PkgPath()
	if pkg != "" {
		pkg = t.importPackage(pkg) + "."
	}

	name := typ.Name()
	if name == "" {
		name = typ.String()
	}
	return pkg + name
}
