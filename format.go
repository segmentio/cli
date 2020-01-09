package cli

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"

	yaml "gopkg.in/yaml.v3"
)

// Printer is an interface implemented for high-level printing formats.
type Printer interface {
	Print(interface{})
}

// Flusher is an interface implemented by types that buffer content.
type Flusher interface {
	Flush()
}

// PrintFlusher is an interface implemented by printers that may buffer content
// until they are flushed.
type PrintFlusher interface {
	Printer
	Flusher
}

// Format returns a Printer which formats printed values.
//
// Typical usage looks like this:
//
//	p, err := cli.Format(config.Format, os.Stdout)
//	if err != nil {
//		return err
//	}
//	defer p.Flush()
//	...
//	p.Print(v1)
//	p.Print(v2)
//	p.Print(v3)
//
// The package supports three formats: text, json, and yaml. All formats
// einterpret the `json` struct tag to configure the names of the fields
// and the behavior of the formatting operation.
//
// The text format also interprets `fmt` tags as carrying the formatting
// string passed in calls to functions of the `fmt` package.
//
// If the format name is not supported, the function returns a usage error.
func Format(format string, output io.Writer) (PrintFlusher, error) {
	switch format {
	case "json":
		return newJsonFormat(output), nil
	case "yaml":
		return newYamlFormat(output), nil
	case "text":
		return newTextFormat(output), nil
	default:
		return nil, &Usage{Err: fmt.Errorf("unsupported output format: %q", format)}
	}
}

type jsonFormat struct{ *json.Encoder }

func newJsonFormat(w io.Writer) jsonFormat {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return jsonFormat{e}
}

func (p jsonFormat) Print(v interface{}) { p.Encode(v) }

func (p jsonFormat) Flush() {}

type yamlFormat struct{ *yaml.Encoder }

func newYamlFormat(w io.Writer) yamlFormat {
	return yamlFormat{yaml.NewEncoder(w)}
}

func (p yamlFormat) Print(v interface{}) {
	b, _ := json.Marshal(v)

	var x interface{}
	yaml.Unmarshal(b, &x)

	p.Encode(x)
}

func (p yamlFormat) Flush() { p.Close() }

type textFormat struct {
	w  io.Writer
	tw tabwriter.Writer
	tt reflect.Type // last type seen
}

func newTextFormat(w io.Writer) *textFormat {
	return &textFormat{w: w}
}

func (p *textFormat) Print(x interface{}) {
	switch x.(type) {
	case encoding.TextMarshaler, fmt.Formatter, fmt.Stringer, error:
		p.print(x)
		return
	}
	switch v := reflect.ValueOf(x); v.Kind() {
	case reflect.Struct:
		p.printStruct(v)
	case reflect.Slice:
		p.printSlice(v)
	case reflect.Map:
		p.printMap(v)
	default:
		p.print(x)
	}
}

func (p *textFormat) printStruct(v reflect.Value) {
	t := v.Type()

	if t != p.tt {
		p.reset(t)

		i := 0
		p.forEachStructFieldName(v, func(name string) {
			if i != 0 {
				io.WriteString(&p.tw, "\t")
			}
			io.WriteString(&p.tw, name)
			i++
		})

		io.WriteString(&p.tw, "\n")
	}

	i := 0
	p.forEachStructFieldValue(v, func(format string, value interface{}) {
		if i != 0 {
			io.WriteString(&p.tw, "\t")
		}
		io.WriteString(&p.tw, p.format(format, value))
		i++
	})

	io.WriteString(&p.tw, "\n")
}

func (p *textFormat) printSlice(v reflect.Value) {
	for i, n := 0, v.Len(); i < n; i++ {
		p.Print(v.Index(i).Interface())
	}
}

func (p *textFormat) printMap(v reflect.Value) {
	t := v.Type()

	if t != p.tt {
		p.reset(t)

		for i, k := range sortedMapKeys(v) {
			if i != 0 {
				io.WriteString(&p.tw, "\t")
			}
			io.WriteString(&p.tw, normalizeColumnName(p.format("%v", k.Interface())))
		}

		io.WriteString(&p.tw, "\n")
	}

	for i, k := range sortedMapKeys(v) {
		if i != 0 {
			io.WriteString(&p.tw, "\t")
		}
		io.WriteString(&p.tw, p.format("%v", v.MapIndex(k).Interface()))
	}

	io.WriteString(&p.tw, "\n")
}

