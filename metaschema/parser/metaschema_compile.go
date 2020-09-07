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

func (metaschema *Metaschema) linkAssemblies(list []Assembly) error {
	var err error
	for i, a := range list {
		if a.Ref != "" {
			a.Def, err = metaschema.GetDefineAssembly(a.Ref)
			if err != nil {
				return err
			}
			a.Metaschema = metaschema
			metaschema.registerDependency(a.Ref, a.Def)
			list[i] = a
		}
	}
	return nil
}

func (metaschema *Metaschema) linkFields(list []Field) error {
	var err error
	for i, f := range list {
		if f.Ref != "" {
			f.Def, err = metaschema.GetDefineField(f.Ref)
			if err != nil {
				return err
			}
			f.Metaschema = metaschema
			metaschema.registerDependency(f.Ref, f.Def)
			list[i] = f
		}
	}
	return nil
}

func (metaschema *Metaschema) linkFlags(list []Flag) error {
	var err error
	for i, f := range list {
		if f.Ref != "" {
			f.Def, err = metaschema.GetDefineFlag(f.Ref)
			if err != nil {
				return err
			}
			f.Metaschema = metaschema
			list[i] = f
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
		if err = metaschema.linkAssemblies(da.Model.Assembly); err != nil {
			return err
		}
		if err = metaschema.linkFields(da.Model.Field); err != nil {
			return err
		}
		for _, c := range da.Model.Choice {
			if err = metaschema.linkAssemblies(c.Assembly); err != nil {
				return err
			}
			if err = metaschema.linkFields(c.Field); err != nil {
				return err
			}
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
