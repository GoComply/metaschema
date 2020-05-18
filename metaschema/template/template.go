package template

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/gocomply/metaschema/metaschema/parser"
	"github.com/iancoleman/strcase"
	"github.com/markbates/pkger"
)

func GenerateTypes(metaschema *parser.Metaschema, outputDir string) error {
	t, err := newTemplate()
	if err != nil {
		return err
	}

	packageName := metaschema.GoPackageName()
	f, err := os.Create(fmt.Sprintf("%s/%s/%s.go", outputDir, packageName, packageName))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "types.tmpl", metaschema); err != nil {
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

func getImports(metaschema parser.Metaschema) string {
	var imports strings.Builder
	imports.WriteString("import (\n")
	if metaschema.ContainsRootElement() {
		imports.WriteString("\t\"encoding/xml\"\n")
	}

	for _, im := range metaschema.ImportedDependencies() {
		imports.WriteString(fmt.Sprintf("\n\t\"github.com/docker/oscalkit/types/oscal/%s\"\n", im.GoPackageName()))
	}

	imports.WriteString(")")

	return imports.String()
}

func newTemplate() (*template.Template, error) {
	in, err := pkger.Open("/metaschema/template/types.tmpl")
	if err != nil {
		return nil, err
	}
	defer in.Close()

	out, err := ioutil.TempFile("/tmp", "gocomply_metaschema.tmpl")
	if err != nil {
		return nil, err
	}
	defer out.Close()
	defer os.Remove(out.Name())

	_, err = io.Copy(out, in)
	if err != nil {
		return nil, err
	}

	return template.New("types.tmpl").Funcs(template.FuncMap{
		"toCamel":    strcase.ToCamel,
		"getImports": getImports,
	}).ParseFiles(out.Name())
}
