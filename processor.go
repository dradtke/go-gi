package main

import (
	"container/list"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

/* --- Enums --- */

type EnumDefinition struct {
	EnumName string
	CType    string
	Values   []EnumValue
}

type EnumValue struct {
	Name     string
	EnumName string
	Value    int64
}

func ProcessEnum(info *BaseInfo, code *bytes.Buffer, tmpl *template.Template) {
	name := info.GetName()
	prefix := GetCPrefix(info.GetNamespace())
	def := &EnumDefinition{EnumName:name, CType:prefix+name}

	numValues := info.GetNEnumValues()
	for i := 0; i < numValues; i++ {
		value := info.GetEnumValue(i)
		valDef := EnumValue{Name:CamelCase(value.GetName()), EnumName:name, Value:value.GetValue()}
		def.Values = append(def.Values, valDef)
	}

	err := tmpl.ExecuteTemplate(code, "enum", def)
	if err != nil {
		fmt.Println(err.Error())
	}
}

/* --- Objects --- */

type ObjectDefinition struct {
	ObjectName    string
	InterfaceName string
	CType         string
	CastFunc      string
}

func (info *BaseInfo) GetObjectDefinition() ObjectDefinition {
	name := info.GetName()
	prefix := GetCPrefix(info.GetNamespace())
	return ObjectDefinition{
		ObjectName:    name,
		InterfaceName: name + "Like",
		CType:         prefix + name,
		CastFunc:      "As" + name,
	}
}

func ProcessObject(info *BaseInfo, code *bytes.Buffer, tmpl *template.Template) {
	var err error
	def := info.GetObjectDefinition()

	// write object definition
	err = tmpl.ExecuteTemplate(code, "object-definition", def)
	if err != nil {
		fmt.Println(err.Error())
	}

	// write interface definition
	err = tmpl.ExecuteTemplate(code, "interface-definition", def)
	if err != nil {
		fmt.Println(err.Error())
	}

	implementAll(def, info, code, tmpl)

	// object methods
	methodMap := make(map[string] bool)
	writeMethods(def, info, code, tmpl, methodMap, "")

	for hasParent(info) {
		info = info.GetParent()
		writeMethods(def, info, code, tmpl, methodMap, info.GetName())
	}
}

func writeMethods(def ObjectDefinition, info *BaseInfo, code *bytes.Buffer, tmpl *template.Template, methods map[string] bool, className string) {
	numMethods := info.GetNObjectMethods()
	for i := 0; i < numMethods; i++ {
		method := info.GetObjectMethod(i)
		flags := method.GetFunctionFlags()
		name := method.GetName()

		if methods[name] {
			continue
		}
		methods[name] = true

		goargs, gorets, cargs, crets, err := readParams(method, flags)
		if err != nil {
			// TODO: log this
			continue
		}

		fn := FunctionDefinition{
			Name:name,
			Owner:&def,
			ClassName:def.ObjectName,
			ForGo:ArgsAndRets{Args:goargs, Rets:gorets},
			ForC:ArgsAndRets{Args:cargs, Rets:crets},
			Flags:flags,
			Info:method,
		}
		if className != "" {
			fn.ClassName = className
		}

		var marshal bytes.Buffer
		for _, param := range cargs {
			switch param.Dir {
				case In, InOut: tmpl.ExecuteTemplate(&marshal, "c-marshal", param)
				case Out: tmpl.ExecuteTemplate(&marshal, "c-decl", param)
			}
		}
		fn.ArgMarshalBody = marshal.String()
		marshal.Reset()
		for _, ret := range crets {
			tmpl.ExecuteTemplate(&marshal, "go-marshal", ret)
		}
		fn.RetMarshalBody = marshal.String()
		if className == "" {
			tmpl.ExecuteTemplate(code, "go-function", fn)
		}
		tmpl.ExecuteTemplate(code, "go-function-wrapper", fn)
	}
}

func implementAll(def ObjectDefinition, face *BaseInfo, code *bytes.Buffer, tmpl *template.Template) {
	impl := face.GetObjectDefinition()
	impl.ObjectName = def.ObjectName
	err := tmpl.ExecuteTemplate(code, "object-implement", impl)
	if err != nil {
		fmt.Println(err.Error())
	}

	if hasParent(face) {
		implementAll(def, face.GetParent(), code, tmpl)
	}
}

func hasParent(info *BaseInfo) bool {
	return info.GetName() != "Object" && !info.IsFundamental()
}

/* -- Functions -- */

type ArgsAndRets struct {
	Args []Parameter
	Rets []Parameter
}

type FunctionDefinition struct {
	Name string
	Owner *ObjectDefinition
	ClassName string
	ForGo ArgsAndRets
	ForC ArgsAndRets
	ArgMarshalBody string
	RetMarshalBody string
	Flags FunctionFlags
	Info *BaseInfo
}

func (def FunctionDefinition) GoName() string {
	return CamelCase(def.Name)
}

func (def FunctionDefinition) CName() string {
	return def.Info.GetSymbol()
}

func (def FunctionDefinition) HasOwner() bool {
	return def.Owner != nil
}

func (def FunctionDefinition) ReturnsValue() bool {
	return len(def.ForC.Rets) > 0
}

func (def FunctionDefinition) Arglist(wrapper bool) string {
	var result []string
	index := 0
	if (def.Owner != nil && !wrapper) {
		result = make([]string, len(def.ForGo.Args) + 1)
		result[index] = "self " + def.Owner.InterfaceName
		index++
	} else {
		result = make([]string, len(def.ForGo.Args))
	}
	for i, arg := range def.ForGo.Args {
		result[i + index] = arg.Name + " " + arg.GoType
	}
	return strings.Join(result, ", ")
}

func (def FunctionDefinition) Retlist() string {
	result := make([]string, len(def.ForGo.Rets))
	for i, ret := range def.ForGo.Rets {
		result[i] = ret.Name + " " + ret.GoType
	}
	return strings.Join(result, ", ")
}

func (def FunctionDefinition) MarshaledValues() string {
	var result []string
	index := 0
	if (def.Owner != nil) {
		result = make([]string, len(def.ForC.Args) + 1)
		result[index] = "self.As" + def.Owner.ObjectName + "()"
		index++
	} else {
		result = make([]string, len(def.ForC.Args))
	}
	for i, param := range def.ForC.Args {
		name := param.CName()
		if param.Dir == Out || param.Dir == InOut {
			name = "&" + name
		}
		result[i + index] = name
	}
	return strings.Join(result, ", ")
}

func (def FunctionDefinition) CRet() Parameter {
	return def.ForC.Rets[0]
}

type Parameter struct {
	Name string
	Dir Direction
	GoType string
	CType string
	Info *BaseInfo
}

func (val Parameter) CName() string {
	return "c_" + val.Name
}

func (val Parameter) IsPointer() bool {
	// GErrors should always be handled as pointers
	if val.CType == "GError" {
		return true
	} else if val.Info == nil {
		return false
	}
	return val.Info.GetType().IsPointer()
}

func returnsValue(typ *BaseInfo) bool {
	// a function doesn't return a value iff its tag is void and not a pointer
	// a void pointer represents an arbitrary value
	return typ.IsPointer() || typ.GetTag() != VoidTag
}

func readParams(info *BaseInfo, flags FunctionFlags) ([]Parameter, []Parameter, []Parameter, []Parameter, error) {
	goargList := list.New()
	goretList := list.New()
	cargList := list.New()
	cretList := list.New()
	marshalError := errors.New("couldn't marshal type")

	ret := info.GetReturnType()
	if returnsValue(ret) {
		var (
			gotype, ctype string
			ok bool
		)
		tag := ret.GetTag()
		if tag == VoidTag && ret.IsPointer() {
			gotype = GoVoidPointer
			ctype = CVoidPointer
		} else {
			if gotype, ok = TypeTagToGo[tag]; !ok {
				return nil, nil, nil, nil, marshalError
			}
			if ctype, ok = TypeTagToC[tag]; !ok {
				return nil, nil, nil, nil, marshalError
			}
		}
		cretList.PushBack(Parameter{Name:"retval", Dir:Out, GoType:gotype, CType:ctype, Info:nil})
	}

	n := info.GetNArgs()
	for i := 0; i < n; i++ {
		param := info.GetArg(i)
		dir := param.GetDirection()
		name := param.GetName()

		var (
			gotype, ctype string
			ok bool
		)
		tag := param.GetType().GetTag()
		if tag == VoidTag && param.GetType().IsPointer() {
			gotype = GoVoidPointer
			ctype = CVoidPointer
		} else {
			if gotype, ok = TypeTagToGo[tag]; !ok {
				return nil, nil, nil, nil, marshalError
			}
			if ctype, ok = TypeTagToC[tag]; !ok {
				return nil, nil, nil, nil, marshalError
			}
		}

		p := Parameter{Name:name, Dir:dir, GoType:gotype, CType:ctype, Info:param}
		cargList.PushBack(p)
		if dir == In || dir == InOut {
			goargList.PushBack(p)
		}
		if dir == Out || dir == InOut {
			goretList.PushBack(p)
		}
	}

	if flags.Throws {
		goretList.PushBack(Parameter{Name:"error", Dir:Out, GoType:"error", CType:"GError", Info:nil})
	}

	goArgs := make([]Parameter, goargList.Len())
	for e, i := goargList.Front(), 0; e != nil; e, i = e.Next(), i + 1 {
		goArgs[i] = e.Value.(Parameter)
	}

	goRets := make([]Parameter, goretList.Len())
	for e, i := goretList.Front(), 0; e != nil; e, i = e.Next(), i + 1 {
		goRets[i] = e.Value.(Parameter)
	}

	cArgs := make([]Parameter, cargList.Len())
	for e, i := cargList.Front(), 0; e != nil; e, i = e.Next(), i + 1 {
		cArgs[i] = e.Value.(Parameter)
	}

	cRets := make([]Parameter, cretList.Len())
	for e, i := cretList.Front(), 0; e != nil; e, i = e.Next(), i + 1 {
		cRets[i] = e.Value.(Parameter)
	}

	return goArgs, goRets, cArgs, cRets, nil
}
