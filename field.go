package statusx

import (
	"bytes"
	"sort"
)

func NewErrorField(in string, field string, msg string) *ErrorField {
	return &ErrorField{
		In:    in,
		Field: field,
		Msg:   msg,
	}
}

type ErrorField struct {
	Field string `json:"field" xml:"field"` // Field path: prop.slice[2].a
	Msg   string `json:"msg"   xml:"msg"`   // Msg message
	In    string `json:"in"    xml:"in"`    // In location eq. body, query, header, path, formData
}

func (s ErrorField) String() string {
	return s.Field + " in " + s.In + " - " + s.Msg
}

type ErrorFields []*ErrorField

func (fs ErrorFields) String() string {
	if len(fs) == 0 {
		return ""
	}

	sort.Slice(fs, func(i, j int) bool {
		return fs[i].Field < fs[j].Field
	})

	buf := &bytes.Buffer{}
	buf.WriteString("<")
	for i, f := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(f.String())
	}
	buf.WriteString(">")
	return buf.String()
}
