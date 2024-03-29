package cli

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const uintSize = 32 << (^uint(0) >> 32 & 1)

type decodeFunc func(reflect.Value, []string) error

// structDecoder is a map of `structFieldDecoder` instances for all of the
// fields in a struct, which is expected to represent the options for a CLI
// command. Each map key is the final flag specified for the field, and each
// corresponding value is the decoder for that field.
type structDecoder map[string]structFieldDecoder

func (s structDecoder) decode(value reflect.Value, options map[string][]string) error {
	for option, values := range options {
		f := s[option]
		v := value.FieldByIndex(f.index)

		switch err := f.decode(v, values).(type) {
		case nil:
		case *Usage:
			err.Err = fmt.Errorf("decoding %q: %w", option, err.Err)
			return err
		default:
			return &Usage{Err: fmt.Errorf("decoding %q: %w", option, err)}
		}
	}
	return nil
}

// structFieldDecoder collects together a `structField` with a decode function
// appropriate for the field type.
type structFieldDecoder struct {
	index   []int
	flags   []string
	envvars []string
	help    string
	argtyp  string
	defval  string
	hidden  bool
	boolean bool
	slice   bool
	decode  decodeFunc
}

// makeStructDecoder creates a parser and struct decoder based on the given
// struct type, which is expected to represent the options for a command. The
// decoder automatically includes an additional "--help" Boolean decoder.
//
// The returned parser is programmed with flag alternatives (aliases) and
// additional metadata so that a command line can be parsed correctly.
//
// The final argument is the value of the "help" tag for the struct field named
// "_", if it exists.
func makeStructDecoder(t reflect.Type) (parser, structDecoder, string) {
	p := makeParser()
	s := structDecoder{
		"--help": structFieldDecoder{
			index:   nil,
			flags:   []string{"-h", "--help"},
			help:    "Show this help message",
			boolean: true,
			decode:  decodeBool,
		},
	}

	forEachStructField(t, nil, func(field structField) {
		boolean := field.isBoolean()
		decoder := makeStructFieldDecoder(field)

		for i, flag := range field.flags {
			flag = strings.TrimSpace(flag)
			if _, exists := p.aliases[flag]; exists {
				panic("repeated flag in configuration struct: " + flag)
			}

			if _, exists := p.options[flag]; exists {
				panic("repeated flag in configuration struct: " + flag)
			}

			if n := len(field.flags) - 1; i < n {
				p.aliases[flag] = strings.TrimSpace(field.flags[n])
			} else {
				p.options[flag] = option{boolean: boolean}
				s[flag] = decoder
			}
		}
	})

	if helpField, ok := t.FieldByName("_"); ok {
		return p, s, helpField.Tag.Get("help")
	}

	return p, s, ""
}

// makeStructFieldDecoder creates a decoder for a struct field, containing a
// decode function appropriate for the field type.
func makeStructFieldDecoder(f structField) structFieldDecoder {
	var decode decodeFunc
	switch f.typ.Kind() {
	case reflect.Slice:
		decode = makeSliceDecoder(f.typ)
	default:
		decode = makeValueDecoder(f.typ)
	}
	if decode == nil {
		panic("makeFieldDecoder called with unsupported type: " + f.typ.String())
	}
	return structFieldDecoder{
		index:   f.index,
		flags:   f.flags,
		envvars: f.envvars,
		help:    f.help,
		defval:  f.defval,
		hidden:  f.hidden,
		boolean: f.isBoolean(),
		slice:   f.isSlice(),
		decode:  decode,
		argtyp:  typeNameOf(f.typ),
	}
}

