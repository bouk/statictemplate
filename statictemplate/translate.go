package statictemplate

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"io"
	"path"
	"text/template/parse"

	"github.com/bouk/statictemplate/internal"
	"golang.org/x/tools/go/types/typeutil"
)

var builtinFuncs map[string]*types.Func

func init() {
	var err error
	_, _, builtinFuncs, err = internal.ImportFuncMap("github.com/bouk/statictemplate/funcs.Funcs")
	if err != nil {
		panic(err)
	}
}

const varPrefix = "_Var"

type scope map[string]types.Type

// TranslateInstruction specifies a single function to be generated from a template
type TranslateInstruction struct {
	FunctionName string
	TemplateName string
	Dot          types.Type
}

// Translate is a convenience method for New(template).Translate(pkg, instructions)
func Translate(template interface{}, pkg string, instructions []TranslateInstruction) ([]byte, error) {
	translator := New(template)
	return translator.Translate(pkg, instructions)
}

// Translator converts a template with a set of instructions to Go code
type Translator struct {
	Funcs map[string]*types.Func

	scopes               []scope
	template             wrappedTemplate
	id                   int
	specializedFunctions map[wrappedTemplate]*typeutil.Map
	errorFunctions       *typeutil.Map
	generatedFunctions   []string
	imports              map[string]string
}

// New creates a new instance of Translator
func New(template interface{}) *Translator {
	wrapped := wrap(template)
	return &Translator{
		Funcs: map[string]*types.Func{},

		scopes: []scope{
			make(scope),
		},
		specializedFunctions: make(map[wrappedTemplate]*typeutil.Map),
		errorFunctions:       &typeutil.Map{},
		imports:              make(map[string]string),
		template:             wrapped,
	}
}

