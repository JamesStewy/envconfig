package envconfig

import (
	"encoding/base64"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrUnexportedField is the error returned by the Init* functions when a field of the config struct is not exported and the option AllowUnexported is not used.
	ErrUnexportedField = errors.New("envconfig: unexported field")
	// ErrNotAPointer is the error returned by the Init* functions when the configuration object is not a pointer.
	ErrNotAPointer = errors.New("envconfig: value is not a pointer")
	// ErrInvalidValueKind is the error returned by the Init* functions when the configuration object is not a struct.
	ErrInvalidValueKind = errors.New("envconfig: invalid value kind, only works on structs")
)

// ConfInfo stores information about a configuration struct.
type ConfInfo []*Field

func (cinfo *ConfInfo) append(fld *Field) {
	*cinfo = append(*cinfo, fld)
}

// Read reads the configuration from environment variables and populates the conf object.
func (cinfo *ConfInfo) Read() error {
	for _, fld := range *cinfo {
		if err := fld.setValue(); err != nil {
			return err
		}
	}
	return nil
}

type context struct {
	config          *ConfInfo
	name            fieldName
	optional        bool
	allowUnexported bool
}

// Unmarshaler is the interface implemented by objects that can unmarshal a environment variable string of themselves.
type Unmarshaler interface {
	Unmarshal(s string) error
}

// Options is used to customize the behavior of envconfig. Use it with InitWithOptions.
type Options struct {
	// Prefix allows specifying a prefix for each key.
	Prefix string

	// AllOptional determines whether to not throw errors by default for any key
	// that is not found. AllOptional=true means errors will not be thrown.
	AllOptional bool

	// AllowUnexported allows unexported fields to be present in the passed config.
	AllowUnexported bool
}

// Init reads the configuration from environment variables and populates the conf object.
// conf must be a pointer
func Init(conf interface{}) error {
	return InitWithOptions(conf, Options{})
}

// InitWithPrefix reads the configuration from environment variables and populates the conf object.
// conf must be a pointer.
// Each key read will be prefixed with the prefix string.
func InitWithPrefix(conf interface{}, prefix string) error {
	return InitWithOptions(conf, Options{Prefix: prefix})
}

// InitWithOptions reads the configuration from environment variables and populates the conf object.
// conf must be a pointer.
func InitWithOptions(conf interface{}, opts Options) error {
	cinfo, err := ParseWithOptions(conf, opts)
	if err != nil {
		return err
	}
	return cinfo.Read()
}

// Parse returns a ConfInfo object and sets up the conf object to be populated with Read().
// conf must be a pointer
func Parse(conf interface{}) (*ConfInfo, error) {
	return ParseWithOptions(conf, Options{})
}

// ParseWithPrefix returns a ConfInfo object and sets up the conf object to be populated with Read().
// conf must be a pointer.
// Each key will be prefixed with the prefix string.
func ParseWithPrefix(conf interface{}, prefix string) (*ConfInfo, error) {
	return ParseWithOptions(conf, Options{Prefix: prefix})
}

// ParseWithOptions returns a ConfInfo object and sets up the conf object to be populated with Read().
// conf must be a pointer.
func ParseWithOptions(conf interface{}, opts Options) (*ConfInfo, error) {
	value := reflect.ValueOf(conf)
	if value.Kind() != reflect.Ptr {
		return nil, ErrNotAPointer
	}

	elem := value.Elem()

	for elem.Kind() == reflect.Ptr {
		if elem.IsNil() {
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		elem = elem.Elem()
	}

	if elem.Kind() != reflect.Struct {
		return nil, ErrInvalidValueKind
	}

	name := fieldName{}
	if opts.Prefix != "" {
		name = name.Append(opts.Prefix)
	}

	cinfo := &ConfInfo{}
	return cinfo, readStruct(elem, &context{
		config:          cinfo,
		name:            name,
		optional:        opts.AllOptional,
		allowUnexported: opts.AllowUnexported,
	})
}

type tag struct {
	customName string
	optional   bool
	skip       bool
	defaultVal string
	note       string
}

func parseTag(s string) *tag {
	var t tag

	escape := false
	tokens := []string{""}
	for _, r := range s {
		if !escape {
			switch r {
			case '\\':
				escape = true
				continue
			case ',':
				tokens = append(tokens, "")
				continue
			}
		}
		escape = false
		tokens[len(tokens)-1] += string(r)
	}

	for _, v := range tokens {
		switch {
		case v == "-":
			t.skip = true
		case v == "optional":
			t.optional = true
		case strings.HasPrefix(v, "default="):
			t.defaultVal = strings.TrimPrefix(v, "default=")
		case strings.HasPrefix(v, "note="):
			t.note = strings.TrimPrefix(v, "note=")
		default:
			t.customName = v
		}
	}

	return &t
}

func readStruct(value reflect.Value, ctx *context) (err error) {
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		name := value.Type().Field(i).Name

		tag := parseTag(value.Type().Field(i).Tag.Get("envconfig"))
		if tag.skip || !field.CanSet() {
			if !field.CanSet() && !ctx.allowUnexported {
				return ErrUnexportedField
			}
			continue
		}

	doRead:
		switch field.Kind() {
		case reflect.Ptr:
			// it's a pointer, create a new value and restart the switch
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
			goto doRead
		case reflect.Struct:
			err = readStruct(field, &context{
				config:          ctx.config,
				name:            ctx.name.Append(name),
				optional:        ctx.optional || tag.optional,
				allowUnexported: ctx.allowUnexported,
			})
		default:
			ctx.config.append(&Field{
				name:            ctx.name.Append(name),
				value:           field,
				customName:      tag.customName,
				defaultVal:      tag.defaultVal,
				note:            tag.note,
				optional:        ctx.optional || tag.optional,
				allowUnexported: ctx.allowUnexported,
			})
		}

		if err != nil {
			return err
		}
	}

	return err
}

var (
	durationType    = reflect.TypeOf(new(time.Duration)).Elem()
	unmarshalerType = reflect.TypeOf(new(Unmarshaler)).Elem()
)

func isDurationField(t reflect.Type) bool {
	return t.AssignableTo(durationType)
}

func isUnmarshaler(t reflect.Type) bool {
	return t.Implements(unmarshalerType) || reflect.PtrTo(t).Implements(unmarshalerType)
}

func parseWithUnmarshaler(v reflect.Value, str string) error {
	var u = v.Addr().Interface().(Unmarshaler)
	return u.Unmarshal(str)
}

func parseDuration(v reflect.Value, str string) error {
	d, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	v.SetInt(int64(d))

	return nil
}

func parseBoolValue(v reflect.Value, str string) error {
	val, err := strconv.ParseBool(str)
	if err != nil {
		return err
	}
	v.SetBool(val)

	return nil
}

func parseIntValue(v reflect.Value, str string) error {
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	v.SetInt(val)

	return nil
}

func parseUintValue(v reflect.Value, str string) error {
	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	v.SetUint(val)

	return nil
}

func parseFloatValue(v reflect.Value, str string) error {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	v.SetFloat(val)

	return nil
}

func parseBytesValue(v reflect.Value, str string) error {
	val, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	v.SetBytes(val)

	return nil
}
