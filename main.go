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
	"log"
	"os"
	"path/filepath"
	"strings"
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
	ns := strings.ToLower(namespace)
	fmt.Println("generating " + namespace + " bindings...")
	typelib, err := LoadNamespace(namespace)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer FreeTypelib(typelib)

	gopath := strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator))
	giTemplatesRel := filepath.Join("src", "github.com", "dradtke", "go-gi", "templates")
	outputDirRel := filepath.Join("src", "gi", ns)
	var giTemplates, outputDir string

	// find a) the templates directory, and b) a place to put output files
	for _, dir := range gopath {
		if giTemplates == "" {
			f := filepath.Join(dir, giTemplatesRel)
			if _, err := os.Stat(f); !os.IsNotExist(err) {
				giTemplates = f
			}
		}
		if outputDir == "" {
			f := filepath.Join(dir, outputDirRel)
			if err := os.MkdirAll(f, 0755); err != nil {
				fmt.Println(err.Error())
			} else {
				outputDir = f
			}
		}
	}
	if giTemplates == "" {
		log.Fatal("template folder not found")
	}
	if outputDir == "" {
		log.Fatal("no writable output directory found")
	}

	var code bytes.Buffer
	tmpl := template.Must(template.New("go-gi").ParseGlob(filepath.Join(giTemplates, "*")))

	code.WriteString("package " + ns + "\n\n")

	n := GetNumInfos(namespace)
	for i := 0; i < n; i++ {
		info := GetInfo(namespace, i)
		switch info.Type {
			case Enum: ProcessEnum(info, &code, tmpl)
			case Object: ProcessObject(info, &code, tmpl)
		}
		info.Free()
	}

	file, err := os.Create(filepath.Join(outputDir, ns + ".go"))
	if err != nil {
		log.Fatal(err.Error())
	}

	code.WriteTo(file)
	file.Close()
	fmt.Println("bindings written to " + file.Name())
	fmt.Println("run \"go build gi/" + ns + "\" to compile them")
}
