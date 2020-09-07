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

func (metaschema *Metaschema) calculateMultiplexers() []Multiplexer {
	uniq := map[string]MultiplexedModel{}
	for _, da := range metaschema.DefineAssembly {
		for i, a := range da.Model.Assembly {
			if a.requiresMultiplexer() {
				uniq[a.GoTypeName()] = &da.Model.Assembly[i]
			}
		}
	}

	result := make([]Multiplexer, 0, len(uniq))
	for _, v := range uniq {
		result = append(result, Multiplexer{
			MultiplexedModel: v,
		})
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
