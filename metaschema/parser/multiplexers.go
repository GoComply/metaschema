package parser

import (
	"sort"
)

// Multiplexer represents model to be generated in go code that does not exists in the metaschema.
// Such model is only needed for serialization&deserialization as json and xml schemas differ
//materially in their structure.
type Multiplexer struct {
	MultiplexedModel MultiplexedModel
	Metaschema       *Metaschema
}

func (mplex *Multiplexer) GoTypeName() string {
	return mplex.GoTypeNameOriginal() + "Multiplexer"
}

func (mplex *Multiplexer) GetMetaschema() *Metaschema {
	return mplex.Metaschema
}

func (mplex *Multiplexer) InJsonMap() bool {
	return mplex.MultiplexedModel.groupAs().ByKey()
}

func (mplex *Multiplexer) JsonKey() string {
	return mplex.MultiplexedModel.IndexBy()
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
		for i := range da.Model.Assembly {
			if requiresMultiplexer(&da.Model.Assembly[i]) {
				mplex := Multiplexer{
					MultiplexedModel: &da.Model.Assembly[i],
					Metaschema:       metaschema,
				}
				existing := metaschema.getMultiplexer(mplex.GoTypeName())
				if existing != nil {
					metaschema.registerDependency(mplex.GoTypeName(), existing)
				} else {
					uniq[mplex.GoTypeName()] = mplex
				}
			}
		}
		for i := range da.Model.Field {
			if requiresMultiplexer(&da.Model.Field[i]) {
				mplex := Multiplexer{
					MultiplexedModel: &da.Model.Field[i],
					Metaschema:       metaschema,
				}
				existing := metaschema.getMultiplexer(mplex.GoTypeName())
				if existing != nil {
					metaschema.registerDependency(mplex.GoTypeName(), existing)
				} else {
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

type MultiplexedModel interface {
	GoTypeName() string
	groupAs() *GroupAs
	IndexBy() string
}

func requiresMultiplexer(mm MultiplexedModel) bool {
	return mm.groupAs() != nil && mm.groupAs().requiresMultiplexer()
}
