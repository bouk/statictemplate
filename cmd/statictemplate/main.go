package main

import (
	"flag"
	"fmt"
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

var typeNameRe = regexp.MustCompile(`^([^:]+):([^:]+):([\*\[\]]*)(?:([^\.]+)\.)?([A-Za-z][A-Za-z0-9]*)$`)

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
	targets     compilationTargets
	packageName string
	outputFile  string
	glob        string
)

func init() {
	flag.Var(&targets, "t", "Target to process, supports multiple. The format is <function name>:<template name>:<type of the template argument>")
	flag.StringVar(&packageName, "package", "", "Name of the package of the result file. Defaults to name of the folder of the output file")
	flag.StringVar(&outputFile, "o", "template.go", "Name of the output file")
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

	writeTemplate(file, targets, templateFiles)
	if err = file.Close(); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		log.Fatal(err)
	}
	outputFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	cmd := exec.Command("go", "run", goFile)
	cmd.Stdout = outputFile
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
