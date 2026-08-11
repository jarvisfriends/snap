// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package exui

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"gopkg.in/yaml.v3"
)

// outputFormat selects how a selection example prints its confirmed result
// on stdout. "pretty" is the human default; the structured forms exist so
// scripts can consume the same tools without parsing prose.
var outputFormat = flag.String("output", "pretty",
	`result format: pretty, values, json, yaml, or xml`)

// Field is one named result value, in output order (pretty/values/xml keep
// this order; json and yaml marshal maps, whose keys the encoders sort).
type Field struct{ Key, Value string }

// F builds a Field: exui.F("date", "2026-07-12").
func F(key, value string) Field { return Field{Key: key, Value: value} }

// fieldsXML adapts a field list to encoding/xml, emitting each field as an
// element named after its key inside <result>.
type fieldsXML []Field

// MarshalXML implements xml.Marshaler.
func (f fieldsXML) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "result"
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, fl := range f {
		el := xml.StartElement{Name: xml.Name{Local: fl.Key}}
		if err := e.EncodeElement(fl.Value, el); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// FinishFields ends a value-producing example and never returns. With
// ok=false nothing is printed and the exit code is 1 (unchanged contract).
// With ok=true the fields render in the format chosen by --output:
//
//	pretty   date:  2026-07-12         (aligned when multiple fields)
//	values   2026-07-12                (bare values, one per line)
//	json     {"date":"2026-07-12"}     (encoding/json)
//	yaml     date: 2026-07-12          (gopkg.in/yaml.v3)
//	xml      <result><date>2026-07-12</date></result>  (encoding/xml)
func FinishFields(ok bool, fields ...Field) {
	if !ok {
		os.Exit(1)
	}
	out, err := renderFields(*outputFormat, fields)
	if err != nil {
		Fatal(err)
	}
	fmt.Print(out)
	os.Exit(0)
}

// renderFields is FinishFields' pure core, split out for tests.
func renderFields(format string, fields []Field) (string, error) {
	switch format {
	case "values":
		var b strings.Builder
		for _, f := range fields {
			b.WriteString(f.Value)
			b.WriteByte('\n')
		}
		return b.String(), nil
	case "json":
		out, err := json.Marshal(fieldMap(fields))
		return string(out) + "\n", err
	case "yaml":
		out, err := yaml.Marshal(fieldMap(fields))
		return string(out), err
	case "xml":
		out, err := xml.Marshal(fieldsXML(fields))
		return string(out) + "\n", err
	default: // pretty
		w := 0
		for _, f := range fields {
			w = max(w, lipgloss.Width(f.Key))
		}
		var b strings.Builder
		for _, f := range fields {
			pad := strings.Repeat(" ", w-lipgloss.Width(f.Key)+2)
			b.WriteString(f.Key + ":" + pad + f.Value + "\n")
		}
		return b.String(), nil
	}
}

// fieldMap converts fields for the map-marshaling encoders (both sort keys,
// so output stays deterministic).
func fieldMap(fields []Field) map[string]string {
	m := make(map[string]string, len(fields))
	for _, f := range fields {
		m[f.Key] = f.Value
	}
	return m
}
