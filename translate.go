package statictemplate

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"reflect"
	"runtime"
	"text/template"
	"text/template/parse"
)

const VarPrefix = "_Var"

type scope map[string]reflect.Type

func Translate(name, text string, dot reflect.Type) (string, error) {
	return (&translator{
		funcs: map[string]interface{}{
			"print":  fmt.Sprint,
			"printf": fmt.Sprintf,
		},
		scopes: []scope{
			make(scope),
		},
	}).translate(name, text, dot)
}

type translator struct {
	funcs  template.FuncMap
	scopes []scope
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

func (t *translator) translate(name, text string, dot reflect.Type) (string, error) {
	trees, err := parse.Parse(name, text, "", "", t.funcs)
	if err != nil {
		return "", err
	}
	tree := trees[name]

	typeName := dot.Name()
	if typeName == "" {
		typeName = dot.String()
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "func %s(w io.Writer, dot %s) error {\n", name, typeName)
	if err := t.translateNode(tree.Root, &buf, dot); err != nil {
		return "", err
	}
	fmt.Fprintf(&buf, "return nil\n}")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("%s: %v", buf.String(), err)
	}
	return string(formatted), nil
}

func (t *translator) translateNode(node parse.Node, w io.Writer, dot reflect.Type) error {
	switch node := node.(type) {
	case *parse.ListNode:
		for _, item := range node.Nodes {
			if err := t.translateNode(item, w, dot); err != nil {
				return err
			}
		}
		return nil
	case *parse.TextNode:
		_, err := fmt.Fprintf(w, "_, _ = io.WriteString(w, %q)\n", node.Text)
		return err
	case *parse.ActionNode:
		pipe := node.Pipe
		if len(pipe.Decl) == 0 {
			io.WriteString(w, "_, _ = fmt.Fprint(w, ")
		} else if len(pipe.Decl) == 1 {
			ident := pipe.Decl[0].Ident[0][1:]
			if t.inScope(ident) {
				fmt.Fprintf(w, "%s%s = ", VarPrefix, ident)
			} else {
				fmt.Fprintf(w, "%s%s := ", VarPrefix, ident)
			}
		} else {
			return fmt.Errorf("Only support single variable for assignment")
		}

		typ, err := t.translatePipe(w, dot, pipe)
		if err != nil {
			return err
		}
		if len(pipe.Decl) == 1 {
			ident := pipe.Decl[0].Ident[0][1:]
			t.addToScope(ident, typ)
		}

		if len(node.Pipe.Decl) == 0 {
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
	default:
		return fmt.Errorf("Unknown Node %s", node.Type())
	}
}

func writeTruthiness(w io.Writer, kind reflect.Kind) error {
	switch kind {
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
		return fmt.Errorf("Don't know how to evaluate %s", kind)
	}
}

func (t *translator) translateScoped(w io.Writer, dot reflect.Type, nodeType parse.NodeType, pipe *parse.PipeNode, list, elseList *parse.ListNode) error {
	io.WriteString(w, "if eval := ")
	typ, err := t.translatePipe(w, dot, pipe)
	if err != nil {
		return err
	}
	io.WriteString(w, "; ")
	if err := writeTruthiness(w, typ.Kind()); err != nil {
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

	if err := t.translateNode(list, w, dot); err != nil {
		return err
	}

	if nodeType == parse.NodeRange {
		io.WriteString(w, "}\n")
	}

	t.popScope()
	io.WriteString(w, "}")
	if elseList != nil {
		io.WriteString(w, " else {\n")
		if err := t.translateNode(elseList, w, dot); err != nil {
			return err
		}
		io.WriteString(w, "}")
	}
	io.WriteString(w, "\n")
	return nil
}

func (t *translator) translatePipe(w io.Writer, dot reflect.Type, pipe *parse.PipeNode) (reflect.Type, error) {
	return t.translateCommand(w, dot, pipe.Cmds[len(pipe.Cmds)-1], pipe.Cmds[:len(pipe.Cmds)-1])
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
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
			// TODO(bouk): do
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
	typ, err := t.translateArg(w, dot, node.Node)
	if err != nil {
		return nil, err
	}
	return t.translateFieldChain(w, dot, typ, node.Field, args, nextCommands)
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
	fmt.Fprintf(w, "%s%s", VarPrefix, ident)
	return t.translateFieldChain(w, dot, typ, node.Ident[1:], args, nextCommands)
}

func (t *translator) translateFunction(w io.Writer, dot reflect.Type, ident *parse.IdentifierNode, args []parse.Node, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	f := t.funcs[ident.Ident]
	_, err := fmt.Fprint(w, GetFunctionName(f))
	if err != nil {
		return nil, err
	}
	typ := reflect.TypeOf(f)
	// TODO(bouk): support err return value
	if typ.NumOut() != 1 {
		return nil, fmt.Errorf("Only support 1 output variable %s", GetFunctionName(f))
	}

	err = t.translateCall(w, dot, args, nextCommands)
	return typ.Out(0), err
}

func (t *translator) translateField(w io.Writer, dot reflect.Type, field *parse.FieldNode, args []parse.Node, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	io.WriteString(w, "dot")
	return t.translateFieldChain(w, dot, dot, field.Ident, args, nextCommands)
}

func (t *translator) translateFieldChain(w io.Writer, dot reflect.Type, typ reflect.Type, fields []string, args []parse.Node, nextCommands []*parse.CommandNode) (reflect.Type, error) {
	for i, name := range fields {
		if method, ok := typ.MethodByName(name); ok {
			// TODO(bouk): support second err out
			if method.Type.NumOut() != 1 {
				return nil, fmt.Errorf("Only support single output argument %s", method.Name)
			}
			fmt.Fprintf(w, ".%s", name)

			var err error
			if i == len(fields)-1 {
				err = t.translateCall(w, dot, args, nextCommands)
			} else {
				err = t.translateCall(w, dot, nil, nil)
			}
			if err != nil {
				return nil, err
			}

			typ = method.Type.Out(0)
		} else if field, ok := typ.FieldByName(name); ok {
			fmt.Fprintf(w, ".%s", name)
			typ = field.Type
		} else {
			return nil, fmt.Errorf("Unknown field %s for type %s", name, typ.Name())
		}
	}
	return typ, nil
}
