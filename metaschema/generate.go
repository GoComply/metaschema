package metaschema

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gocomply/metaschema/metaschema/parser"
	"github.com/gocomply/metaschema/metaschema/templates"
)

func Generate(metaschemaDir, goModule, outputDir string) error {
	files, err := ioutil.ReadDir(metaschemaDir)
	if err != nil {
		return err
	}
	for _, metaschemaPath := range files {
		if !strings.HasSuffix(metaschemaPath.Name(), ".xml") {
			continue
		}
		fmt.Println("Processing ", metaschemaPath.Name())
		f, err := os.Open(fmt.Sprintf("%s/%s", metaschemaDir, metaschemaPath.Name()))
		if err != nil {
			return err
		}
		defer f.Close()

		meta, err := decode(metaschemaDir, goModule, f)
		if err != nil {
			return err
		}

		if err := templates.GenerateAll(meta, outputDir); err != nil {
			return err
		}
	}
	return nil
}

func decode(metaschemaDir, goModule string, r io.Reader) (*parser.Metaschema, error) {
	var meta parser.Metaschema

	d := xml.NewDecoder(r)

	if err := d.Decode(&meta); err != nil {
		return nil, fmt.Errorf("Error decoding metaschema: %s", err)
	}

	for _, imported := range meta.Import {
		if imported.Href == nil {
			return nil, fmt.Errorf("import element in %s is missing 'href' attribute", r)
		}
		imf, err := os.Open(fmt.Sprintf("%s/%s", metaschemaDir, imported.Href.URL.String()))
		if err != nil {
			return nil, err
		}
		defer imf.Close()

		importedMeta, err := decode(metaschemaDir, goModule, imf)

		if err != nil {
			return nil, err
		}

		meta.ImportedMetaschema = append(meta.ImportedMetaschema, *importedMeta)
	}
	err := meta.LinkDefinitions()
	meta.GoMod = goModule

	return &meta, err
}
