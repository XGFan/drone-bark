package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

type Foo struct {
	A int32
	B float32
}

type Bar struct {
	C string
	D *string
}

type M struct {
	Map map[string]string
}

type M2 struct {
	Map *map[string]string
}

type A struct {
	E Foo
	F *Bar
}

type AWithNoName struct {
	Foo
	*Bar
}

type X struct {
	Strings []string
	Ints    []int
	Floats  []float32
}

type Book struct {
	Name Name
	Cat  Catalog
}
type Name struct {
	English string
	Chinese string
}
type Catalog struct {
	Name   string
	Parent *Catalog
}

type Map map[string]string

type TypeCase struct {
	Name      string
	Env       Map
	Data      interface{}
	WantErr   bool
	CheckFunc func(interface{}) bool
}

func (m Map) Lookup(keys []string) (string, bool) {
	s, ok := m[strings.Join(keys, "_")]
	return s, ok
}

func (m Map) FindPrefix(keys []string) map[string]string {
	newMap := make(map[string]string, 0)
	key := strings.Join(keys, "_")
	prefix := key + "_"
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			newMap[k[len(prefix):]] = v
		}
	}
	return newMap
}

func TestParse(t *testing.T) {
	env := make(map[string]string, 0)
	hasValueEnv := map[string]string{"A": "1", "B": "3", "C": "CCC", "D": "DDD"}
	tests := []TypeCase{
		{"int", env, 1, true, nil},
		{"float", env, 1.0, true, nil},
		{"string", env, "111", true, nil},
		{"struct", hasValueEnv, Foo{}, true, nil},
		{"struct ptr", hasValueEnv, &Foo{}, false, func(i interface{}) bool {
			return i.(*Foo).A == 1
		}},
		{"interface", env, bytes.NewReader(make([]byte, 0)), false, nil},
		{"map", env, env, true, nil},
		{"map ptr", env, &env, true, nil},
		{"foo", hasValueEnv, &Foo{}, false, func(i interface{}) bool {
			return i.(*Foo).A == 1 && i.(*Foo).B == 3.0
		}},
		{"bar", hasValueEnv, &Bar{}, false, func(i interface{}) bool {
			d := i.(*Bar).D
			return i.(*Bar).C == "CCC" && d != nil && *d == "DDD"
		}},
		{"A", map[string]string{"E_A": "1", "E_B": "3", "F_C": "CCC", "F_D": "DDD"}, &A{}, false, func(i interface{}) bool {
			a := i.(*A)
			return a.E.A == 1 &&
				a.E.B == 3.0 &&
				a.F.C == "CCC" &&
				*(a.F.D) == "DDD"
		}},
		{"Map1", map[string]string{"Map_A": "1", "Map_B": "3", "Map_C": "CCC", "Map_D": "DDD"}, &M{}, false, func(i interface{}) bool {
			m := i.(*M).Map
			return m["A"] == "1" &&
				m["B"] == "3" &&
				m["C"] == "CCC" &&
				m["D"] == "DDD"
		}},
		{"Map2", map[string]string{"Map_A": "1", "Map_B": "3", "Map_C": "CCC", "Map_D": "DDD"}, &M2{}, false, func(i interface{}) bool {
			m := *(i.(*M2).Map)
			return m["A"] == "1" &&
				m["B"] == "3" &&
				m["C"] == "CCC" &&
				m["D"] == "DDD"
		}},
		{"AWithNoName", map[string]string{"A": "1", "B": "3", "C": "CCC", "D": "DDD"}, &AWithNoName{}, false, func(i interface{}) bool {
			a := i.(*AWithNoName)
			return a.A == 1 &&
				a.B == 3.0 &&
				a.C == "CCC" &&
				*(a.D) == "DDD"
		}},
		{"Slice", map[string]string{"Strings": "A,B,C", "Ints": "1,2,3", "Floats": "1,2,3", "D": "DDD"}, &X{}, false, func(i interface{}) bool {
			x := i.(*X)
			return reflect.DeepEqual(x.Strings, []string{"A", "B", "C"}) &&
				reflect.DeepEqual(x.Ints, []int{1, 2, 3}) &&
				reflect.DeepEqual(x.Floats, []float32{1, 2, 3})
		}},
		{"NestedInNested", map[string]string{"Name_English": "book", "Name_Chinese": "书",
			"Cat_Name":               "Aa1分类",
			"Cat_Parent_Name":        "Aa分类",
			"Cat_Parent_Parent_Name": "A分类"}, &Book{}, false, func(i interface{}) bool {
			book := i.(*Book)
			return book.Name.Chinese == "书" &&
				book.Name.English == "book" &&
				book.Cat.Name == "Aa1分类" &&
				book.Cat.Parent.Name == "Aa分类" &&
				book.Cat.Parent.Parent.Name == "A分类" &&
				book.Cat.Parent.Parent.Parent == nil
		}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if err := Parse(tt.Env, tt.Data, []string{}); (err != nil) != tt.WantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.WantErr)
			}
			if tt.CheckFunc != nil && !tt.CheckFunc(tt.Data) {
				t.Errorf("check fail %+v", tt.Data)
			}
		})
	}
}
