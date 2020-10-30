package parser

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type GoType interface {
	GoTypeName() string
	GetMetaschema() *Metaschema
}

type GoStructItem interface {
	GoComment() string
	GoMemLayout() string
	GoName() string
	GoTypeNameMultiplexed() string
	JsonName() string
	XmlAnnotation() string
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

type Model struct {
	Assembly []Assembly `xml:"assembly"`
	Field    []Field    `xml:"field"`
	Choice   []Choice   `xml:"choice"`
	Prose    *struct{}  `xml:"prose"`
}

func (m *Model) GoStructItems() []GoStructItem {
	res := make([]GoStructItem, len(m.Assembly)+len(m.Field))
	for i, _ := range m.Field {
		res[i] = &m.Field[i]
	}
	for i, _ := range m.Assembly {
		res[i+len(m.Field)] = &m.Assembly[i]
	}
	return res
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

func handleMultiline(comment string) string {
	return strings.ReplaceAll(comment, "\n", "\n // ")
}

type JsonKey struct {
	FlagName string `xml:"flag-name,attr"`
}
