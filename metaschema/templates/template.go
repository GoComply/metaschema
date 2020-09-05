package templates

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gocomply/metaschema/metaschema/parser"
	"github.com/markbates/pkger"
)

func GenerateAll(metaschema *parser.Metaschema, outputDir string) error {
	return GenerateTypes(metaschema, outputDir)
}

func GenerateTypes(metaschema *parser.Metaschema, outputDir string) error {
	t, err := newTemplate(outputDir)
	if err != nil {
		return err
	}

	packageName := metaschema.GoPackageName()
	dir := filepath.Join(outputDir, packageName)
	err = os.MkdirAll(dir, os.FileMode(0722))
	if err != nil {
		return err
	}
	f, err := os.Create(fmt.Sprintf("%s/generated_models.go", dir))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, metaschema); err != nil {
		return err
	}

	p, err := format.Source(buf.Bytes())
	if err != nil {
		return errors.New(err.Error() + " in following file:\n" + string(buf.Bytes()))
	}

	_, err = f.Write(p)
	if err != nil {
		return err
	}

	return nil
}

func newTemplate(outputDir string) (*template.Template, error) {
	getImports := func(metaschema parser.Metaschema) string {
		var imports strings.Builder
		imports.WriteString("import (\n")
		if metaschema.ContainsRootElement() {
			imports.WriteString("\t\"encoding/xml\"\n")
		}

		for _, im := range metaschema.ImportedDependencies() {
			imports.WriteString(fmt.Sprintf("\n\t\"%s/%s/%s\"\n", metaschema.GoMod, outputDir, im.GoPackageName()))
		}

		imports.WriteString(")")

		return imports.String()
	}

	in, err := pkger.Open("/metaschema/templates/generated_models.tmpl")
	if err != nil {
		return nil, err
	}
	defer in.Close()

	tempText, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	return template.New("generated_models.tmpl").Funcs(template.FuncMap{
		"getImports": getImports,
	}).Parse(string(tempText))
}
