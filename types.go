package main

/*
#cgo pkg-config: glib-2.0
#include <glib.h>
*/
import "C"

const GoVoidPointer = "interface{}"
const CVoidPointer = "gpointer"

var TypeTagToGo = map[TypeTag]string{
	VoidTag:     "",
	BooleanTag:  "bool",
	Int8Tag:     "int8",
	Uint8Tag:    "uint8",
	Int16Tag:    "int16",
	Uint16Tag:   "uint16",
	Int32Tag:    "int32",
	Uint32Tag:   "uint32",
	Int64Tag:    "int64",
	Uint64Tag:   "uint64",
	FloatTag:    "float32",
	DoubleTag:   "float64",
	GTypeTag:    "int",
	Utf8Tag:     "string",
	FilenameTag: "string",
	// TODO: figure out how to do complex types
	/*
		ArrayTag
		InterfaceTag
		GListTag
		GSListTag
		GHashTag
		ErrorTag
	*/
	// another basic type
	//UnicharTag
}

var TypeTagToC = map[TypeTag]string{
	VoidTag:     "",
	BooleanTag:  "gboolean",
	Int8Tag:     "gint8",
	Uint8Tag:    "guint8",
	Int16Tag:    "gint16",
	Uint16Tag:   "guint16",
	Int32Tag:    "gint32",
	Uint32Tag:   "guint32",
	Int64Tag:    "gint64",
	Uint64Tag:   "guint64",
	FloatTag:    "gfloat",
	DoubleTag:   "gdouble",
	GTypeTag:    "gint",
	Utf8Tag:     "gchar",
	FilenameTag: "gchar",
}
