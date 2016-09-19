package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

var (
	valueReferenceRe = regexp.MustCompile(`^(?:(.+)\.)?([A-Za-z][A-Za-z0-9]*)$`)
	typeNameRe       = regexp.MustCompile(`^([^:]+):([^:]+):([\*\[\]]*)(?:(.+)\.)?([A-Za-z][A-Za-z0-9]*)$`)
)

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

func main() {
	flag.Parse()
	if len(targets) == 0 || flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	if packageName == "" {
		absOutputFile, err := filepath.Abs(outputFile)
		if err != nil {
			log.Fatal(err)
		}
		packageName = filepath.Base(filepath.Dir(absOutputFile))
	}

	var templateFiles []string
	for i := 0; i < flag.NArg(); i++ {
		matches, err := filepath.Glob(flag.Arg(i))
		if err != nil {
			log.Fatal(err)
		}
		templateFiles = append(templateFiles, matches...)
	}
	if len(templateFiles) == 0 {
		log.Fatal("no files found matching glob")
	}

	dir, err := ioutil.TempDir("", "statictemplate")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	goFile := filepath.Join(dir, "generate.go")
	file, err := os.Create(goFile)
	if err != nil {
		log.Fatal(err)
	}

	var funcMapImport, funcMapName string
	if funcMap != "" {
		values := valueReferenceRe.FindStringSubmatch(funcMap)
		if values == nil || values[1] == "" {
			log.Fatal(fmt.Errorf("invalid funcs value %q, expected <import>.<name>", funcMap))
		}
		funcMapImport = fmt.Sprintf("funcMapImport %q\n", values[1])
		funcMapName = fmt.Sprintf("funcMapImport.%s", values[2])
	}

	if err = writeTemplate(file, targets, templateFiles, html, funcMapImport, funcMapName); err != nil {
		log.Fatal(err)
	}
	if err = file.Close(); err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	if devOutputFile != "" {
		buf.WriteString("// +build !dev\n\n")
	}
	cmd := exec.Command("go", "run", goFile)
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		log.Fatal(err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		log.Fatal(err)
	}
	file, err = os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = file.Write(src); err != nil {
		log.Fatal(err)
	}
	file.Close()

	if devOutputFile != "" {
		buf.Reset()
		if err = writeDevTemplate(&buf, targets, templateFiles, html, funcMapImport, funcMapName, packageName); err != nil {
			log.Fatal(err)
		}
		src, err := format.Source(buf.Bytes())
		if err != nil {
			log.Fatal(err)
		}

		if contents, err := ioutil.ReadFile(devOutputFile); err != nil || !bytes.Equal(contents, src) {
			if err := os.MkdirAll(filepath.Dir(devOutputFile), 0755); err != nil {
				log.Fatal(err)
			}
			file, err := os.Create(devOutputFile)
			if err != nil {
				log.Fatal(err)
			}
			if _, err = buf.WriteTo(file); err != nil {
				log.Fatal(err)
			}
			file.Close()
		}
	}
}
