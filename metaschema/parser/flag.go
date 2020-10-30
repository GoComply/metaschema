package parser

import (
	"fmt"
	"github.com/iancoleman/strcase"
)

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
