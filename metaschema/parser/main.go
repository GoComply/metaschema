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
	AsTypeBoolean         AsType = "boolean"
	AsTypeEmpty           AsType = "empty"
	AsTypeString          AsType = "string"
	AsTypeMixed           AsType = "mixed"
	AsTypeMarkupLine      AsType = "markup-line"
	AsTypeMarkupMultiLine AsType = "markup-multiline"
	AsTypeDate            AsType = "date"
	AsTypeDateTimeTZ      AsType = "dateTime-with-timezone"
	AsTypeNCName          AsType = "NCName"
	AsTypeEmail           AsType = "email"
	AsTypeURI             AsType = "uri"
	AsTypeBase64          AsType = "base64Binary"

	ShowDocsXML     ShowDocs = "xml"
	ShowDocsJSON    ShowDocs = "json"
	ShowDocsXMLJSON ShowDocs = "xml json"
)

var ShowDocsOptions = []ShowDocs{
	ShowDocsXML,
	ShowDocsJSON,
	ShowDocsXMLJSON,
}

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

// DefineAssembly is a definition for for an object or element that contains
// structured content
type DefineAssembly struct {
	Name     string `xml:"name,attr"`
	ShowDocs string `xml:"show-docs,attr"`
	Address  string `xml:"address,attr"`

	JsonKey     *JsonKey  `xml:"json-key"`
	Flags       []Flag    `xml:"flag"`
	FormalName  string    `xml:"formal-name"`
	Description string    `xml:"description"`
	Remarks     *Remarks  `xml:"remarks"`
	Model       *Model    `xml:"model"`
	Examples    []Example `xml:"example"`
	Metaschema  *Metaschema
}

func (da *DefineAssembly) GoTypeName() string {
	return strcase.ToCamel(da.Name)
}

func (da *DefineAssembly) RepresentsRootElement() bool {
	return da.Name == "catalog" || da.Name == "profile" || da.Name == "declarations" || da.Name == "system-security-plan"
}

func (a *DefineAssembly) GoComment() string {
	return handleMultiline(a.Description)
}

func (a *DefineAssembly) GetMetaschema() *Metaschema {
	return a.Metaschema
}

type DefineField struct {
	Name     string `xml:"name,attr"`
	ShowDocs string `xml:"show-docs,attr"`

	Flags       []Flag    `xml:"flag"`
	FormalName  string    `xml:"formal-name"`
	Description string    `xml:"description"`
	Remarks     *Remarks  `xml:"remarks"`
	Examples    []Example `xml:"example"`
	AsType      AsType    `xml:"as-type,attr"`
	Metaschema  *Metaschema
}

func (df *DefineField) GoTypeName() string {
	return strcase.ToCamel(df.Name)
}

func (df *DefineField) requiresPointer() bool {
	return len(df.Flags) > 0 || df.IsMarkup()
}

func (f *DefineField) GoComment() string {
	return handleMultiline(f.Description)
}

func (df *DefineField) GetMetaschema() *Metaschema {
	return df.Metaschema
}

func (df *DefineField) IsMarkup() bool {
	return df.AsType == AsTypeMarkupMultiLine
}

type DefineFlag struct {
	Name     string   `xml:"name,attr"`
	AsType   datatype `xml:"as-type,attr"`
	ShowDocs ShowDocs `xml:"show-docs,attr"`

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
	Any      *struct{}  `xml:"any"`
}

type Assembly struct {
	Description string   `xml:"description"`
	Remarks     *Remarks `xml:"remarks"`
	Ref         string   `xml:"ref,attr"`
	GroupAs     *GroupAs `xml:"group-as"`
	Def         *DefineAssembly
	Metaschema  *Metaschema
}

func (a *Assembly) GoComment() string {
	if a.Description != "" {
		return handleMultiline(a.Description)
	}
	return a.Def.GoComment()
}

func (a *Assembly) GoTypeName() string {
	return a.Def.GoTypeName()
}

func (a *Assembly) GoTypeNameMultiplexed() string {
	if requiresMultiplexer(a) {
		return (&Multiplexer{MultiplexedModel: a}).GoTypeName()
	}
	return a.GoTypeName()
}

func (a *Assembly) GoMemLayout() string {
	if requiresMultiplexer(a) {
		return ""
	}
	if a.GroupAs != nil {
		return "[]"
	}
	return "*"
}