func (p *textFormat) reset(t reflect.Type) {
	p.Flush()
	p.tw.Init(p.w, 0, 4, 2, ' ', tabwriter.DiscardEmptyColumns)
	p.tt = t
}

func (p *textFormat) Flush() {
	if p.tt != nil {
		p.tt = nil
		p.tw.Flush()
	}
}

func (p *textFormat) print(v interface{}) {
	p.Flush() // in case there is buffered content
	io.WriteString(p.w, p.format("%v\n", v))
}

func (p *textFormat) format(f string, v interface{}) string {
	switch m := v.(type) {
	case fmt.Formatter, fmt.Stringer, error:
		// Takes priority over encoding.TextMarshaler, handled by the call to
		// fmt.Sprintf below.
	case encoding.TextMarshaler:
		b, _ := m.MarshalText()
		return string(b)
	}
	return fmt.Sprintf(f, v)
}

func (p *textFormat) forEachStructFieldName(v reflect.Value, do func(string)) {
	p.forEachStructField(v, func(name, _ string, _ reflect.Value) { do(name) })
}

func (p *textFormat) forEachStructFieldValue(v reflect.Value, do func(string, interface{})) {
	p.forEachStructField(v, func(_, format string, value reflect.Value) {
		do(format, value.Interface())
	})
}

func (p *textFormat) forEachStructField(v reflect.Value, do func(string, string, reflect.Value)) {
	t := v.Type()
	n := t.NumField()

	for i := 0; i < n; i++ {
		f := t.Field(i)

		if f.PkgPath != "" { // unexported
			continue
		}

		if f.Anonymous {
			p.forEachStructField(v.Field(i), do)
			continue
		}

		name, hasName := f.Tag.Lookup("json")
		name = strings.Split(name, ",")[0]
		if name == "-" {
			continue
		}
		if !hasName {
			name = f.Name
		}

		format, hasFormat := f.Tag.Lookup("fmt")
		if !hasFormat {
			format = "%v"
		}

		do(normalizeColumnName(name), format, v.Field(i))
	}
}

func normalizeColumnName(name string) string {
	return strings.ReplaceAll(strings.ToUpper(snakecase(name)), "_", " ")
}

// FormatList returns a Printer which formats lists of printed values.
//
// Typical usage looks like this:
//
//	p, err := cli.FormatList(config.Format, os.Stdout)
//	if err != nil {
//		return err
//	}
//	defer p.Flush()
//	...
//	p.Print(v1)
//	p.Print(v2)
//	p.Print(v3)
//
// The package supports three formats: text, json, and yaml. All formats
// einterpret the `json` struct tag to configure the names of the fields
// and the behavior of the formatting operation.
//
// The text format also interprets `fmt` tags as carrying the formatting
// string passed in calls to functions of the `fmt` package.
//
// If the format name is not supported, the function returns a usage error.
func FormatList(format string, output io.Writer) (PrintFlusher, error) {
	switch format {
	case "json":
		return newJsonFormatList(output), nil
	case "yaml":
		return newYamlFormatList(output), nil
	case "text":
		return newTextFormat(output), nil
	default:
		return nil, &Usage{Err: fmt.Errorf("unsupported output format: %q", format)}
	}
}

type jsonFormatList struct {
	writer io.Writer
	values []json.RawMessage
}

func newJsonFormatList(w io.Writer) *jsonFormatList {
	return &jsonFormatList{writer: w}
}

func (p *jsonFormatList) Print(v interface{}) {
	b, _ := json.Marshal(v)
	p.values = append(p.values, json.RawMessage(b))
}

func (p *jsonFormatList) Flush() {
	e := json.NewEncoder(p.writer)
	e.SetIndent("", "  ")
	e.Encode(p.values)
	p.values = nil
}

type yamlFormatList struct {
	writer io.Writer
	buffer bytes.Buffer
	enc    *json.Encoder
	dec    *json.Decoder
	values []interface{}
}

func newYamlFormatList(w io.Writer) *yamlFormatList {
	f := &yamlFormatList{writer: w}
	f.enc = json.NewEncoder(&f.buffer)
	f.dec = json.NewDecoder(&f.buffer)
	return f
}

func (p *yamlFormatList) Print(v interface{}) {
	var value interface{}
	p.enc.Encode(v)
	p.dec.Decode(&value)
	p.values = append(p.values, value)
}

func (p *yamlFormatList) Flush() {
	e := yaml.NewEncoder(p.writer)
	e.SetIndent(2)
	e.Encode(p.values)
	e.Close()
	p.values = nil
}
