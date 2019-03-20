package zbus

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	type T struct {
		Name string
		Age  float64
	}

	arg := T{"Azmy", 36}
	request, err := NewRequest("my-id", "object", "DoSomething", "arg1", 2, arg)

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	msg, err := request.Encode()
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	loaded, err := LoadRequest(msg)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, request.ID, loaded.ID); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, request.Object, loaded.Object); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, request.Method, loaded.Method); !ok {
		t.Error()
	}

	must := func(o interface{}, err error) interface{} {
		if ok := assert.NoError(t, err); !ok {
			t.Fatal()
		}
		return o
	}

	if ok := assert.Equal(t, "arg1", must(loaded.Value(0, reflect.TypeOf("")))); !ok {
		t.Error()
	}

	intArg := must(loaded.Value(1, reflect.TypeOf(0)))
	if ok := assert.Equal(t, 2, intArg); !ok {
		t.Error()
	}

	structArg := must(loaded.Value(2, reflect.TypeOf(T{})))
	if ok := assert.Equal(t, arg, structArg); !ok {
		t.Error()
	}
}
