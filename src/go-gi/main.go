package main

/*
#cgo pkg-config: glib-2.0
#include <glib-object.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go-gi <namespace>")
		return
	}

	C.g_type_init()
	namespace := os.Args[1]
	typelib, err := LoadNamespace(namespace)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer FreeTypelib(typelib)

	var code bytes.Buffer
	tmpl, err := template.ParseGlob("templates/*")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	n := GetNumInfos(namespace)
	for i := 0; i < n; i++ {
		info := GetInfo(namespace, i)
		switch info.Type {
			case Enum: ProcessEnum(info, &code, tmpl)
			case Object: ProcessObject(info, &code, tmpl)
		}
		info.Free()
	}

	code.WriteTo(os.Stdout)
}
