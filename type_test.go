package reflectx

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func TestInt(t *testing.T) {
	tp := TypeOf(1)
	// {"kind":"int"}
	t.Log(tp.String())
	assert(tp.Kind == "int")
}

func TestString(t *testing.T) {
	tp := TypeOf("a")
	// {"kind":"string"}
	t.Log(tp.String())
	assert(tp.Kind == "string")
}

func TestFloat32(t *testing.T) {
	tp := TypeOf(float32(3.14))
	// {"kind":"float32"}
	t.Log(tp.String())
	assert(tp.Kind == "float32")
}

func TestStruct(t *testing.T) {
	tp := TypeOf(struct {
		Name string
		Age  int64
	}{})
	// {"kind":"struct","struct":{"Age":{"kind":"int64"},"Name":{"kind":"string"}},"extra":{"field_0":"Name","field_1":"Age","fields_number":2}}
	t.Log(tp.String())
	assert(tp.Kind == "struct")
	assert(tp.Extra[NumField] == 2)
	assert(tp.Extra[Field+"0"] == "Name")
	assert(tp.Extra[Field+"1"] == "Age")
	assert(tp.Struct["Name"].Kind == "string")
	assert(tp.Struct["Age"].Kind == "int64")
}

func TestArray(t *testing.T) {
	tp := TypeOf([16]int{})
	// {"kind":"array","struct":{"element_type":{"kind":"int"}},"extra":{"array_length":16}}
	t.Log(tp.String())
	assert(tp.Kind == "array")
	assert(tp.Extra[ArrayLength] == 16)
	assert(tp.Struct[ElementType].Kind == "int")
}

func TestMap(t *testing.T) {
	tp := TypeOf(map[string]int{})
	// {"kind":"map","struct":{"key_type":{"kind":"string"},"value_type":{"kind":"int"}}}
	t.Log(tp.String())
	assert(tp.Kind == "map")
	assert(tp.Struct[KeyType].Kind == "string")
	assert(tp.Struct[ValueType].Kind == "int")
}

func TestPoinerInt(t *testing.T) {
	tp := TypeOf(new(int))
	// {"kind":"ptr","struct":{"element_type":{"kind":"int"}}}
	t.Log(tp.String())
	assert(tp.Kind == "ptr")
	assert(tp.Struct[ElementType].Kind == "int")
}

func TestSlice(t *testing.T) {
	tp := TypeOf([]string{})
	// {"kind":"slice","struct":{"element_type":{"kind":"string"}}}
	t.Log(tp.String())
	assert(tp.Kind == "slice")
	assert(tp.Struct[ElementType].Kind == "string")
}

func TestFunc(t *testing.T) {
	tp := TypeOf(func() {})
	// {"kind":"func","extra":{"in_number":0,"out_number":0}}
	t.Log(tp.String())
	assert(tp.Kind == "func")
	assert(tp.Extra[NumIn] == 0)
	assert(tp.Extra[NumOut] == 0)
}

func TestFunc2(t *testing.T) {
	tp := TypeOf(func(string, int) (bool, struct{}) { return false, struct{}{} })
	// {"kind":"func","struct":{"in_0":{"kind":"string"},"in_1":{"kind":"int"},
	// "out_0":{"kind":"bool"},"out_1":{"kind":"struct","extra":{"fields_number":0}}},
	// "extra":{"in_number":2,"out_number":2}}
	t.Log(tp.String())
	assert(tp.Kind == "func")
	assert(tp.Extra[NumIn] == 2)
	assert(tp.Extra[NumOut] == 2)
	assert(tp.Struct[In+"0"].Kind == "string")
	assert(tp.Struct[In+"1"].Kind == "int")
	assert(tp.Struct[Out+"0"].Kind == "bool")
	assert(tp.Struct[Out+"1"].Kind == "struct")
}

func TestGoTypeJson(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	buf := bytes.Buffer{}
	_ = json.NewEncoder(&buf).Encode(Person{Name: "Bob", Age: 36})
	t.Log(buf.String()) // {"Name":"Bob","Age":36}

	tp := TypeOf(Person{})
	t.Log(tp.String()) // {"kind":"struct","struct":{"Age":{"kind":"int"},"Name":{"kind":"string"}},"extra":{"field_0":"Name","field_1":"Age","fields_number":2}}

	obj := reflect.New(tp.GoType()).Interface()
	_ = json.NewDecoder(&buf).Decode(obj)
	t.Logf("%+v", obj) // &{Name:Bob Age:36}

	person := *((*Person)(reflect.ValueOf(obj).UnsafePointer()))
	t.Logf("%+v", person) // {Name:Bob Age:36}
}

