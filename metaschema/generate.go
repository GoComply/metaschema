package metaschema

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/gocomply/metaschema/metaschema/parser"
)

func Generate(metaschemaDir string) error {

	metaschemaPaths := map[string]string{
		"validation_root": "oscal_metadata_metaschema.xml",
		"nominal_catalog": "oscal_control-common_metaschema.xml",
		"catalog":         "oscal_catalog_metaschema.xml",
		"profile":         "oscal_profile_metaschema.xml",
		"implementation":  "oscal_implementation-common_metaschema.xml",
		"ssp":             "oscal_ssp_metaschema.xml",
		"component":       "oscal_component_metaschema.xml",
	}

	for _, metaschemaPath := range metaschemaPaths {
		f, err := os.Open(fmt.Sprintf("%s/%s", metaschemaDir, metaschemaPath))
		if err != nil {
			return err
		}
		defer f.Close()

		meta, err := decode(metaschemaDir, f)
		if err != nil {
			return err
		}

		if err := GenerateTypes(meta); err != nil {
			return err
		}
	}
	return nil
}

func decode(metaschemaDir string, r io.Reader) (*parser.Metaschema, error) {
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

		importedMeta, err := decode(metaschemaDir, imf)
		if err != nil {
			return nil, err
		}

		meta.ImportedMetaschema = append(meta.ImportedMetaschema, *importedMeta)
	}
	err := meta.LinkDefinitions()

	return &meta, err
}
