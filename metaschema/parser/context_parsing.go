package parser

import (
	"encoding/xml"
)

type parentElement interface {
	RegisterChild(GoStructItem)
}

type contextParser []parentElement

var context contextParser

func (cp *contextParser) Push(p parentElement) {
	*cp = append(*cp, p)
}

func (cp *contextParser) Pop() {
	*cp = (*cp)[:len(*cp)-1]
}

func (cp *contextParser) RegisterChild(child GoStructItem) {
	(*cp)[len(*cp)-1].RegisterChild(child)
}

func (m *Model) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type model Model
	context.Push(m)
	defer context.Pop()
	return d.DecodeElement((*model)(m), &start)
}

func (m *Model) RegisterChild(child GoStructItem) {
	m.sortedChilds = append(m.sortedChilds, child)
}

func (a *Assembly) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type assembly Assembly
	err := d.DecodeElement((*assembly)(a), &start)
	context.RegisterChild(a)
	return err
}

func (f *Field) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type field Field
	err := d.DecodeElement((*field)(f), &start)
	context.RegisterChild(f)
	return err
}
