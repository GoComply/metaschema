package parser

import (
	"sort"
)

// Multiplexer represents model to be generated in go code that does not exists in the metaschema.
// Such model is only needed for serialization&deserialization as json and xml schemas differ
//materially in their structure.
type Multiplexer struct {
	Name             string
	MultiplexedModel MultiplexedModel
}

func (mplex *Multiplexer) GoTypeName() string {
	return mplex.GoTypeNameOriginal() + "Multiplexer"
}

func (mplex *Multiplexer) GoTypeNameOriginal() string {
	return mplex.MultiplexedModel.GoTypeName()
}

func (metaschema *Metaschema) getMultiplexer(name string) *Multiplexer {
	for _, m := range metaschema.ImportedMetaschema {
		mplex := m.getMultiplexer(name)
		if mplex != nil {
			return mplex
		}
	}
	for i, mplex := range metaschema.Multiplexers {
		if mplex.GoTypeName() == name {
			return &metaschema.Multiplexers[i]
		}
	}
	return nil
}

func (metaschema *Metaschema) calculateMultiplexers() []Multiplexer {
	uniq := map[string]Multiplexer{}
	for _, da := range metaschema.DefineAssembly {
		for i, a := range da.Model.Assembly {
			if a.requiresMultiplexer() {
				mplex := Multiplexer{MultiplexedModel: &da.Model.Assembly[i]}
				if metaschema.getMultiplexer(mplex.GoTypeName()) == nil {
					uniq[mplex.GoTypeName()] = mplex
				}
			}
		}
	}

	result := make([]Multiplexer, 0, len(uniq))
	for _, v := range uniq {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GoTypeName() < result[j].GoTypeName() })
	return result
}

func newMultiplexer(multiplexedModel MultiplexedModel) Multiplexer {
	return Multiplexer{
		MultiplexedModel: multiplexedModel,
	}

}

type MultiplexedModel interface {
	GoName() string
	GoTypeName() string
	groupAs() *GroupAs
}