// Translate converts a template with a set of instructions to Go code
func (t *Translator) Translate(pkg string, instructions []TranslateInstruction) ([]byte, error) {
	var result []resultEntry

	for _, instruction := range instructions {
		temp, err := t.template.Lookup(instruction.TemplateName)
		if err != nil {
			return nil, err
		}
		functionName, err := t.generateTemplate(temp, instruction.Dot)
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

func (t *Translator) importPackage(name string) string {
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

func (t *Translator) generateFunctionName() string {
	name := fmt.Sprintf("fun%d", t.id)
	t.id++
	return name
}

func (t *Translator) pushScope() {
	t.scopes = append(t.scopes, make(scope))
}

func (t *Translator) popScope() {
	t.scopes = t.scopes[:len(t.scopes)-1]
}

// Checks whether identifier is in scope
func (t *Translator) inScope(name string) bool {
	_, ok := t.scopes[len(t.scopes)-1][name]
	return ok
}

// Checks whether identifier is in scope, or add it otherwise
func (t *Translator) addToScope(name string, typ types.Type) {
	t.scopes[len(t.scopes)-1][name] = typ
}

func (t *Translator) findVariable(name string) (types.Type, error) {
	for i := len(t.scopes) - 1; i >= 0; i-- {
		if typ, ok := t.scopes[i][name]; ok {
			return typ, nil
		}
	}
	return nil, fmt.Errorf("Can't find variable %s in scope", name)
}

type sortedTypes []types.Type

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

func (t *Translator) translateNode(w io.Writer, node parse.Node, dot types.Type) error {
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
				fmt.Fprintf(writer, "%s%s = ", varPrefix, ident)
			} else {
				fmt.Fprintf(writer, "%s%s := ", varPrefix, ident)
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
			if !t.inScope(ident) {
				fmt.Fprintf(writer, "\n_ = %s%s", varPrefix, ident)
			}
			t.addToScope(ident, typ)
		}

		if len(node.Pipe.Decl) == 0 {
			basic, ok := typ.(*types.Basic)
			if ok && basic.Kind() == types.String {
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

func typeIsNil(typ types.Type) bool {
	return typ == nil || types.Identical(typ, types.Typ[types.UntypedNil])
}

func writeTruthiness(w io.Writer, typ types.Type) error {
	if typeIsNil(typ) {
		_, err := io.WriteString(w, "eval != nil")
		return err
	}
	switch typ := typ.(type) {
	case *types.Array, *types.Map, *types.Slice:
		_, err := io.WriteString(w, "len(eval) != 0")
		return err
	case *types.Pointer, *types.Chan:
		_, err := io.WriteString(w, "eval != nil")
		return err
	case *types.Struct:
		_, err := io.WriteString(w, "true")
		return err
	case *types.Basic:
		info := typ.Info()
		if info&types.IsNumeric != 0 {
			_, err := io.WriteString(w, "eval != 0")
			return err
		} else if info&types.IsString != 0 {
			_, err := io.WriteString(w, "len(eval) != 0")
			return err
		} else if info&types.IsBoolean != 0 {
			_, err := io.WriteString(w, "eval")
			return err
		}
		return fmt.Errorf("Don't know how to evaluate %s", typ)
	default:
		return fmt.Errorf("Don't know how to evaluate %s", typ)
	}
}

func (t *Translator) generateTemplate(temp wrappedTemplate, typ types.Type) (string, error) {
	funcs, ok := t.specializedFunctions[temp]
	if !ok {
		funcs = &typeutil.Map{}
		t.specializedFunctions[temp] = funcs
	}
	functionName, ok := funcs.At(typ).(string)
	if !ok {
		functionName = t.generateFunctionName()
		funcs.Set(typ, functionName)

		var buf bytes.Buffer
		typeName := "interface{}"
		if !typeIsNil(typ) {
			typeName = t.typeName(typ)
		}

		fmt.Fprintf(&buf, "// %s(", temp.Name())
		if typeIsNil(typ) {
			buf.WriteString("nil")
		} else {
			buf.WriteString(typeName)
		}
		t.importPackage("io")
		fmt.Fprintf(&buf, ")\nfunc %s(w io.Writer, dot %s) error {\n", functionName, typeName)
		oldScopes := t.scopes
		t.scopes = []scope{make(scope)}
		if err := t.translateNode(&buf, temp.Tree().Root, typ); err != nil {
			return "", err
		}
		t.scopes = oldScopes
		buf.WriteString("return nil\n}\n")

		t.generatedFunctions = append(t.generatedFunctions, buf.String())
	}

	return functionName, nil
}

func (t *Translator) translateTemplate(w io.Writer, dot types.Type, node *parse.TemplateNode) error {
	var buf bytes.Buffer
	typ, err := t.translatePipe(&buf, dot, node.Pipe)
	if err != nil {
		return err
	}
	temp, err := t.template.Lookup(node.Name)
	if err != nil {
		return err
	}
	name, err := t.generateTemplate(temp, typ)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "if err := %s(w, ", name)
	buf.WriteTo(w)
	_, err = io.WriteString(w, "); err != nil {\nreturn err\n}\n")
	return err
}

func (t *Translator) translateScoped(w io.Writer, dot types.Type, nodeType parse.NodeType, pipe *parse.PipeNode, list, elseList *parse.ListNode) error {
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
		io.WriteString(w, "dot := eval\n_ = dot\n")
	}

	if nodeType == parse.NodeRange {
		var elem types.Type
		switch typ := typ.(type) {
		case *types.Chan:
			elem = typ.Elem()
		case *types.Slice:
			elem = typ.Elem()
		case *types.Array:
			elem = typ.Elem()
		default:
			return fmt.Errorf("range over non-iterable: %v", pipe.Pos)
		}

		switch len(pipe.Decl) {
		case 0:
			io.WriteString(w, "for _, dot := range eval {\n_ = dot\n")
		case 1:
			ident := pipe.Decl[0].Ident[0][1:]
			fmt.Fprintf(w, "for _, %s%s := range eval {\ndot := %s%s\n_ = dot\n", varPrefix, ident, varPrefix, ident)
			t.addToScope(ident, elem)
		case 2:
			index := pipe.Decl[0].Ident[0][1:]
			ident := pipe.Decl[1].Ident[0][1:]
			t.addToScope(index, types.Typ[types.Int64])
			t.addToScope(ident, elem)
			fmt.Fprintf(w, "for %s%s, %s%s := range eval {\n_ = %s%s\ndot := %s%s\n_ = dot\n", varPrefix, index, varPrefix, ident, varPrefix, index, varPrefix, ident)
		default:
			return fmt.Errorf("Too many declarations for range")
		}

		if err := t.translateNode(w, list, elem); err != nil {
			return err
		}

		io.WriteString(w, "}\n")
	} else {
		switch len(pipe.Decl) {
		case 0:
		case 1:
			ident := pipe.Decl[0].Ident[0][1:]
			fmt.Fprintf(w, "%s%s := eval\n_ = %s%s\n", varPrefix, ident, varPrefix, ident)
			t.addToScope(ident, typ)
		default:
			return fmt.Errorf("Too many declarations")
		}

		if err := t.translateNode(w, list, dot); err != nil {
			return err
		}
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

func (t *Translator) translatePipe(w io.Writer, dot types.Type, pipe *parse.PipeNode) (types.Type, error) {
	if pipe == nil {
		io.WriteString(w, "nil")
		return types.Typ[types.UntypedNil], nil
	} else {
		return t.translateCommand(w, dot, pipe.Cmds[len(pipe.Cmds)-1], pipe.Cmds[:len(pipe.Cmds)-1])
	}
}

func (t *Translator) translateCall(w io.Writer, dot types.Type, args []parse.Node, nextCommands []*parse.CommandNode) error {
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

func (t *Translator) translateCommand(w io.Writer, dot types.Type, cmd *parse.CommandNode, nextCommands []*parse.CommandNode) (types.Type, error) {
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
		return types.Typ[types.Bool], err
	case *parse.DotNode:
		_, err := io.WriteString(w, "dot")
		return dot, err
	case *parse.NilNode:
		return nil, fmt.Errorf("nil is not a command")
	case *parse.NumberNode:
		if action.IsInt {
			_, err := fmt.Fprint(w, action.Int64)
			return types.Typ[types.Int64], err
		} else if action.IsUint {
			_, err := fmt.Fprint(w, action.Uint64)
			return types.Typ[types.Uint64], err
		} else if action.IsFloat {
			_, err := fmt.Fprint(w, action.Float64)
			return types.Typ[types.Float64], err
		} else if action.IsComplex {
			_, err := fmt.Fprint(w, action.Complex128)
			return types.Typ[types.Complex128], err
		} else {
			return nil, fmt.Errorf("Unknown number node %v", action)
		}
	case *parse.StringNode:
		_, err := fmt.Fprintf(w, "%q", action.Text)
		return types.Typ[types.String], err
	default:
		return nil, fmt.Errorf("Unknown pipe node %s, %s", action.String(), action.Type())
	}
}

func (t *Translator) translateArg(w io.Writer, dot types.Type, arg parse.Node) (types.Type, error) {
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
		return types.Typ[types.Bool], err
	case *parse.DotNode:
		_, err := io.WriteString(w, "dot")
		return dot, err
	case *parse.NilNode:
		_, err := io.WriteString(w, "nil")
		return types.Typ[types.UntypedNil], err
	case *parse.NumberNode:
		if arg.IsInt {
			_, err := fmt.Fprint(w, arg.Int64)
			return types.Typ[types.Int64], err
		} else {
			return nil, fmt.Errorf("Unknown number node %v", arg)
		}
	case *parse.StringNode:
		_, err := fmt.Fprintf(w, "%q", arg.Text)
		return types.Typ[types.String], err
	default:
		return nil, fmt.Errorf("Unknown arg %s, %s", arg.String(), arg.Type())
	}
}

func (t *Translator) translateChain(w io.Writer, dot types.Type, node *parse.ChainNode, args []parse.Node, nextCommands []*parse.CommandNode) (types.Type, error) {
	var buf bytes.Buffer
	typ, err := t.translateArg(&buf, dot, node.Node)
	if err != nil {
		return nil, err
	}
	return t.translateFieldChain(w, dot, &buf, typ, node.Field, args, nextCommands)
}

func (t *Translator) translateVariable(w io.Writer, dot types.Type, node *parse.VariableNode, args []parse.Node, nextCommands []*parse.CommandNode) (types.Type, error) {
	ident := node.Ident[0][1:]
	if len(node.Ident) > 1 && (len(args) != 0 || len(nextCommands) != 0) {
		return nil, fmt.Errorf("Can't call variable %s", node.Ident[0])
	}
	typ, err := t.findVariable(ident)
	if err != nil {
		return nil, err
	}

	return t.translateFieldChain(w, dot, constantWriterTo(varPrefix+ident), typ, node.Ident[1:], args, nextCommands)
}

func (t *Translator) generateErrorFunction(typ types.Type) string {
	name, ok := t.errorFunctions.At(typ).(string)
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
		t.errorFunctions.Set(typ, name)
	}
	return name
}

func (t *Translator) getFunction(ident string) (*types.Signature, string, error) {
	if f, ok := t.Funcs[ident]; ok {
		pkgName := t.importPackage(f.Pkg().Path())
		return f.Type().(*types.Signature), fmt.Sprintf("%s.%s", pkgName, f.Name()), nil
	} else if f, ok := builtinFuncs[ident]; ok {
		pkgName := t.importPackage(f.Pkg().Path())
		return f.Type().(*types.Signature), fmt.Sprintf("%s.%s", pkgName, f.Name()), nil
	} else {
		return nil, "", fmt.Errorf("unknown function %s", ident)
	}
}

func (t *Translator) translateFunction(w io.Writer, dot types.Type, ident *parse.IdentifierNode, args []parse.Node, nextCommands []*parse.CommandNode) (types.Type, error) {
	typ, fName, err := t.getFunction(ident.Ident)
	if err != nil {
		return nil, err
	}

	numOut := typ.Results().Len()

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

	return typ.Results().At(0).Type(), nil
}

func (t *Translator) translateField(w io.Writer, dot types.Type, field *parse.FieldNode, args []parse.Node, nextCommands []*parse.CommandNode) (types.Type, error) {
	return t.translateFieldChain(w, dot, constantWriterTo("dot"), dot, field.Ident, args, nextCommands)
}

func (t *Translator) translateFieldChain(w io.Writer, dot types.Type, dotCode io.WriterTo, typ types.Type, fields []string, args []parse.Node, nextCommands []*parse.CommandNode) (types.Type, error) {
	var buf bytes.Buffer
	guards := []string{}
	for i, name := range fields {
		obj, _, _ := types.LookupFieldOrMethod(typ, true, nil, name)

		switch obj := obj.(type) {
		case *types.Func:
			sig := obj.Type().(*types.Signature)
			out := sig.Results()
			typ = out.At(0).Type()
			numOut := out.Len()
			if numOut == 2 {
				guards = append(guards, fmt.Sprintf("%s(", t.generateErrorFunction(typ)))
			} else if numOut != 1 {
				return nil, fmt.Errorf("Only support 1, 2 output variable %s.%s", t.typeName(typ), obj.Name)
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
		case *types.Var:
			fmt.Fprintf(&buf, ".%s", name)
			typ = obj.Type()
		default:
			return nil, fmt.Errorf("Unknown field %s for type %s", name, typ.String())
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

func (t *Translator) typeName(typ types.Type) string {
	switch obj := typ.(type) {
	case *types.Named:
		name := obj.Obj()
		return t.importPackage(name.Pkg().Path()) + "." + name.Name()
	case *types.Pointer:
		return fmt.Sprintf("*%s", t.typeName(obj.Elem()))
	case *types.Slice:
		return fmt.Sprintf("[]%s", t.typeName(obj.Elem()))
	case *types.Map:
		return fmt.Sprintf("map[%s]%s", t.typeName(obj.Key()), t.typeName(obj.Elem()))
	case *types.Chan:
		return fmt.Sprintf("chan %s", t.typeName(obj.Elem()))
	case *types.Array:
		return fmt.Sprintf("[%d]%s", obj.Len(), t.typeName(obj.Elem()))
	default:
		return typ.String()
	}
}
