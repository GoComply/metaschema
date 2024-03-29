package parser

import (
	"fmt"
)

func (metaschema *Metaschema) Compile() error {
	err := metaschema.linkDefinitions()
	if err != nil {
		return err
	}
	metaschema.Multiplexers = metaschema.calculateMultiplexers()
	return nil
}

func (metaschema *Metaschema) registerDependency(name string, dependency GoType) {
	if dependency.GetMetaschema() != metaschema {
		if metaschema.Dependencies == nil {
			metaschema.Dependencies = make(map[string]GoType)
		}
		if _, ok := metaschema.Dependencies[name]; !ok {
			metaschema.Dependencies[name] = dependency
		}
	}
}

func (metaschema *Metaschema) linkItems(list []GoStructItem) error {
	for i := range list {
		err := list[i].compile(metaschema)
		if err != nil {
			return err
		}
	}
	return nil
}
func (metaschema *Metaschema) linkAssemblies(list []Assembly) error {
	for i := range list {
		a := &list[i]
		err := a.compile(metaschema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (metaschema *Metaschema) linkFields(list []Field) error {
	for i := range list {
		f := &list[i]
		err := f.compile(metaschema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (metaschema *Metaschema) linkFlags(list []Flag) error {
	for i := range list {
		f := &list[i]
		err := f.compile(metaschema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (metaschema *Metaschema) linkDefinitions() error {
	var err error
	for _, da := range metaschema.DefineAssembly {
		if err = metaschema.linkFlags(da.Flags); err != nil {
			return err
		}
		if err = metaschema.linkItems(da.Model.sortedChilds); err != nil {
			return err
		}
		if err = metaschema.linkAssemblies(da.Model.Assembly); err != nil {
			return err
		}
		if err = metaschema.linkFields(da.Model.Field); err != nil {
			return err
		}
	}

	for _, df := range metaschema.DefineField {
		if err = metaschema.linkFlags(df.Flags); err != nil {
			return err
		}
	}
	return nil
}

func (metaschema *Metaschema) GetDefineField(name string) (*DefineField, error) {
	for _, v := range metaschema.DefineField {
		if name == v.Name {
			if v.Metaschema == nil {
				v.Metaschema = metaschema
			}
			return &v, nil
		}
	}
	for _, m := range metaschema.ImportedMetaschema {
		f, err := m.GetDefineField(name)
		if err == nil {
			return f, err
		}
	}
	return nil, fmt.Errorf("Could not find define-field element with name='%s'.", name)
}

func (metaschema *Metaschema) GetDefineAssembly(name string) (*DefineAssembly, error) {
	for _, v := range metaschema.DefineAssembly {
		if name == v.Name {
			if v.Metaschema == nil {
				v.Metaschema = metaschema
			}
			return &v, nil
		}
	}
	for _, m := range metaschema.ImportedMetaschema {
		a, err := m.GetDefineAssembly(name)
		if err == nil {
			return a, err
		}
	}
	return nil, fmt.Errorf("Could not find define-assembly element with name='%s'.", name)
}

func (metaschema *Metaschema) GetDefineFlag(name string) (*DefineFlag, error) {
	for _, v := range metaschema.DefineFlag {
		if name == v.Name {
			if v.Metaschema == nil {
				v.Metaschema = metaschema
			}
			return &v, nil
		}
	}
	for _, m := range metaschema.ImportedMetaschema {
		f, err := m.GetDefineFlag(name)
		if err == nil {
			return f, err
		}
	}
	return nil, fmt.Errorf("Could not find define-flag element with name='%s'.", name)
}
