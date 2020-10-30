package parser

import (
	"encoding/xml"
	"fmt"
	"github.com/iancoleman/strcase"
	"net/url"
	"sort"
	"strings"
)

const (
	AsTypeBoolean            AsType = "boolean"
	AsTypeEmpty              AsType = "empty"
	AsTypeString             AsType = "string"
	AsTypeMixed              AsType = "mixed"
	AsTypeMarkupLine         AsType = "markup-line"
	AsTypeMarkupMultiLine    AsType = "markup-multiline"
	AsTypeDate               AsType = "date"
	AsTypeDateTimeTZ         AsType = "dateTime-with-timezone"
	AsTypeNCName             AsType = "NCName"
	AsTypeEmail              AsType = "email"
	AsTypeURI                AsType = "uri"
	AsTypeBase64             AsType = "base64Binary"
	AsTypeIDRef              AsType = "IDREF"
	AsTypeNMToken            AsType = "NMTOKEN"
	AsTypeID                 AsType = "ID"
	AsTypeAnyURI             AsType = "anyURI"
	AsTypeURIRef             AsType = "uri-reference"
	AsTypeUUID               AsType = "uuid"
	AsTypeNonNegativeInteger AsType = "nonNegativeInteger"
)

type GoType interface {
	GoTypeName() string
	GetMetaschema() *Metaschema
}

// Metaschema is the root metaschema element
type Metaschema struct {
	XMLName xml.Name `xml:"http://csrc.nist.gov/ns/oscal/metaschema/1.0 METASCHEMA"`
	Top     string   `xml:"top,attr"`
	Root    string   `xml:"root,attr"`

	// SchemaName describes the scope of application of the data format. For
	// example "OSCAL Catalog"
	SchemaName *SchemaName `xml:"schema-name"`

	// ShortName is a coded version of the schema name for use when a string-safe
	// identifier is needed. For example "oscal-catalog"
	ShortName *ShortName `xml:"short-name"`

	// Remarks are paragraphs describing the metaschema
	Remarks *Remarks `xml:"remarks,omitempty"`

	// Import is a URI to an external metaschema
	Import []Import `xml:"import"`

	// DefineAssembly is one or more assembly definitions
	DefineAssembly []DefineAssembly `xml:"define-assembly"`

	// DefineField is one or more field definitions
	DefineField []DefineField `xml:"define-field"`

	// DefineFlag is one or more flag definitions
	DefineFlag []DefineFlag `xml:"define-flag"`

	ImportedMetaschema []Metaschema
	Dependencies       map[string]GoType
	Multiplexers       []Multiplexer
	GoMod              string
}

func (metaschema *Metaschema) ImportedDependencies() []*Metaschema {
	result := make(map[string]*Metaschema)
	for _, dep := range metaschema.Dependencies {
		m := dep.GetMetaschema()
		result[m.GoPackageName()] = m
	}

	ret := make([]*Metaschema, 0, len(result))
	for _, v := range result {
		ret = append(ret, v)
	}
	sort.Slice(ret, func(i, j int) bool {
		return strings.Compare(ret[i].GoPackageName(), ret[j].GoPackageName()) > 0
	})
	return ret
}

func (metaschema *Metaschema) GoPackageName() string {
	return strings.ReplaceAll(strings.ToLower(metaschema.Root), "-", "_")
}

func (Metaschema *Metaschema) ContainsRootElement() bool {
	for _, v := range Metaschema.DefineAssembly {
		if v.RepresentsRootElement() {
			return true
		}
	}
	return false
}

type DefineFlag struct {
	Name   string `xml:"name,attr"`
	AsType AsType `xml:"as-type,attr"`

	FormalName  string    `xml:"formal-name"`
	Description string    `xml:"description"`
	Remarks     *Remarks  `xml:"remarks"`
	Examples    []Example `xml:"example"`
	Metaschema  *Metaschema
}

func (df *DefineFlag) GoTypeName() string {
	return strcase.ToCamel(df.Name)
}

func (df *DefineFlag) GetMetaschema() *Metaschema {
	return df.Metaschema
}

type Model struct {
	Assembly []Assembly `xml:"assembly"`
	Field    []Field    `xml:"field"`
	Choice   []Choice   `xml:"choice"`
	Prose    *struct{}  `xml:"prose"`
}

type Flag struct {
	Name     string `xml:"name,attr"`
	AsType   AsType `xml:"as-type,attr"`
	Required string `xml:"required,attr"`

	Description string   `xml:"description"`
	Remarks     *Remarks `xml:"remarks"`
	Values      []Value  `xml:"value"`
	Ref         string   `xml:"ref,attr"`
	Def         *DefineFlag
	Metaschema  *Metaschema
}

