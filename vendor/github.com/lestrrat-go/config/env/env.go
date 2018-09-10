package env

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/camelcase"
	"github.com/lestrrat-go/config/env/internal/structtag"
	pdebug "github.com/lestrrat-go/pdebug"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func Unmarshal(v interface{}) error {
	return NewDecoder(System).Decode(v)
}

func NewDecoder(src Source) *Decoder {
	return &Decoder{
		prefix: "",
		src:    src,
		sep:    "_",
	}
}

func (d *Decoder) Prefix(s string) *Decoder {
	d.prefix = s
	return d
}

func (d *Decoder) Decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	// v must be a pointer, otherwise we can't set to it
	if rv.Kind() != reflect.Ptr {
		return errors.New(`value must be a pointer`)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, prefixKey{}, d.prefix)
	ctx = context.WithValue(ctx, separatorKey{}, d.sep)

	// rv is where we *STORE* the value
	_, err := decodeValue(ctx, rv, d.src)
	return err
}

func getEnvName(t reflect.StructField) string {
	// By default, use the field name
	name := t.Name

	// Look for env, or envconfig tags
	for _, n := range []string{"env", "envconfig"} {
		tag, ok := structtag.Lookup(t.Tag, n)
		if !ok || len(tag) <= 0 {
			continue
		}

		// TODO: we may need to parse this tag later?
		name = tag
	}

	// Now, this may need tweaking.
	if ok, err := structtag.BoolValue(t.Tag, "split_words"); err == nil && ok {
		// Convert `CamelCase` into `CAMEL_CASE`
		name = strings.ToUpper(strings.Join(camelcase.Split(name), "_"))
	}

	return strings.ToUpper(name)
}

func getSeparator(ctx context.Context) string {
	return ctx.Value(separatorKey{}).(string)
}

func getPrefix(ctx context.Context) string {
	p := ctx.Value(prefixKey{})
	if p == nil {
		return ""
	}
	return p.(string)
}

func storePrefix(ctx context.Context, n string) context.Context {
	return context.WithValue(ctx, prefixKey{}, n)
}

func addPrefix(ctx context.Context, n string) string {
	if p := getPrefix(ctx); len(p) > 0 {
		n = p + getSeparator(ctx) + n
	}
	return n
}

var zeroval reflect.Value

func convertValue(t reflect.Type, s string) (reflect.Value, error) {
	if convertCustom(t) {
		return convertCustomValue(t, s)
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(s), nil
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return zeroval, errors.Wrap(err, `failed to parse as boolean`)
		}
		return reflect.ValueOf(b), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 0, t.Bits())
		if err != nil {
			return zeroval, errors.Wrap(err, `failed to parse as int`)
		}
		rv := reflect.New(t).Elem()
		rv.SetInt(i)
		return rv, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, t.Bits())
		if err != nil {
			return zeroval, errors.Wrap(err, `failed to parse as uint`)
		}
		rv := reflect.New(t).Elem()
		rv.SetUint(i)
		return rv, nil
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(s, t.Bits())
		if err != nil {
			return zeroval, errors.Wrap(err, `failed to parse as float`)
		}
		rv := reflect.New(t).Elem()
		rv.SetFloat(i)
		return rv, nil
	case reflect.Slice:
		elems := strings.Split(s, ",")
		rv := reflect.MakeSlice(t, 0, len(elems))
		for i, elem := range elems {
			ev, err := convertValue(t.Elem(), elem)
			if err != nil {
				return zeroval, errors.Wrapf(err, `failed to convert slice element %d`, i)
			}
			rv = reflect.Append(rv, ev)
		}
		return rv, nil
	case reflect.Map:
		elems := strings.Split(s, ",")
		rv := reflect.MakeMap(t)
		for _, elem := range elems {
			i := strings.IndexByte(elem, '=')
			if i < 1 || len(elem)-1 <= i {
				return zeroval, errors.New(`invalid map element syntax`)
			}
			k, err := convertValue(t.Key(), elem[:i])
			if err != nil {
				return zeroval, errors.Wrap(err, `failed to convert map key`)
			}
			v, err := convertValue(t.Elem(), elem[i+1:])
			if err != nil {
				return zeroval, errors.Wrap(err, `failed to convert map value`)
			}
			rv.SetMapIndex(k, v)
		}
		return rv, nil
	default:
		return zeroval, errors.Errorf(`unknown type for conversion: %s`, t)
	}
}

// TODO: convertCustom and convertCustomValue should really be one method
var ifUnmarshal = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

func convertCustom(t reflect.Type) bool {
	if t.Implements(ifUnmarshal) || reflect.PtrTo(t).Implements(ifUnmarshal) {
		return true
	}

	if t.PkgPath() != "time" {
		return false
	}

	if n := t.Name(); n != "Duration" && n != "Time" {
		return false
	}
	return true
}

