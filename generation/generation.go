package generation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"
)

const (
	//ArgumentPrefix argument name prefix
	ArgumentPrefix = "arg"
	//ReturnPrefix return name prefix
	ReturnPrefix = "ret"
)

// Generator builds a generator code
func Generator(fqn string) (*jen.File, error) {
	parts := strings.Split(fqn, "+")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid interface fqn name expecting format package+Interface")
	}

	f := jen.NewFile("main")
	f.Func().Id("main").Params().Block(
		jen.Id("opt").Op(":=").Qual("github.com/threefoldtech/zbus/generation", "NewOptions").Call(),
		jen.Id("inf").Op(":=").Parens(jen.Op("*").Qual(parts[0], parts[1])).Call(jen.Nil()),
		jen.If(
			jen.Id("err").Op(":=").Qual("github.com/threefoldtech/zbus/generation", "Generate").Call(jen.Id("opt"), jen.Id("inf")),
			jen.Id("err").Op("!=").Nil(),
		).Block(
			jen.Panic(jen.Id("err")),
		),
	)

	return f, nil
}

//Generate generates stubs for given interface
func Generate(opt Options, inf interface{}) error {
	typ := reflect.TypeOf(inf)
	if typ.Kind() != reflect.Ptr {
		//this will probably happen if only not called by our
		//intermediate tool
		return fmt.Errorf("inf kind is not a pointer")
	}

	elem := typ.Elem()
	if elem.Kind() != reflect.Interface {
		return fmt.Errorf("not an interface")
	}

	return generateSub(opt, elem).Save("/dev/stdout")
}

func generateSub(opt Options, typ reflect.Type) *jen.File {
	stub := fmt.Sprintf("%sStub", typ.Name())
	f := jen.NewFile(opt.Package)

	//define the struct
	f.Type().Id(stub).Struct(
		jen.Id("client").Qual("github.com/threefoldtech/zbus", "Client"),
		jen.Id("module").Qual("", "string"),
		jen.Id("object").Qual("github.com/threefoldtech/zbus", "ObjectID"),
	)

	//generate the constructor
	f.Func().Id(fmt.Sprintf("New%s", stub)).Params(
		jen.Id("client").Qual("github.com/threefoldtech/zbus", "Client"),
	).Params(jen.Op("*").Id(stub)).Block(
		jen.Return(
			jen.Op("&").Id(stub).
				Block(
					jen.Id("client").Op(":").Id("client").Op(","),
					jen.Id("module").Op(":").Lit(opt.Module).Op(","),
					jen.Id("object").Op(":").Qual("github.com/threefoldtech/zbus", "ObjectID").Block(
						jen.Id("Name").Op(":").Lit(opt.Name).Op(","),
						jen.Id("Version").Op(":").Lit(opt.Version).Op(","),
					).Op(","),
				),
		),
	)
	//generate the methods
	for i := 0; i < typ.NumMethod(); i++ {
		generateFunc(f, stub, typ.Method(i))
	}

	return f
}

func generateFunc(f *jen.File, name string, method reflect.Method) {
	f.Func().Parens(jen.Id("s").Op("*").Id(name)).Id(method.Name).
		Params(getMethodParams(&method)...).
		Params(getMethodReturn(&method)...).Block(
		getMethodBody(&method)...,
	)
}

func getMethodBody(m *reflect.Method) []jen.Code {
	typ := m.Type

	var code []jen.Code
	args := []jen.Code{
		jen.Id("s").Dot("module"),
		jen.Id("s").Dot("object"),
		jen.Lit(m.Name),
	}

	for i := 0; i < typ.NumIn(); i++ {
		stmt := jen.Id(fmt.Sprintf("%s%d", ArgumentPrefix, i))
		if typ.IsVariadic() && i == typ.NumIn()-1 {
			stmt = stmt.Op("...")
		}

		args = append(
			args,
			stmt,
		)
	}

	code = append(
		code,
		jen.List(jen.Id("result"), jen.Id("err")).Op(":=").Id("s").Dot("client").Dot("Request").
			Call(args...),
		jen.If(
			jen.Id("err").Op("!=").Nil().Block(
				jen.Panic(jen.Id("err")),
			),
		),
	)

	for i := 0; i != typ.NumOut(); i++ {
		name := fmt.Sprintf("%s%d", ReturnPrefix, i)
		code = append(
			code,
			jen.If(jen.Id("err").Op(":=").Id("result").Dot("Unmarshal").
				Call(
					jen.Lit(i),
					jen.Op("&").Id(name),
				), jen.Id("err").Op("!=").Nil()).Block(
				jen.Panic(jen.Id("err")),
			),
		)
	}
	code = append(
		code,
		jen.Return(),
	)
	return code
}

func getMethodReturn(m *reflect.Method) []jen.Code {
	var code []jen.Code
	typ := m.Type
	for i := 0; i < typ.NumOut(); i++ {
		argName := fmt.Sprintf("%s%d", ReturnPrefix, i)
		argType := typ.Out(i)

		code = append(
			code,
			getTypeCode(jen.Id(argName), argType),
		)
	}

	return code
}

func getTypeCode(s *jen.Statement, t reflect.Type) *jen.Statement {
	switch t.Kind() {
	case reflect.Slice:
		return getTypeCode(
			s.Op("[]"),
			t.Elem(),
		)
	default:
		name := t.Name()
		if name == "" {
			name = "interface{}"
		}

		return s.Qual(t.PkgPath(), name)
	}
}

func getMethodParams(m *reflect.Method) []jen.Code {
	var code []jen.Code
	typ := m.Type
	for i := 0; i < typ.NumIn(); i++ {
		argName := fmt.Sprintf("%s%d", ArgumentPrefix, i)
		argType := typ.In(i)
		stmt := jen.Id(argName)

		if typ.IsVariadic() && i == typ.NumIn()-1 {
			code = append(
				code,
				getTypeCode(stmt.Op("..."), argType.Elem()),
			)
			continue
		}
		code = append(
			code,
			getTypeCode(stmt, argType),
		)
	}

	return code
}