func (f *Flag) GoComment() string {
	if f.Description != "" {
		return handleMultiline(f.Description)
	}
	return handleMultiline(f.Def.Description)
}

func (f *Flag) GoDatatype() (string, error) {
	dt := f.AsType
	if dt == "" {
		if f.Ref == "" {
			// workaround bug: inline definition without type hint https://github.com/usnistgov/OSCAL/pull/570
			return "string", nil
		}
		dt = f.Def.AsType
	}

	if dt == "" {
		return "string", nil
	}
	if goDatatypeMap[dt] == "" {
		return "", fmt.Errorf("Unknown as-type='%s' found at <%s> definition", dt, f.Ref)
	}
	return goDatatypeMap[dt], nil
}

func (f *Flag) GoTypeName() string {
	if f.Name != "" {
		return strcase.ToCamel(f.Name)
	}
	return f.Def.GoTypeName()
}

func (f *Flag) GoName() string {
	return strcase.ToCamel(f.JsonName())
}

func (f *Flag) JsonName() string {
	return f.XmlName()
}
func (f *Flag) XmlName() string {
	if f.Name != "" {
		return f.Name
	}
	return f.Def.Name
}

type Choice struct {
	Field    []Field    `xml:"field"`
	Assembly []Assembly `xml:"assembly"`
}

type GroupAs struct {
	Name   string `xml:"name,attr"`
	InJson string `xml:"in-json,attr"`
	InXml  string `xml:"in-xml,attr"`
}

func (ga *GroupAs) SingletonOrArray() bool {
	// default: SINGLETON_OR_ARRAY
	return ga.InJson == "" || ga.InJson == "SINGLETON_OR_ARRAY"
}

func (ga *GroupAs) ByKey() bool {
	return ga.InJson == "BY_KEY"
}

func (ga *GroupAs) requiresMultiplexer() bool {
	return ga.ByKey() || ga.SingletonOrArray()
}

type Import struct {
	Href *Href `xml:"href,attr"`
}

// Remarks are descriptions for a particular metaschema, assembly, field, flag
// or example
type Remarks struct {
	P []P `xml:"p"`

	InnerXML string `xml:",innerxml"`
}

type Value struct {
	InnerXML string `xml:",innerxml"`
}

type Title struct {
	Code []string `xml:"code"`
	Q    []string `xml:"q"`

	InnerXML string `xml:",innerxml"`
}

type ShortName struct {
	Code []string `xml:"code"`
	Q    []string `xml:"q"`

	InnerXML string `xml:",innerxml"`
}

type SchemaName struct {
	Code []string `xml:"code"`
	Q    []string `xml:"q"`

	InnerXML string `xml:",innerxml"`
}

type Example struct {
	Href *Href  `xml:"href,attr"`
	Path string `xml:"path,attr"`

	Description string   `xml:"description"`
	Remarks     *Remarks `xml:"remarks"`

	InnerXML string `xml:",innerxml"`
}

type P struct {
	A      []A      `xml:"a"`
	Code   []string `xml:"code"`
	Q      []string `xml:"q"`
	EM     []string `xml:"em"`
	Strong []string `xml:"strong"`

	CharData string `xml:",chardata"`
}

type A struct {
	XMLName xml.Name `xml:"a"`
	Href    *Href    `xml:"href,attr"`

	CharData      string `xml:",chardata"`
	ProcessedLink string `xml:"-"`
}

func (a *A) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type anchor A

	if err := d.DecodeElement((*anchor)(a), &start); err != nil {
		return err
	}

	if a.Href != nil {
		a.ProcessedLink = fmt.Sprintf("%s (%s)", a.CharData, a.Href.URL.String())
	}

	return nil
}

type Href struct {
	URL *url.URL
}

func (h *Href) UnmarshalXMLAttr(attr xml.Attr) error {
	URL, err := url.Parse(attr.Value)
	if err != nil {
		return err
	}

	h.URL = URL

	return nil
}

type AsType string

var goDatatypeMap = map[AsType]string{
	AsTypeString:             "string",
	AsTypeIDRef:              "string",
	AsTypeNCName:             "string",
	AsTypeNMToken:            "string",
	AsTypeID:                 "string",
	AsTypeURIRef:             "string",
	AsTypeURI:                "string",
	AsTypeUUID:               "string",
	AsTypeNonNegativeInteger: "uint64",
}

func handleMultiline(comment string) string {
	return strings.ReplaceAll(comment, "\n", "\n // ")
}

type JsonKey struct {
	FlagName string `xml:"flag-name,attr"`
}
