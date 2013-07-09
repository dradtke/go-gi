package main

/*
#cgo pkg-config: glib-2.0 gobject-2.0
#include <glib.h>
#include <glib-object.h>

gboolean check_version(gint major, gint minor) {
	return GLIB_CHECK_VERSION(major, minor, 0);
}
*/
import "C"
import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run main.go <namespace>")
		return
	}

	// don't do this for GLib 2.36 and higher
	if C.check_version(C.gint(2), C.gint(36)) == 0 {
		C.g_type_init()
	}

	namespace := os.Args[1]
	typelib, err := LoadNamespace(namespace)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer FreeTypelib(typelib)

	var code bytes.Buffer
	giTemplates := Search(os.Getenv("GOPATH"), filepath.Join("src", "github.com", "dradtke", "go-gi", "templates"))
	tmpl, err := template.ParseGlob(filepath.Join(giTemplates, "*"))
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
