package main // import "bou.ke/statictemplate"

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	htmlTemplate "html/template"
	textTemplate "text/template"

	"bou.ke/statictemplate/internal"
	"bou.ke/statictemplate/statictemplate"
	"golang.org/x/tools/go/loader"
)

type dotType struct {
	packagePath string
	typeName    string
	prefix      string
}

type compilationTarget struct {
	functionName string
	templateName string
	dot          dotType
}

type compilationTargets []compilationTarget

func (c *compilationTargets) String() string {
	return ""
}

var typeNameRe = regexp.MustCompile(`^([^:]+):([^:]+):([\*\[\]]*)(?:(.+)\.)?([A-Za-z][A-Za-z0-9]*)$`)

func (c *compilationTargets) Set(value string) error {
	values := typeNameRe.FindStringSubmatch(value)
	if values == nil {
		return fmt.Errorf("expect compilation target in functionName:templateName:typeName format, got %q", value)
	}
	*c = append(*c, compilationTarget{values[1], values[2], dotType{
		prefix:      values[3],
		packagePath: values[4],
		typeName:    values[5],
	}})
	return nil
}

func (c compilationTargets) ToInstructions() (ins []statictemplate.TranslateInstruction, err error) {
	var conf loader.Config
	conf.Import("runtime")
	for _, t := range c {
		if p := t.dot.packagePath; p != "" {
			conf.Import(p)
		}
	}
	var prog *loader.Program
	prog, err = conf.Load()
	if err != nil {
		return
	}
	for _, t := range c {
		var pack *types.Package
		if t.dot.packagePath != "" {
			pack = prog.Package(t.dot.packagePath).Pkg
		}
		typVal, err := types.Eval(conf.Fset, pack, 0, t.dot.prefix+t.dot.typeName)
		if err != nil {
			return nil, err
		}
		ins = append(ins, statictemplate.TranslateInstruction{
			FunctionName: t.functionName,
			TemplateName: t.templateName,
			Dot:          typVal.Type,
		})
	}
	return
}

var (
	targets       compilationTargets
	packageName   string
	outputFile    string
	devOutputFile string
	glob          string
	html          bool
	funcMap       string
)

func init() {
	flag.Var(&targets, "t", "Target to process, supports multiple. The format is <function name>:<template name>:<type of the template argument>")
	flag.StringVar(&packageName, "package", "", "Name of the package of the result file. Defaults to name of the folder of the output file")
	flag.StringVar(&outputFile, "o", "template.go", "Name of the output file")
	flag.StringVar(&devOutputFile, "dev", "", "Name of the dev output file")
	flag.BoolVar(&html, "html", false, "Interpret templates as HTML, to enable Go's automatic HTML escaping")
	flag.StringVar(&funcMap, "funcs", "", "A reference to a custom Funcs map to include")
}

func parse(html bool, funcs map[string]*types.Func, files ...string) (interface{}, error) {
	var dummyFuncs map[string]interface{}
	if funcs != nil {
		dummyFuncs = make(map[string]interface{})
		for key := range funcs {
			dummyFuncs[key] = func() string {
				return ""
			}
		}
	}
	if html {
		t := htmlTemplate.New("")
		if dummyFuncs != nil {
			t.Funcs(dummyFuncs)
		}
		return t.ParseFiles(files...)
	} else {
		t := textTemplate.New("")
		if dummyFuncs != nil {
			t.Funcs(dummyFuncs)
		}
		return t.ParseFiles(files...)
	}
}

func main() {
	flag.Parse()
	if len(targets) == 0 || flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := work(); err != nil {
		log.Fatal(err)
	}
}

func work() error {
	if packageName == "" {
		absOutputFile, err := filepath.Abs(outputFile)
		if err != nil {
			return err
		}
		packageName = filepath.Base(filepath.Dir(absOutputFile))
	}

	var templateFiles []string
	for i := 0; i < flag.NArg(); i++ {
		matches, err := filepath.Glob(flag.Arg(i))
		if err != nil {
			return err
		}
		templateFiles = append(templateFiles, matches...)
	}
	if len(templateFiles) == 0 {
		log.Fatal("no files found matching glob")
	}

	funcMapImport, funcMapName, funcs, err := internal.ImportFuncMap(funcMap)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if devOutputFile != "" {
		buf.WriteString("// +build !dev\n\n")
	}

	template, err := parse(html, funcs, templateFiles...)
	if err != nil {
		return err
	}

	translator := statictemplate.New(template)
	translator.Funcs = funcs
	ins, err := targets.ToInstructions()
	if err != nil {
		return err
	}
	byts, err := translator.Translate(packageName, ins)
	if err != nil {
		return err
	}
	buf.Write(byts)

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return err
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	if _, err = file.Write(src); err != nil {
		return err
	}
	file.Close()

	if devOutputFile != "" {
		buf.Reset()
		if err = writeDevTemplate(&buf, targets, templateFiles, html, funcMapImport, funcMapName, packageName); err != nil {
			return err
		}
		src, err := format.Source(buf.Bytes())
		if err != nil {
			return err
		}

		if contents, err := ioutil.ReadFile(devOutputFile); err != nil || !bytes.Equal(contents, src) {
			if err := os.MkdirAll(filepath.Dir(devOutputFile), 0755); err != nil {
				return err
			}
			file, err := os.Create(devOutputFile)
			if err != nil {
				return err
			}
			if _, err = file.Write(src); err != nil {
				return err
			}
			file.Close()
		}
	}
	return nil
}