func convertCustomValue(t reflect.Type, s string) (reflect.Value, error) {
	switch {
	case t.Implements(ifUnmarshal) || reflect.PtrTo(t).Implements(ifUnmarshal):
		rv := reflect.New(t) // Note: ptr to T
		rets := rv.MethodByName("UnmarshalEnv").Call([]reflect.Value{reflect.ValueOf(s)})
		if len(rets) == 0 {
			panic("did not get return value from calling UnmarshalEnv")
		}

		if !rets[0].IsNil() {
			return zeroval, errors.Wrap(rets[0].Interface().(error), "error calling UnmarshalEnv")
		}
		return rv.Elem(), nil
	case t.PkgPath() == "time" && t.Name() == "Time":
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return zeroval, errors.Wrap(err, `failed to parse time`)
		}
		return reflect.ValueOf(v), nil
	case t.PkgPath() == "time" && t.Name() == "Duration":
		v, err := time.ParseDuration(s)
		if err != nil {
			return zeroval, errors.Wrap(err, `failed to parse duration`)
		}
		return reflect.ValueOf(v), nil
	default:
		return zeroval, errors.Errorf(`unsupported struct type: %s`, t)
	}
}

func assignIfSuccessful(rv reflect.Value, cb func(reflect.Value) (ret bool, err error)) (assigned bool, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("assignIfSuccessful").BindError(&err)
		defer g.End()
	}
	if rv.Kind() == reflect.Interface {
		// Since we can't expect an implementation for interface,
		// nil sets to the value of a struct field even if environment variable is set.
		return false, nil
	}

	if rv.Kind() == reflect.Ptr {
		// We have a pointer. Does the thing point to anything?
		if rv.Elem().IsValid() {
			// Okay, the pointer does point to something. In this case, the
			// caller has already explicitly initialized the value, so we
			// should directly manipulate this value that we have
			rv = rv.Elem()
		} else {
			result := rv // save rv to a temporary variable

			// It doesn't. We need to create a value that we can muck with
			// and play with.
			// The `reflect.New()` call creates a new *pointer* to the element
			// type pointed by rv.
			// Since we don't want to deal with indirection, we dereference it
			// using `.Elem()` at the end. The result is assigned to rv, which
			// can now be manipulated and populated, but does not YET assign to
			// the original container (pointer)
			if pdebug.Enabled {
				pdebug.Printf("Creating new value of type %s", rv.Type().Elem())
			}
			rv = reflect.New(rv.Type().Elem()).Elem()

			defer func() {
				if err != nil {
					// if there was an error, there's nothing to do.
					if pdebug.Enabled {
						pdebug.Printf("Error in callback: %s", err)
					}
					return
				}

				// if this was not an error, we need to check if any value
				// has been assigned to rv.
				if !assigned {
					if pdebug.Enabled {
						pdebug.Printf("Assigned is false")
					}
					return
				}

				if pdebug.Enabled {
					pdebug.Printf("Setting value to result")
				}

				// Now that we know the value is valid, we assign
				result.Set(rv.Addr())
			}()
		}
	}

	return cb(rv)
}

func decodeValue(ctx context.Context, rv reflect.Value, src Source) (assigned bool, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("decodeValue").BindError(&err)
		defer g.End()
	}

	return assignIfSuccessful(rv, func(rv reflect.Value) (ret bool, err error) {
		if pdebug.Enabled {
			g := pdebug.Marker("deodeValue callback").BindError(&err)
			defer g.End()
		}

		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		switch rv.Kind() {
		case reflect.Struct:
			return decodeStructValue(ctx, rv, src)
		default:
			return false, errors.Errorf(`unknown kind: %s`, rv.Kind())
		}
	})
}

func decodeStructValue(ctx context.Context, rv reflect.Value, src Source) (assigned bool, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("decodeStructValue").BindError(&err)
		defer g.End()
	}

	if rv.Kind() != reflect.Struct {
		return false, errors.New(`expected struct kind`)
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if sf.PkgPath != "" {
			continue // can't handle unexported fields
		}

		if pdebug.Enabled {
			pdebug.Printf("Checking field '%s'", sf.Name)
		}

		ok, err := assignIfSuccessful(rv.Field(i), func(fv reflect.Value) (bool, error) {
			n := addPrefix(ctx, getEnvName(sf))

			if !convertCustom(fv.Type()) {
				sft := sf.Type
				if sft.Kind() == reflect.Interface {
					// Here isn't executed in normal case.
					return false, errors.New("interface is not decoded")
				}

				if sft.Kind() == reflect.Ptr {
					if pdebug.Enabled {
						pdebug.Printf("this is a pointer")
					}
					sft = sft.Elem()
				}

				switch sft.Kind() {
				case reflect.Interface:
					// Since we can't expect an implementation for interface,
					// pointer to nil interface sets to the value of a struct field
					// even if environment variable is set.
					return false, nil
				case reflect.Struct:
					// Lookee here! it's a struct. we first have to muck with the preix
					ok, err := decodeStructValue(storePrefix(ctx, n), fv, src)
					if err != nil {
						return false, errors.Wrap(err, `failed to decode struct value`)
					}
					return ok, nil
				}
				// default case, fallthrough
			}
			if pdebug.Enabled {
				pdebug.Printf("Looking up environment variable '%s'", n)
			}
			v, ok := src.LookupEnv(n)
			if !ok {
				if pdebug.Enabled {
					pdebug.Printf("Environment variable '%s' not found", n)
				}
				return false, nil
			}

			converted, err := convertValue(fv.Type(), v)
			if err != nil {
				return false, errors.Wrap(err, `failed to convert value`)
			}
			if pdebug.Enabled {
				pdebug.Printf("Conversion done, setting value")
			}
			fv.Set(converted)
			return true, nil
		})
		if err != nil {
			return false, err
		}
		if ok {
			if pdebug.Enabled {
				pdebug.Printf("Assgned = true for field '%s'", sf.Name)
			}
			assigned = true
		}
	}
	return assigned, nil
}
