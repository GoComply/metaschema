package parser

import (
	"github.com/iancoleman/strcase"
)

// DefineAssembly is a definition for for an object or element that contains
// structured content
type DefineAssembly struct {
	Name    string `xml:"name,attr"`
	Address string `xml:"address,attr"`

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

func (a *Assembly) XmlGroupping() string {
	if a.GroupAs == nil {
		return ""
	}
	if a.GroupAs.InXml == "UNGROUPED" || a.GroupAs.InXml == "" {
		return ""
	}
	if a.GroupAs.InXml == "GROUPED" {
		return a.GroupAs.Name + ">"
	}
	panic("Not implemented group-as/@in-xml=" + a.GroupAs.InXml)
	return ""
}

func (a *Assembly) XmlAnnotation() string {
	return a.XmlGroupping() + a.XmlName() + ",omitempty"
}
