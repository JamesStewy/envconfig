package envconfig

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// Field represents a single field in a configuration struct.
type Field struct {
	name            fieldName
	value           reflect.Value
	strValue        string
	customName      string
	defaultVal      string
	note            string
	optional        bool
	allowUnexported bool
}

// Name returns the full name of the field.
func (fld *Field) Name() string {
	return fld.name.String()
}

// Value returns the environment variable value used to set this field.
// Value will return an empty string until Read() is called on the ConfInfo object containing this field.
func (fld *Field) Value() string {
	return fld.strValue
}

// Default returns the default value for this field.
func (fld *Field) Default() string {
	return fld.defaultVal
}

// Note returns any notes set for this field.
func (fld *Field) Note() string {
	return fld.note
}

// Optional returns whether or not this field is optional.
func (fld *Field) Optional() bool {
	return fld.optional
}

// Keys returns a slice containing all environment keys that will be tried when populating this field.
func (fld *Field) Keys() []string {
	if fld.customName != "" {
		return []string{fld.customName}
	}
	return fld.name.Keys()
}

func (fld *Field) setValue() (err error) {
	return fld.setField(fld.value)
}

var byteSliceType = reflect.TypeOf([]byte(nil))

func (fld *Field) setField(value reflect.Value) (err error) {
	str, err := fld.readValue()
	if err != nil {
		return err
	}

	if len(str) == 0 && fld.optional {
		return nil
	}

	fld.strValue = str

	isSliceNotUnmarshaler := value.Kind() == reflect.Slice && !isUnmarshaler(value.Type())
	switch {
	case isSliceNotUnmarshaler && value.Type() == byteSliceType:
		return parseBytesValue(value, str)

	case isSliceNotUnmarshaler:
		return fld.setSliceField(value, str)

	default:
		return fld.parseValue(value, str)
	}
}

func (fld *Field) setSliceField(value reflect.Value, str string) error {
	elType := value.Type().Elem()
	tnz := newSliceTokenizer(str)

	slice := reflect.MakeSlice(value.Type(), value.Len(), value.Cap())

	for tnz.scan() {
		token := tnz.text()

		el := reflect.New(elType).Elem()

		if err := fld.parseValue(el, token); err != nil {
			return err
		}

		slice = reflect.Append(slice, el)
	}

	value.Set(slice)

	return tnz.Err()
}

func (fld *Field) parseValue(v reflect.Value, str string) (err error) {
	vtype := v.Type()

	// Special case when the type is a map: we need to make the map
	switch vtype.Kind() {
	case reflect.Map:
		v.Set(reflect.MakeMap(vtype))
	}

	// Special case for Unmarshaler
	if isUnmarshaler(vtype) {
		return parseWithUnmarshaler(v, str)
	}

	// Special case for time.Duration
	if isDurationField(vtype) {
		return parseDuration(v, str)
	}

	kind := vtype.Kind()
	switch kind {
	case reflect.Bool:
		err = parseBoolValue(v, str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = parseIntValue(v, str)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = parseUintValue(v, str)
	case reflect.Float32, reflect.Float64:
		err = parseFloatValue(v, str)
	case reflect.Ptr:
		v.Set(reflect.New(vtype.Elem()))
		return fld.parseValue(v.Elem(), str)
	case reflect.String:
		v.SetString(str)
	case reflect.Struct:
		err = fld.parseStruct(v, str)
	default:
		return fmt.Errorf("envconfig: kind %v not supported", kind)
	}

	return
}

// NOTE(vincent): this is only called when parsing structs inside a slice.
func (fld *Field) parseStruct(value reflect.Value, token string) error {
	tokens := strings.Split(token[1:len(token)-1], ",")
	if len(tokens) != value.NumField() {
		return fmt.Errorf("envconfig: struct token has %d fields but struct has %d", len(tokens), value.NumField())
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		t := tokens[i]

		if err := fld.parseValue(field, t); err != nil {
			return err
		}
	}

	return nil
}

func (fld *Field) readValue() (string, error) {
	keys := fld.Keys()

	var str string

	for _, key := range keys {
		str = os.Getenv(key)
		if str != "" {
			break
		}
	}

	if str != "" {
		return str, nil
	}

	if fld.defaultVal != "" {
		return fld.defaultVal, nil
	}

	if fld.optional {
		return "", nil
	}

	return "", fmt.Errorf("envconfig: keys %s not found", strings.Join(keys, ", "))
}

type fieldName []string

func (name fieldName) String() string {
	return strings.Join(name, ".")
}

func (name fieldName) Append(newfield string) fieldName {
	tmp := make(fieldName, len(name)+1)
	copy(tmp, name)
	tmp[len(name)] = newfield
	return tmp
}

func (name fieldName) Keys() (res []string) {
	var buf bytes.Buffer  // this is the buffer where we put extra underscores on "word" boundaries
	var buf2 bytes.Buffer // this is the buffer with the standard naming scheme

	for j, part := range name {
		if j > 0 {
			buf.WriteRune('_')
			buf2.WriteRune('_')
		}

		n := []rune(part)
		for i, r := range part {
			prevOrNextLower := i+1 < len(n) && i-1 > 0 && (unicode.IsLower(n[i+1]) || unicode.IsLower(n[i-1]))
			if i > 0 && unicode.IsUpper(r) && prevOrNextLower {
				buf.WriteRune('_')
			}

			buf.WriteRune(r)
			buf2.WriteRune(r)
		}
	}

	tmp := make(map[string]struct{})
	tmp[strings.ToLower(buf.String())] = struct{}{}
	tmp[strings.ToUpper(buf.String())] = struct{}{}
	tmp[strings.ToLower(buf2.String())] = struct{}{}
	tmp[strings.ToUpper(buf2.String())] = struct{}{}

	for k := range tmp {
		res = append(res, k)
	}

	sort.Strings(res)
	return
}
