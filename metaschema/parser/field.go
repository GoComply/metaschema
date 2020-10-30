package parser

import (
	"github.com/iancoleman/strcase"
)

type DefineField struct {
	Name string `xml:"name,attr"`

	Flags        []Flag    `xml:"flag"`
	FormalName   string    `xml:"formal-name"`
	Description  string    `xml:"description"`
	Remarks      *Remarks  `xml:"remarks"`
	Examples     []Example `xml:"example"`
	AsType       AsType    `xml:"as-type,attr"`
	JsonValueKey string    `xml:"json-value-key"`
	Metaschema   *Metaschema
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
	return df.AsType == AsTypeMarkupMultiLine || df.AsType == AsTypeMarkupLine
}

func (df *DefineField) JsonName() string {
	if df.JsonValueKey != "" {
		return df.JsonValueKey
	}
	return "value"
}

func (df *DefineField) Empty() bool {
	return df.AsType == AsTypeEmpty
}

func (df *DefineField) GoName() string {
	return strcase.ToCamel(df.JsonName())
}

type Field struct {
	Required string `xml:"required,attr"`

	Description string   `xml:"description"`
	Remarks     *Remarks `xml:"remarks"`
	Ref         string   `xml:"ref,attr"`
	GroupAs     *GroupAs `xml:"group-as"`
	InXml       string   `xml:"in-xml,attr"`
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

func (f *Field) GoTypeNameMultiplexed() string {
	return f.GoTypeName()
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

func (f *Field) XmlAnnotation() string {
	if f.InXml == "UNWRAPPED" {
		return ",any"
	}
	return f.XmlName() + ",omitempty"

}

func (f *Field) compile(metaschema *Metaschema) error {
	if f.Ref != "" {
		var err error
		f.Metaschema = metaschema
		f.Def, err = f.Metaschema.GetDefineField(f.Ref)
		if err != nil {
			return err
		}
		f.Metaschema.registerDependency(f.Ref, f.Def)
	}
	return nil
}