// forEachStructField executes the provided function for every field in a type,
// in index order. The index argument tracks the indices needed to retrieve a
// field when calling `reflect.Value.FieldByIndex` on the top struct type; pass
// `nil` when calling for that top type.
//
// The per-field function is not called for unexported fields or fields named
// "_".
//
// Most struct field attributes are derived from the field's tags. In
// particular, the value of `envvars` is computed from the `env` tag:
// * If the tag is empty, `envvars` is a list of all long options, converted to
//   environment variable name equivalents.
// * If the tag is `-`, `envvars` is `nil`.
// * Otherwise, `envvars` is only the single tag value.
func forEachStructField(t reflect.Type, index []int, do func(structField)) {
	for i, n := 0, t.NumField(); i < n; i++ {
		f := t.Field(i)

		fieldIndex := make([]int, 0, len(index)+len(f.Index))
		fieldIndex = append(fieldIndex, index...)
		fieldIndex = append(fieldIndex, f.Index...)

		if f.Anonymous {
			forEachStructField(f.Type, fieldIndex, do)
			continue
		}

		if f.PkgPath != "" { // unexported
			continue
		}

		if f.Name == "_" {
			continue
		}

		if !isSupportedFieldType(f.Type) {
			panic("configuration struct contains unsupported field type: " + f.Name + " " + f.Type.String())
		}

		var splitFlags = strings.Split(f.Tag.Get("flag"), ",")
		flags := make([]string, len(splitFlags))
		for i := range splitFlags {
			flags[i] = strings.TrimSpace(splitFlags[i])
		}
		var envvars []string

		switch env := f.Tag.Get("env"); env {
		case "":
			for _, f := range flags {
				if isLongFlag(f) {
					envvars = append(envvars, envNameOf(f))
				}
			}
		case "-":
			envvars = nil
		default:
			envvars = append(envvars, env)
		}

		hidden, err := strconv.ParseBool(f.Tag.Get("hidden"))
		if err != nil {
			hidden = false
		}

		do(structField{
			typ:     f.Type,
			index:   fieldIndex,
			envvars: envvars,
			flags:   flags,
			help:    f.Tag.Get("help"),
			defval:  f.Tag.Get("default"),
			hidden:  hidden,
		})
	}
}

// envNameOf gets a environment variable name that is equivalent to the given
// flag.
func envNameOf(s string) string {
	return strings.ToUpper(snakecase(flagNameOf(s)))
}

// flagNameOf gets the name of the given flag without prefixed hyphens.
func flagNameOf(s string) string {
	switch {
	case strings.HasPrefix(s, "--"):
		return strings.TrimPrefix(s, "--")
	case strings.HasPrefix(s, "-"):
		return strings.TrimPrefix(s, "-")
	default:
		return s
	}
}

// makeValueDecoder returns a decode function for values of the given type, or
// nil if the type isn't supported.
func makeValueDecoder(t reflect.Type) decodeFunc {
	switch t {
	case durationType:
		return decodeDuration
	case timeType:
		return decodeTime
	}
	switch {
	case isTextUnmarshaler(t):
		return decodeTextUnmarshaler
	case isBinaryUnmarshaler(t):
		return decodeBinaryUnmarshaler
	}
	switch t.Kind() {
	case reflect.Bool:
		return decodeBool
	case reflect.Int:
		return decodeInt
	case reflect.Int8:
		return decodeInt8
	case reflect.Int16:
		return decodeInt16
	case reflect.Int32:
		return decodeInt32
	case reflect.Int64:
		return decodeInt64
	case reflect.Uint:
		return decodeUint
	case reflect.Uint8:
		return decodeUint8
	case reflect.Uint16:
		return decodeUint16
	case reflect.Uint32:
		return decodeUint32
	case reflect.Uint64:
		return decodeUint64
	case reflect.Uintptr:
		return decodeUintptr
	case reflect.Float32:
		return decodeFloat32
	case reflect.Float64:
		return decodeFloat64
	case reflect.String:
		return decodeString
	}
	return nil
}

func makeSliceDecoder(t reflect.Type) decodeFunc {
	if isTextUnmarshaler(t) {
		return decodeTextUnmarshaler
	}
	if isBinaryUnmarshaler(t) {
		return decodeBinaryUnmarshaler
	}
	e := t.Elem()
	f := makeValueDecoder(e)
	z := reflect.Zero(e)
	return func(v reflect.Value, a []string) error {
		for i := 0; i < len(a); i++ {
			v.Set(reflect.Append(v, z))
			if err := f(v.Index(v.Len()-1), a[i:i+1]); err != nil {
				return err
			}
		}
		return nil
	}
}

func assertArgumentCount(a []string, n int) error {
	switch {
	case len(a) < n:
		return &Usage{Err: fmt.Errorf("not enough arguments, expected %d but got %d", n, len(a))}
	case len(a) > n:
		return &Usage{Err: fmt.Errorf("too many arguments, expected %d but got %d", n, len(a))}
	}
	return nil
}

func decodeBool(v reflect.Value, a []string) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}
	x, err := strconv.ParseBool(a[0])
	if err != nil {
		return err
	}
	v.SetBool(x)
	return nil
}

func decodeInt(v reflect.Value, a []string) error {
	return decodeIntSize(v, a, uintSize)
}

func decodeInt8(v reflect.Value, a []string) error {
	return decodeIntSize(v, a, 8)
}

func decodeInt16(v reflect.Value, a []string) error {
	return decodeIntSize(v, a, 16)
}

func decodeInt32(v reflect.Value, a []string) error {
	return decodeIntSize(v, a, 32)
}

func decodeInt64(v reflect.Value, a []string) error {
	return decodeIntSize(v, a, 64)
}

