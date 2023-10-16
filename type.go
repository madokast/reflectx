package reflectx

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"reflect"
	"strconv"
)

// Type is a serializable representation of a Go type.
//
// Field Kind is same as reflect.Type::Kind()
// Struct gives the complete info of the type if the Kind field cannot specify the type
type Type struct {
	Kind   string          `json:"kind,omitempty"`   // reflect.Type.Kind().String()
	Struct map[string]Type `json:"struct,omitempty"` // struct info of the type
	Extra  map[string]any  `json:"extra,omitempty"`  // extra info of the type. Such as array length
}

const (
	ElementType = "element_type"  // element type of Array Chan Pointer Slice
	ArrayLength = "array_length"  // length of Array
	KeyType     = "key_type"      // key type of Map
	ValueType   = "value_type"    // value type of Map
	NumIn       = "in_number"     // input parameter count of Func
	NumOut      = "out_number"    // output parameter count of Func
	In          = "in_"           // specify the index of in type
	Out         = "out_"          // specify the index of out type
	NumField    = "fields_number" // a struct type's field count
	Field       = "field_"        // specify the inedex of of field name
	ChanDir     = "chan_dir"      // dir of Chan
	Variadic    = "variadic"      // variadic of Func
)

var kindMap = map[string]reflect.Kind{}

// TypeOf returns the type of i
func TypeOf(i any) Type {
	goType := reflect.TypeOf(i)
	if goType == nil {
		panic("can not get type of nil")
	}
	return FromGoType(goType)
}

// FromGoType converts the std reflect.Type to Type
func FromGoType(goType reflect.Type) Type {
	kind := goType.Kind()
	xType := Type{Kind: kind.String(), Struct: map[string]Type{}, Extra: map[string]any{}}
	switch kind {
	case reflect.Invalid:
		panic("invalid kind of type")
	case reflect.Array:
		xType.Struct[ElementType] = FromGoType(goType.Elem())
		xType.Extra[ArrayLength] = goType.Len()
	case reflect.Chan:
		xType.Struct[ElementType] = FromGoType(goType.Elem())
		xType.Extra[ChanDir] = int(goType.ChanDir())
	case reflect.Func:
		xType.Extra[NumIn] = goType.NumIn()
		xType.Extra[NumOut] = goType.NumOut()
		for i := 0; i < goType.NumIn(); i++ {
			xType.Struct[In+strconv.Itoa(i)] = FromGoType(goType.In(i))
		}
		for i := 0; i < goType.NumOut(); i++ {
			xType.Struct[Out+strconv.Itoa(i)] = FromGoType(goType.Out(i))
		}
		xType.Extra[Variadic] = goType.IsVariadic()
	case reflect.Interface:
		panic("Interface type are not impl")
	case reflect.Map:
		xType.Struct[KeyType] = FromGoType(goType.Key())
		xType.Struct[ValueType] = FromGoType(goType.Elem())
	case reflect.Pointer:
		xType.Struct[ElementType] = FromGoType(goType.Elem())
	case reflect.Slice:
		xType.Struct[ElementType] = FromGoType(goType.Elem())
	case reflect.Struct:
		xType.Extra[NumField] = goType.NumField()
		for i := 0; i < goType.NumField(); i++ {
			field := goType.Field(i)
			xType.Extra[Field+strconv.Itoa(i)] = field.Name
			xType.Struct[field.Name] = FromGoType(field.Type)
		}
	default: // simple type like int、float64. The Struct is a nil map
		xType.Struct = nil
		xType.Extra = nil
	}
	return xType
}

// GoType returns the std lib reflect.Type representation of the Type
func (t Type) GoType() reflect.Type {
	kind := kindMap[t.Kind]
	switch kind {
	case reflect.Invalid:
		panic("invalid kind of type")
	case reflect.Bool:
		return reflect.TypeOf(false)
	case reflect.Int:
		return reflect.TypeOf(int(0))
	case reflect.Int8:
		return reflect.TypeOf(int8(0))
	case reflect.Int16:
		return reflect.TypeOf(int16(0))
	case reflect.Int32:
		return reflect.TypeOf(int32(0))
	case reflect.Int64:
		return reflect.TypeOf(int64(0))
	case reflect.Uint:
		return reflect.TypeOf(uint(0))
	case reflect.Uint8:
		return reflect.TypeOf(uint8(0))
	case reflect.Uint16:
		return reflect.TypeOf(uint16(0))
	case reflect.Uint32:
		return reflect.TypeOf(uint32(0))
	case reflect.Uint64:
		return reflect.TypeOf(uint64(0))
	case reflect.Uintptr:
		return reflect.TypeOf(uintptr(0))
	case reflect.Float32:
		return reflect.TypeOf(float32(0))
	case reflect.Float64:
		return reflect.TypeOf(float64(0))
	case reflect.Complex64:
		return reflect.TypeOf(complex64(0))
	case reflect.Complex128:
		return reflect.TypeOf(complex128(0))
	case reflect.Array:
		return reflect.ArrayOf(t.Extra[ArrayLength].(int), t.Struct[ElementType].GoType())
	case reflect.Chan:
		return reflect.ChanOf(reflect.ChanDir(t.Extra[ChanDir].(int)), t.Struct[ElementType].GoType())
	case reflect.Func:
		var in, out []reflect.Type
		var variadic bool = t.Extra[ValueType].(bool)
		for i := 0; i < t.Extra[NumIn].(int); i++ {
			in = append(in, t.Struct[In+strconv.Itoa(i)].GoType())
		}
		for i := 0; i < t.Extra[NumOut].(int); i++ {
			in = append(in, t.Struct[Out+strconv.Itoa(i)].GoType())
		}
		return reflect.FuncOf(in, out, variadic)
	case reflect.Interface:
		panic("Interface type are not impl")
	case reflect.Map:
		return reflect.MapOf(t.Struct[KeyType].GoType(), t.Struct[ValueType].GoType())
	case reflect.Pointer:
		return reflect.PointerTo(t.Struct[ElementType].GoType())
	case reflect.Slice:
		return reflect.SliceOf(t.Struct[ElementType].GoType())
	case reflect.String:
		return reflect.TypeOf("")
	case reflect.Struct:
		var fields []reflect.StructField
		for i := 0; i < t.Extra[NumField].(int); i++ {
			fieldName := t.Extra[Field+strconv.Itoa(i)].(string)
			fieldType := t.Struct[fieldName].GoType()
			fields = append(fields, reflect.StructField{Name: fieldName, Type: fieldType})
		}
		return reflect.StructOf(fields)
	default:
		panic("Unknown type " + t.Kind)
	}
}

func (t Type) String() string {
	buf := bytes.Buffer{}
	_ = json.NewEncoder(&buf).Encode(t)
	return buf.String()
}

func init() {
	for i := 0; i <= int(reflect.UnsafePointer); i++ {
		kind := reflect.Kind(i)
		kindMap[kind.String()] = kind
	}
	gob.Register(Type{})
}
