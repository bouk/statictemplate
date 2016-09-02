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

func Translate(name, text string, dot reflect.Type) (string, error) {
	return (&translator{
		funcs: map[string]interface{}{
			"print":  fmt.Sprint,
			"printf": fmt.Sprintf,
		},
	}).translate(name, text, dot)
}

type translator struct {
	funcs template.FuncMap
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
		io.WriteString(w, "_, _ = fmt.Fprint(w, ")
		if _, err := t.translatePipe(w, dot, node.Pipe); err != nil {
			return err
		}
		_, err := io.WriteString(w, ")\n")
		return err
	case *parse.WithNode:
		t.translateIfOrWith(w, dot, node.Type(), node.Pipe, node.List, node.ElseList)
		return nil
	case *parse.IfNode:
		t.translateIfOrWith(w, dot, node.Type(), node.Pipe, node.List, node.ElseList)
		return nil
	default:
		return fmt.Errorf("Unknown Node %s", node.Type())
	}
}

func (t *translator) translateIfOrWith(w io.Writer, dot reflect.Type, nodeType parse.NodeType, pipe *parse.PipeNode, list, elseList *parse.ListNode) error {
	io.WriteString(w, "if eval := ")
	typ, err := t.translatePipe(w, dot, pipe)
	if err != nil {
		return err
	}
	io.WriteString(w, "; ")
	switch typ.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		io.WriteString(w, "len(eval) != 0")
	case reflect.Bool:
		io.WriteString(w, "eval")
	case reflect.Ptr:
		io.WriteString(w, "eval != nil")
	case reflect.Struct:
		io.WriteString(w, "true")
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Float32, reflect.Float64:
		io.WriteString(w, "eval != 0")
	default:
		return fmt.Errorf("Don't know how to evaluate %s", typ)
	}
	io.WriteString(w, "{\n")
	if nodeType == parse.NodeWith {
		io.WriteString(w, "dot := eval\n")
	}

	if err := t.translateNode(list, w, dot); err != nil {
		return err
	}
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
	if len(pipe.Decl) != 0 {
		return nil, fmt.Errorf("Dunno what to do with decls %s", pipe)
	}

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
	case *parse.IdentifierNode:
		return t.translateFunction(w, dot, action, args, nextCommands)
	case *parse.FieldNode:
		return t.translateField(w, dot, action, args, nextCommands)
	case *parse.PipeNode:
		return t.translatePipe(w, dot, action)
	case *parse.ChainNode:
		typ, err := t.translateArg(w, dot, action.Node)
		if err != nil {
			return nil, err
		}
		return t.translateFieldChain(w, dot, typ, action.Field, args, nextCommands)
	}

	if len(args) > 0 || len(nextCommands) > 0 {
		return nil, fmt.Errorf("Dunno what to do with args %v %s %v", cmd.Args, action.Type(), nextCommands)
	}

	switch action := action.(type) {
	case *parse.StringNode:
		_, err := fmt.Fprintf(w, "%q", action.Text)
		return reflect.TypeOf(""), err
	case *parse.NumberNode:
		if action.IsInt {
			_, err := fmt.Fprint(w, action.Int64)
			return reflect.TypeOf(int64(0)), err
		} else {
			return nil, fmt.Errorf("Unknown number node %v", action)
		}
	case *parse.DotNode:
		_, err := io.WriteString(w, "dot")
		return dot, err
	case *parse.BoolNode:
		_, err := fmt.Fprint(w, action.True)
		return reflect.TypeOf(true), err
	case *parse.NilNode:
		return nil, fmt.Errorf("nil is not a command")
	default:
		return nil, fmt.Errorf("Unknown pipe node %s, %s", action.String(), action.Type())
	}
}

func (t *translator) translateArg(w io.Writer, dot reflect.Type, arg parse.Node) (reflect.Type, error) {
	switch arg := arg.(type) {
	case *parse.IdentifierNode:
		return t.translateFunction(w, dot, arg, nil, nil)
	case *parse.FieldNode:
		return t.translateField(w, dot, arg, nil, nil)
	case *parse.PipeNode:
		return t.translatePipe(w, dot, arg)
	case *parse.StringNode:
		_, err := fmt.Fprintf(w, "%q", arg.Text)
		return reflect.TypeOf(""), err
	case *parse.NumberNode:
		if arg.IsInt {
			_, err := fmt.Fprint(w, arg.Int64)
			return reflect.TypeOf(int64(0)), err
		} else {
			return nil, fmt.Errorf("Unknown number node %v", arg)
		}
	case *parse.DotNode:
		_, err := io.WriteString(w, "dot")
		return dot, err
	case *parse.BoolNode:
		_, err := fmt.Fprint(w, arg.True)
		return reflect.TypeOf(true), err
	case *parse.NilNode:
		_, err := io.WriteString(w, "nil")
		return reflect.TypeOf(nil), err
	default:
		return nil, fmt.Errorf("Unknown arg %s, %s", arg.String(), arg.Type())
	}
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