func decodeIntSize(v reflect.Value, a []string, bits int) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}
	x, err := strconv.ParseInt(a[0], 0, bits)
	if err != nil {
		return err
	}
	v.SetInt(x)
	return nil
}

func decodeUint(v reflect.Value, a []string) error {
	return decodeUintSize(v, a, uintSize)
}

func decodeUint8(v reflect.Value, a []string) error {
	return decodeUintSize(v, a, 8)
}

func decodeUint16(v reflect.Value, a []string) error {
	return decodeUintSize(v, a, 16)
}

func decodeUint32(v reflect.Value, a []string) error {
	return decodeUintSize(v, a, 32)
}

func decodeUint64(v reflect.Value, a []string) error {
	return decodeUintSize(v, a, 64)
}

func decodeUintptr(v reflect.Value, a []string) error {
	return decodeUintSize(v, a, 64)
}

func decodeUintSize(v reflect.Value, a []string, bits int) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}
	x, err := strconv.ParseUint(a[0], 0, bits)
	if err != nil {
		return err
	}
	v.SetUint(x)
	return nil
}

func decodeFloat32(v reflect.Value, a []string) error {
	return decodeFloat(v, a, 32)
}

func decodeFloat64(v reflect.Value, a []string) error {
	return decodeFloat(v, a, 64)
}

func decodeFloat(v reflect.Value, a []string, bits int) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}
	x, err := strconv.ParseFloat(a[0], bits)
	if err != nil {
		return err
	}
	v.SetFloat(x)
	return nil
}

func decodeDuration(v reflect.Value, a []string) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}
	x, err := time.ParseDuration(a[0])
	if err != nil {
		return err
	}
	v.SetInt(int64(x))
	return nil
}

func decodeTime(v reflect.Value, a []string) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}

	for _, format := range []string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	} {
		t, err := time.Parse(format, a[0])
		if err == nil {
			v.Set(reflect.ValueOf(t))
			return nil
		}
	}

	return fmt.Errorf("malformed time value: %q", a[0])
}

func decodeString(v reflect.Value, a []string) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}
	v.SetString(a[0])
	return nil
}

func decodeTextUnmarshaler(v reflect.Value, a []string) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}
	b := []byte(a[0])
	u := v.Addr().Interface().(encoding.TextUnmarshaler)
	return u.UnmarshalText(b)
}

func decodeBinaryUnmarshaler(v reflect.Value, a []string) error {
	if err := assertArgumentCount(a, 1); err != nil {
		return err
	}
	b := []byte(a[0])
	u := v.Addr().Interface().(encoding.BinaryUnmarshaler)
	return u.UnmarshalBinary(b)
}

// structField represents a single field in a struct, with its tag values parsed
// out.
type structField struct {
	// typ is the field type.
	typ     reflect.Type
	// index is the index sequence for retrieving this field from its top-level
	// struct using `Type.FieldByIndex`.
	index   []int
	// flags is the list of values for the field's `flag` tag.
	flags   []string
	// envvars is the list of environment variable names calculated from either
	// the field's `flag` tag or its `env` tag.
	envvars []string
	// help is the value of the field's `help` tag.
	help    string
	// defval is the value of the field's `default` tag.
	defval  string
	// hidden is the value of the field's `hidden` tag.
	hidden  bool
}

func (f structField) isBoolean() bool { return f.typ.Kind() == reflect.Bool }
func (f structField) isSlice() bool   { return f.typ.Kind() == reflect.Slice }

var (
	intType               = reflect.TypeOf(0)
	durationType          = reflect.TypeOf(time.Duration(0))
	timeType              = reflect.TypeOf(time.Time{})
	emptyType             = reflect.TypeOf(struct{}{})
	errorType             = reflect.TypeOf((*error)(nil)).Elem()
	textUnmarshalerType   = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	binaryUnmarshalerType = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
)

func isSupportedFieldType(t reflect.Type) bool {
	switch t {
	case durationType, timeType:
		return true
	}
	switch {
	case isTextUnmarshaler(t), isBinaryUnmarshaler(t):
		return true
	}
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return true
	case reflect.Slice:
		return t.Elem().Kind() != reflect.Slice && isSupportedFieldType(t.Elem())
	}
	return false
}

func isTextUnmarshaler(t reflect.Type) bool {
	return reflect.PtrTo(t).Implements(textUnmarshalerType)
}

func isBinaryUnmarshaler(t reflect.Type) bool {
	return reflect.PtrTo(t).Implements(binaryUnmarshalerType)
}

func typeNameOf(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return ""
	case reflect.Slice:
		return typeNameOf(t.Elem()) + "..."
	}
	s := t.String()
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		s = s[i+1:]
	}
	return strings.ReplaceAll(strings.ToLower(snakecase(s)), "_", "-")
}