func TestGoTypeGob(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	gob.Register(Person{})
	buf := bytes.Buffer{}
	_ = gob.NewEncoder(&buf).Encode(Person{Name: "Bob", Age: 36})
	t.Log(buf.String())
	tp := TypeOf(Person{}).GoType()
	t.Log(tp.String())
	obj := reflect.New(tp).Interface()
	_ = gob.NewDecoder(&buf).Decode(obj)
	t.Logf("%+v", obj)

	t.Log(reflect.ValueOf(obj).Kind().String()) // ptr
	person := *((*Person)(reflect.ValueOf(obj).UnsafePointer()))
	t.Logf("%+v", person)
}

func TestRPC(t *testing.T) {
	fun := func(a, b int) int { return a + b }
	funcRegister[TypeOf(fun).String()] = fun
	s := rpc[int](fun, 1, 2)
	t.Log(s)
	assert(s == 3)

	fun2 := func(es []int, a string) []int {
		for i := range es {
			n, _ := strconv.ParseInt(a, 10, 64)
			es[i] += int(n)
		}
		return es
	}
	funcRegister[TypeOf(fun2).String()] = fun2

	r := rpc[[]int](fun2, []int{1, 2, 3}, "100")
	t.Logf("%+v", r)

	assert(len(r) == 3)
	assert(r[0] == 101)
	assert(r[1] == 102)
	assert(r[2] == 103)
}

func TestRPC2(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	fun := func(ps []Person) map[int]int {
		r := map[int]int{}
		for _, p := range ps {
			r[p.Age]++
		}
		return r
	}
	funcRegister[TypeOf(fun).String()] = fun
	m := rpc[map[int]int](fun, []Person{
		{"a", 20}, {"a", 20}, {"a", 22}, {"a", 22}, {"a", 20}, {"a", 20},
	})
	t.Log(m)
	assert(len(m) == 2)
	assert(m[20] == 4)
	assert(m[22] == 2)
}

func BenchmarkDirect(t *testing.B) {
	fun := func(a, b int) int { return a + b }

	ss := 0
	for i := 0; i < t.N; i++ {
		ss += fun(rand.Int(), rand.Int())
	}
	t.Log(ss)
}

func BenchmarkDynamic(t *testing.B) {
	fun := func(a, b int) int { return a + b }
	funcRegister[TypeOf(fun).String()] = fun

	ss := 0
	for i := 0; i < t.N; i++ {
		ss += rpc[int](fun, rand.Int(), rand.Int())
	}
	t.Log(ss)
}

var funcRegister = map[string]interface{}{} // type.str->func

func rpc[R any](fun any, args ...any) (r R) {
	buf := bytes.Buffer{}
	ec := gob.NewEncoder(&buf)

	funType := TypeOf(fun)
	_ = ec.Encode(funType)
	_ = ec.Encode(funType.Extra[NumIn])
	for _, arg := range args {
		_ = ec.Encode(arg)
	}

	ret := doRpc(buf.Bytes())

	_ = gob.NewDecoder(bytes.NewBuffer(ret)).Decode(&r)
	return
}

func doRpc(b []byte) []byte {
	buf := bytes.NewBuffer(b)
	dc := gob.NewDecoder(buf)

	var funType Type
	_ = dc.Decode(&funType)
	fun := funcRegister[funType.String()]
	var funGoType = reflect.TypeOf(fun)

	var numberIn int
	_ = dc.Decode(&numberIn)

	var in []reflect.Value
	for i := 0; i < numberIn; i++ {
		var argType = funGoType.In(i)

		obj := reflect.New(argType).Interface()
		_ = dc.Decode(obj)
		in = append(in, reflect.ValueOf(obj).Elem())
	}

	r := reflect.ValueOf(fun).Call(in)
	rBuf := bytes.Buffer{}
	_ = gob.NewEncoder(&rBuf).EncodeValue(r[0])
	return rBuf.Bytes()
}

func assert(b bool, infos ...any) {
	if !b {
		panic(fmt.Sprint(infos...))
	}
}
