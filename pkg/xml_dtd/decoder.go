package xml_dtd

import (
	"encoding/xml"
	"io"
)

type Decoder struct {
	*xml.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	dtd := &Decoder{
		Decoder: xml.NewDecoder(r),
	}
	dtd.Entity = map[string]string{
		"allowed-values-control-group-property-name": "TODO-FIX-THIS",
		"allowed-values-component_inventory-item_property-name": "TODO-FIX-THIS",
		"allowed-values-component_component_property-name": "TODO-FIX-THIS",
	}
	return dtd
}

func (d *Decoder) Decode(v interface{}) error {
	return d.Decoder.Decode(v)
}
