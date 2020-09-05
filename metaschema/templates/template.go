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

func GenerateAll(metaschema *parser.Metaschema, baseDir string) error {
	pkgDir, err := ensurePkgDir(metaschema, baseDir)
	if err != nil {
		return err
	}
	for _, templateName := range []string{"generated_models"} {
		t, err := newTemplate(baseDir, templateName)
		if err != nil {
			return err
		}
		err = executeTemplate(t, metaschema, fmt.Sprintf("%s/%s.go", pkgDir, templateName))
		if err != nil {
			return err
		}
	}
	return nil
}

func executeTemplate(t *template.Template, metaschema *parser.Metaschema, filename string) error {
	f, err := os.Create(filename)
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
	return err
}

func newTemplate(baseDir, templateName string) (*template.Template, error) {
	getImports := func(metaschema parser.Metaschema) string {
		var imports strings.Builder
		imports.WriteString("import (\n")
		if metaschema.ContainsRootElement() {
			imports.WriteString("\t\"encoding/xml\"\n")
		}

		for _, im := range metaschema.ImportedDependencies() {
			imports.WriteString(fmt.Sprintf("\n\t\"%s/%s/%s\"\n", metaschema.GoMod, baseDir, im.GoPackageName()))
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

	return template.New(templateName + ".tmpl").Funcs(template.FuncMap{
		"getImports": getImports,
	}).Parse(string(tempText))
}

func ensurePkgDir(metaschema *parser.Metaschema, baseDir string) (string, error) {
	dir := filepath.Join(baseDir, metaschema.GoPackageName())
	err := os.MkdirAll(dir, os.FileMode(0722))
	return dir, err
}
