# reflectx

make reflect type serializable, conveniently for RPC

一个可以序列化的反射类型，代替 reflect.Type，便于编写 RPC 代码

## example

TypeOf gets the type

使用 TypeOf 获得类型

```go
	tp := TypeOf(1)
	// {"kind":"int"}
	t.Log(tp.String())

    tp = TypeOf(struct {
		Name string
		Age  int64
	}{})
	// {"kind":"struct","struct":{"Age":{"kind":"int64"},"Name":{"kind":"string"}},"extra":{"field_0":"Name","field_1":"Age","fields_number":2}}
	t.Log(tp.String())

    tp = TypeOf(map[string]int{})
	// {"kind":"map","struct":{"key_type":{"kind":"string"},"value_type":{"kind":"int"}}}
	t.Log(tp.String())

    tp = TypeOf([]string{})
	// {"kind":"slice","struct":{"element_type":{"kind":"string"}}}
	t.Log(tp.String())
```

---

FromGoType converts the std reflect.Type to Type

FromGoType 可以将标准库中的 reflect.Type 转为 Type

---

GoType returns the std lib reflect.Type representation of the Type

GoType 将 Type 转为 标准库中的 reflect.Type


