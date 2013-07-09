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
	"io/ioutil"
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

	fmt.Println("[*] Generating " + namespace + " bindings...")
	// TODO: support specifying the namespace version
	typelib, err := LoadNamespace(namespace, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	defer FreeTypelib(typelib)

	srcRel := filepath.Join("src", "github.com", "dradtke", "go-gi")
	giSnippetsRel := filepath.Join(srcRel, "snippets")
	giTemplatesRel := filepath.Join(srcRel, "templates")
	giBlacklistRel := filepath.Join(srcRel, "blacklist")
	outputDirRel := filepath.Join("src", "gi", ns)
	var giSnippets, giTemplates, giBlacklist, outputDir string

	// find a) the templates directory, and b) a place to put output files
	gopath := strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator))
	for _, dir := range gopath {
		if giSnippets == "" {
			f := filepath.Join(dir, giSnippetsRel)
			if _, err := os.Stat(f); !os.IsNotExist(err) {
				giSnippets = f
			}
		}
		if giTemplates == "" {
			f := filepath.Join(dir, giTemplatesRel)
			if _, err := os.Stat(f); !os.IsNotExist(err) {
				giTemplates = f
			}
		}
		if giBlacklist == "" {
			f := filepath.Join(dir, giBlacklistRel)
			if _, err := os.Stat(f); !os.IsNotExist(err) {
				giBlacklist = f
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
	if giSnippets == "" {
		log.Fatal("snippet folder not found")
	}
	if giTemplates == "" {
		log.Fatal("template folder not found")
	}
	if outputDir == "" {
		log.Fatal("no writable output directory found")
	}

	startingPoint, err := os.Open(filepath.Join(giTemplates, ns + ".go"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	header, err := ioutil.ReadAll(startingPoint)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	blacklist := make(map[string] bool)
	if f, err := os.Open(filepath.Join(giBlacklist, ns)); err == nil {
		if data, err := ioutil.ReadAll(f); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if len(line) == 0 || line[0:1] == "#" {
					continue
				}
				blacklist[line] = true
			}
		}
	}

	var code bytes.Buffer
	code.Write(header)

	tmpl := template.Must(template.New("go-gi").ParseGlob(filepath.Join(giSnippets, "*")))

	// used to prevent duplicate methods, interfaces, etc.
	exists := make(map[string] bool)

	n := GetNumInfos(namespace)
	for i := 0; i < n; i++ {
		info := GetInfo(namespace, i)
		switch info.Type {
			case Enum: ProcessEnum(info, &code, tmpl, &blacklist)
			case Object: ProcessObject(info, &code, tmpl, &exists, &blacklist)
		}
		info.Free()
	}

	file, err := os.Create(filepath.Join(outputDir, ns + ".go"))
	if err != nil {
		log.Fatal(err.Error())
	}

	code.WriteTo(file)
	file.Close()
	fmt.Println("[*] Bindings written to " + file.Name())
	fmt.Println("[*] Run \"go build gi/" + ns + "\" to compile them.")
}