func (a *Assembly) GoName() string {
	return strcase.ToCamel(a.JsonName())
}

func (a *Assembly) JsonName() string {
	if a.GroupAs != nil {
		return a.GroupAs.Name
	}
	return strcase.ToLowerCamel(a.XmlName())
}

func (a *Assembly) XmlName() string {
	return a.Def.Name
}

func (a *Assembly) GoPackageName() string {
	if a.Ref == "" {
		return ""
	} else if a.Def.Metaschema == a.Metaschema {
		return ""
	} else {
		return a.Def.Metaschema.GoPackageName() + "."
	}
}

func (a *Assembly) groupAs() *GroupAs {
	return a.GroupAs
}

func (a *Assembly) IndexBy() string {
	if a.Def == nil {
		panic("Not implemented: IndexBy requires define-assembly to exists")
	}
	if a.Def.JsonKey == nil {
		panic("Not implemented: IndexBy requires define-assembly to define <json-key>")
	}
	return strcase.ToCamel(a.Def.JsonKey.FlagName)
}

type Field struct {
	Required string `xml:"required,attr"`

	Description string   `xml:"description"`
	Remarks     *Remarks `xml:"remarks"`
	Ref         string   `xml:"ref,attr"`
	GroupAs     *GroupAs `xml:"group-as"`
	Def         *DefineField
	Metaschema  *Metaschema
}

func (f *Field) GoComment() string {
	if f.Description != "" {
		return handleMultiline(f.Description)
	}
	return f.Def.GoComment()
}

func (f *Field) requiresPointer() bool {
	return f.Def.requiresPointer()
}

func (f *Field) GoTypeName() string {
	return f.Def.GoTypeName()
}

func (f *Field) GoPackageName() string {
	if f.Ref == "" {
		return ""
	} else if f.Def.Metaschema == f.Metaschema {
		return ""
	} else {
		return f.Def.Metaschema.GoPackageName() + "."
	}
}

func (f *Field) GoMemLayout() string {
	if f.GroupAs != nil {
		return "[]"
	} else if f.requiresPointer() {
		return "*"
	}
	return ""
}

func (f *Field) GoName() string {
	return strcase.ToCamel(f.JsonName())
}

func (f *Field) JsonName() string {
	if f.GroupAs != nil {
		return f.GroupAs.Name
	}
	return f.XmlName()
}

func (f *Field) XmlName() string {
	return f.Def.Name
}

type Flag struct {
	Name     string   `xml:"name,attr"`
	AsType   datatype `xml:"as-type,attr"`
	Required string   `xml:"required,attr"`

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
	return strcase.ToLowerCamel(f.XmlName())
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

func (h *Href) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if h.URL != nil {
		rawURI := h.URL.String()

		return xml.Attr{Name: name, Value: rawURI}, nil
	}

	return xml.Attr{Name: name}, nil
}

type AsType string

type ShowDocs string

func (sd ShowDocs) UnmarshalXMLAttr(attr xml.Attr) error {
	showDocs := ShowDocs(attr.Value)

	for _, showDocsOption := range ShowDocsOptions {
		if showDocs == showDocsOption {
			sd = showDocs
			return nil
		}
	}

	return fmt.Errorf("Show docs option \"%s\" is not a valid option", attr.Value)
}

type datatype string

const (
	datatypeString             datatype = "string"
	datatypeIDRef              datatype = "IDREF"
	datatypeNCName             datatype = "NCName"
	datatypeNMToken            datatype = "NMTOKEN"
	datatypeID                 datatype = "ID"
	datatypeAnyURI             datatype = "anyURI"
	datatypeURIRef             datatype = "uri-reference"
	datatypeURI                datatype = "uri"
	datatypeUUID               datatype = "uuid"
	datatypeNonNegativeInteger datatype = "nonNegativeInteger"
)

var goDatatypeMap = map[datatype]string{
	datatypeString:             "string",
	datatypeIDRef:              "string",
	datatypeNCName:             "string",
	datatypeNMToken:            "string",
	datatypeID:                 "string",
	datatypeURIRef:             "string",
	datatypeURI:                "string",
	datatypeUUID:               "string",
	datatypeNonNegativeInteger: "uint64",
}

func handleMultiline(comment string) string {
	return strings.ReplaceAll(comment, "\n", "\n // ")
}

type JsonKey struct {
	FlagName string `xml:"flag-name,attr"`
}
