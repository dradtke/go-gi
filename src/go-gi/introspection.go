package main

/*
#cgo pkg-config: glib-2.0 gobject-introspection-1.0
#include <glib.h>
#include <girepository.h>
*/
import "C"
import (
	"strings"
)

var prefixes map[string] string = make(map[string] string)

func LoadNamespace(namespace string) (*C.GITypelib, error) {
	ns := GlibString(namespace) ; defer FreeString(ns)
	var err *C.GError
	typelib := C.g_irepository_require(nil, ns, nil, 0, &err)
	if err != nil {
		return nil, NewGError(err)
	}
	return typelib, nil
}

func FreeTypelib(typelib *C.GITypelib) {
	C.g_typelib_free(typelib)
}

func GetNumInfos(namespace string) int {
	ns := GlibString(namespace) ; defer FreeString(ns)
	return GoInt(C.g_irepository_get_n_infos(nil, ns))
}

func GetInfo(namespace string, index int) *BaseInfo {
	ns := GlibString(namespace) ; defer FreeString(ns)
	i := GlibInt(index)
	return NewBaseInfo(C.g_irepository_get_info(nil, ns, i))
}

func GetCPrefix(namespace string) string {
	prefix, ok := prefixes[namespace]
	if ok {
		return prefix
	}
	ns := GlibString(namespace) ; defer FreeString(ns)
	prefix = GoString(C.g_irepository_get_c_prefix(nil, ns))
	prefixes[namespace] = prefix
	return prefix
}

type GError struct {
	Code int
	Message string
}

func (self GError) Error() string {
	return self.Message
}

func NewGError(err *C.GError) GError {
	defer C.g_error_free(err)
	return GError{Code:GoInt(err.code), Message:GoString(err.message)}
}



type InfoType C.GIInfoType
const (
	Function = C.GI_INFO_TYPE_FUNCTION
	Callback = C.GI_INFO_TYPE_CALLBACK
	Struct = C.GI_INFO_TYPE_STRUCT
	Boxed = C.GI_INFO_TYPE_BOXED
	Enum = C.GI_INFO_TYPE_ENUM
	Flags = C.GI_INFO_TYPE_FLAGS
	Object = C.GI_INFO_TYPE_OBJECT
	Interface = C.GI_INFO_TYPE_INTERFACE
	Constant = C.GI_INFO_TYPE_CONSTANT
	//ErrorDomain = C.GI_INFO_TYPE_ERRORDOMAIN
	Union = C.GI_INFO_TYPE_UNION
	Value = C.GI_INFO_TYPE_VALUE
	Signal = C.GI_INFO_TYPE_SIGNAL
	VFunc = C.GI_INFO_TYPE_VFUNC
	Property = C.GI_INFO_TYPE_PROPERTY
	Field = C.GI_INFO_TYPE_FIELD
	Arg = C.GI_INFO_TYPE_ARG
	Type = C.GI_INFO_TYPE_TYPE
	Unresolved = C.GI_INFO_TYPE_UNRESOLVED
)

func InfoTypeToString(typ InfoType) string {
	return GoString(C.g_info_type_to_string((C.GIInfoType)(typ)))
}

type BaseInfo struct {
	ptr *C.GIBaseInfo
	Type InfoType
}

func NewBaseInfo(ptr *C.GIBaseInfo) *BaseInfo {
	typ := (InfoType)(C.g_base_info_get_type(ptr))
	return &BaseInfo{ptr, typ}
}

func (info *BaseInfo) Free() {
	C.g_base_info_unref(info.ptr)
}

type TypeTag C.GITypeTag
const (
	VoidTag = C.GI_TYPE_TAG_VOID
	BooleanTag = C.GI_TYPE_TAG_BOOLEAN
	Int8Tag = C.GI_TYPE_TAG_INT8
	Uint8Tag = C.GI_TYPE_TAG_UINT8
	Int16Tag = C.GI_TYPE_TAG_INT16
	Uint16Tag = C.GI_TYPE_TAG_UINT16
	Int32Tag = C.GI_TYPE_TAG_INT32
	Uint32Tag = C.GI_TYPE_TAG_UINT32
	Int64Tag = C.GI_TYPE_TAG_INT64
	Uint64Tag = C.GI_TYPE_TAG_UINT64
	FloatTag = C.GI_TYPE_TAG_FLOAT
	DoubleTag = C.GI_TYPE_TAG_DOUBLE
	GTypeTag = C.GI_TYPE_TAG_GTYPE
	Utf8Tag = C.GI_TYPE_TAG_UTF8
	FilenameTag = C.GI_TYPE_TAG_FILENAME
	// non-basic types
	ArrayTag = C.GI_TYPE_TAG_ARRAY
	InterfaceTag = C.GI_TYPE_TAG_INTERFACE
	GListTag = C.GI_TYPE_TAG_GLIST
	GSListTag = C.GI_TYPE_TAG_GSLIST
	GHashTag = C.GI_TYPE_TAG_GHASH
	ErrorTag = C.GI_TYPE_TAG_ERROR
	// another basic type
	UnicharTag = C.GI_TYPE_TAG_UNICHAR
)

/* -- Base Info -- */

func (info *BaseInfo) GetName() string {
	return GoString(C.g_base_info_get_name(info.ptr))
}

func (info *BaseInfo) GetFullName() string {
	return strings.ToLower(GoString(C.g_base_info_get_namespace(info.ptr))) + "_" + info.GetName()
}

func (info *BaseInfo) GetNamespace() string {
	return GoString(C.g_base_info_get_namespace(info.ptr))
}

func (info *BaseInfo) IsDeprecated() bool {
	return GoBool(C.g_base_info_is_deprecated(info.ptr))
}

func (info *BaseInfo) GetAttribute(attr string) string {
	_attr := GlibString(attr) ; defer C.g_free((C.gpointer)(_attr))
	return GoString(C.g_base_info_get_attribute(info.ptr, _attr))
}

/* -- Callables -- */

type Transfer C.GITransfer
const (
	Nothing = C.GI_TRANSFER_NOTHING
	Container = C.GI_TRANSFER_CONTAINER
	Everything = C.GI_TRANSFER_EVERYTHING
)

func (info *BaseInfo) IsCallable() bool {
	switch info.Type {
	case Function, Signal, VFunc:
		return true
	}
	return false
}

func (info *BaseInfo) GetReturnType() *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_callable_info_get_return_type((*C.GICallableInfo)(info.ptr))))
}

func (info *BaseInfo) GetCallerOwns() Transfer {
	return (Transfer)(C.g_callable_info_get_caller_owns((*C.GICallableInfo)(info.ptr)))
}

func (info *BaseInfo) MayReturnNull() bool {
	return GoBool(C.g_callable_info_may_return_null((*C.GICallableInfo)(info.ptr)))
}

func (info *BaseInfo) GetReturnAttribute(name string) string {
	_name := GlibString(name) ; defer C.g_free((C.gpointer)(_name))
	return GoString(C.g_callable_info_get_return_attribute((*C.GICallableInfo)(info.ptr), _name))
}

// iterate return attributes?

func (info *BaseInfo) GetNArgs() int {
	return GoInt(C.g_callable_info_get_n_args((*C.GICallableInfo)(info.ptr)))
}

func (info *BaseInfo) GetArg(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_callable_info_get_arg((*C.GICallableInfo)(info.ptr), GlibInt(n))))
}

/* -- Function Info -- */

type FunctionFlags struct {
	IsMethod bool
	IsConstructor bool
	IsGetter bool
	IsSetter bool
	WrapsVFunc bool
	Throws bool
}

func NewFunctionFlags(bits C.GIFunctionInfoFlags) FunctionFlags {
	var flags FunctionFlags
	PopulateFlags(&flags, (C.gint)(bits), []C.gint{
		C.GI_FUNCTION_IS_METHOD,
		C.GI_FUNCTION_IS_CONSTRUCTOR,
		C.GI_FUNCTION_IS_GETTER,
		C.GI_FUNCTION_IS_SETTER,
		C.GI_FUNCTION_WRAPS_VFUNC,
		C.GI_FUNCTION_THROWS,
	})
	return flags
}

func (info *BaseInfo) GetSymbol() string {
	return GoString(C.g_function_info_get_symbol((*C.GIFunctionInfo)(info.ptr)))
}

func (info *BaseInfo) GetFunctionFlags() FunctionFlags {
	return NewFunctionFlags(C.g_function_info_get_flags((*C.GIFunctionInfo)(info.ptr)))
}

func (info *BaseInfo) GetFunctionProperty() *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_function_info_get_property((*C.GIFunctionInfo)(info.ptr))))
}

func (info *BaseInfo) GetFunctionVFunc() *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_function_info_get_vfunc((*C.GIFunctionInfo)(info.ptr))))
}

// invoke?

/* -- Signal Info -- */

type SignalFlags struct {
	RunFirst bool
	RunLast bool
	RunCleanup bool
	NoRecurse bool
	Detailed bool
	Action bool
	NoHooks bool
	MustCollect bool
	Deprecated bool
}

func NewSignalFlags(bits C.GSignalFlags) *SignalFlags {
	var flags SignalFlags
	PopulateFlags(&flags, (C.gint)(bits), []C.gint{
		C.G_SIGNAL_RUN_FIRST,
		C.G_SIGNAL_RUN_LAST,
		C.G_SIGNAL_RUN_CLEANUP,
		C.G_SIGNAL_NO_RECURSE,
		C.G_SIGNAL_DETAILED,
		C.G_SIGNAL_ACTION,
		C.G_SIGNAL_NO_HOOKS,
		C.G_SIGNAL_MUST_COLLECT,
		C.G_SIGNAL_DEPRECATED,
	})
	return &flags
}

func (info *BaseInfo) GetSignalFlags() *SignalFlags {
	return NewSignalFlags(C.g_signal_info_get_flags((*C.GISignalInfo)(info.ptr)))
}

func (info *BaseInfo) GetClassClosure() *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_signal_info_get_class_closure((*C.GISignalInfo)(info.ptr))))
}

func (info *BaseInfo) TrueStopsEmit() bool {
	return GoBool(C.g_signal_info_true_stops_emit((*C.GISignalInfo)(info.ptr)))
}

/* -- VFunc Info -- */

type VFuncFlags struct {
	MustChainUp bool
	MustOverride bool
	MustNotOverride bool
	Throws bool
}

func NewVFuncFlags(bits C.GIVFuncInfoFlags) *VFuncFlags {
	var flags VFuncFlags
	PopulateFlags(&flags, (C.gint)(bits), []C.gint{
		C.GI_VFUNC_MUST_CHAIN_UP,
		C.GI_VFUNC_MUST_OVERRIDE,
		C.GI_VFUNC_MUST_NOT_OVERRIDE,
		C.GI_VFUNC_THROWS,
	})
	return &flags
}

func (info *BaseInfo) GetVFuncFlags() *VFuncFlags {
	return NewVFuncFlags(C.g_vfunc_info_get_flags((*C.GIVFuncInfo)(info.ptr)))
}

func (info *BaseInfo) GetOffset() int {
	// TODO: check for a value of 0xFFFF, which means it's unknown
	return GoInt(C.g_vfunc_info_get_offset((*C.GIVFuncInfo)(info.ptr)))
}

func (info *BaseInfo) GetVFuncSignal() *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_vfunc_info_get_signal((*C.GIVFuncInfo)(info.ptr))))
}

func (info *BaseInfo) GetInvoker() *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_vfunc_info_get_invoker((*C.GIVFuncInfo)(info.ptr))))
}

/* -- RegisteredType Info -- */

func (info *BaseInfo) IsRegisteredType() bool {
	switch info.Type {
	case Enum, Interface, Object, Struct, Union:
		return true
	}
	return false
}

func (info *BaseInfo) GetRegisteredTypeName() string {
	return GoString(C.g_registered_type_info_get_type_name((*C.GIRegisteredTypeInfo)(info.ptr)))
}

func (info *BaseInfo) GetRegisteredTypeInit() string {
	return GoString(C.g_registered_type_info_get_type_init((*C.GIRegisteredTypeInfo)(info.ptr)))
}

// TODO: get gtype?

/* -- Enum Info -- */

func (info *BaseInfo) GetNEnumValues() int {
	return GoInt(C.g_enum_info_get_n_values((*C.GIEnumInfo)(info.ptr)))
}

func (info *BaseInfo) GetEnumValue(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_enum_info_get_value((*C.GIEnumInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetNEnumMethods() int {
	return GoInt(C.g_enum_info_get_n_methods((*C.GIEnumInfo)(info.ptr)))
}

func (info *BaseInfo) GetEnumMethod(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_enum_info_get_method((*C.GIEnumInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetStorageType() TypeTag {
	return (TypeTag)(C.g_enum_info_get_storage_type((*C.GIEnumInfo)(info.ptr)))
}

// this acts on GIValueInfo
func (info *BaseInfo) GetValue() int64 {
	return (int64)(C.g_value_info_get_value((*C.GIValueInfo)(info.ptr)))
}

/* -- Struct Info -- */

func (info *BaseInfo) GetNStructFields() int {
	return GoInt(C.g_struct_info_get_n_fields((*C.GIStructInfo)(info.ptr)))
}

func (info *BaseInfo) GetStructField(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_struct_info_get_field((*C.GIStructInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetNStructMethods() int {
	return GoInt(C.g_struct_info_get_n_methods((*C.GIStructInfo)(info.ptr)))
}

func (info *BaseInfo) GetStructMethod(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_struct_info_get_method((*C.GIStructInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) IsGTypeStruct() bool {
	return GoBool(C.g_struct_info_is_gtype_struct((*C.GIStructInfo)(info.ptr)))
}

func (info *BaseInfo) IsForeign() bool {
	return GoBool(C.g_struct_info_is_foreign((*C.GIStructInfo)(info.ptr)))
}

/* -- Object Info -- */

func (info *BaseInfo) GetObjectTypeName() string {
	return GoString(C.g_object_info_get_type_name((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetObjectTypeInit() string {
	return GoString(C.g_object_info_get_type_init((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) IsAbstract() bool {
	return GoBool(C.g_object_info_get_abstract((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) IsFundamental() bool {
	return GoBool(C.g_object_info_get_fundamental((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetParent() *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_object_info_get_parent((*C.GIObjectInfo)(info.ptr))))
}

func (info *BaseInfo) GetNObjectInterfaces() int {
	return GoInt(C.g_object_info_get_n_interfaces((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetObjectInterface(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_object_info_get_interface((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetNObjectFields() int {
	return GoInt(C.g_object_info_get_n_fields((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetObjectField(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_object_info_get_field((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetNObjectProperties() int {
	return GoInt(C.g_object_info_get_n_properties((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetObjectProperty(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_object_info_get_property((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetNObjectMethods() int {
	return GoInt(C.g_object_info_get_n_methods((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetObjectMethod(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_object_info_get_method((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetNSignals() int {
	return GoInt(C.g_object_info_get_n_signals((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetObjectSignal(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_object_info_get_signal((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetNVFuncs() int {
	return GoInt(C.g_object_info_get_n_vfuncs((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetVFunc(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_object_info_get_vfunc((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetNConstants() int {
	return GoInt(C.g_object_info_get_n_constants((*C.GIObjectInfo)(info.ptr)))
}

func (info *BaseInfo) GetConstant(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_object_info_get_constant((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

/* -- Arg Info -- */

type Direction C.GIDirection
const (
	In = C.GI_DIRECTION_IN
	Out = C.GI_DIRECTION_OUT
	InOut = C.GI_DIRECTION_INOUT
)

type ScopeType C.GIScopeType
const (
	Invalid = C.GI_SCOPE_TYPE_INVALID
	Call = C.GI_SCOPE_TYPE_CALL
	Async = C.GI_SCOPE_TYPE_ASYNC
	Notified = C.GI_SCOPE_TYPE_NOTIFIED
)

func (info *BaseInfo) GetDirection() Direction {
	return (Direction)(C.g_arg_info_get_direction((*C.GIArgInfo)(info.ptr)))
}

func (info *BaseInfo) IsCallerAllocates() bool {
	return GoBool(C.g_arg_info_is_caller_allocates((*C.GIArgInfo)(info.ptr)))
}

func (info *BaseInfo) IsReturnValue() bool {
	return GoBool(C.g_arg_info_is_return_value((*C.GIArgInfo)(info.ptr)))
}

func (info *BaseInfo) IsOptional() bool {
	return GoBool(C.g_arg_info_is_optional((*C.GIArgInfo)(info.ptr)))
}

func (info *BaseInfo) MayBeNull() bool {
	return GoBool(C.g_arg_info_may_be_null((*C.GIArgInfo)(info.ptr)))
}

func (info *BaseInfo) GetOwnershipTransfer() Transfer {
	return (Transfer)(C.g_arg_info_get_ownership_transfer((*C.GIArgInfo)(info.ptr)))
}

func (info *BaseInfo) GetScope() ScopeType {
	return (ScopeType)(C.g_arg_info_get_scope((*C.GIArgInfo)(info.ptr)))
}

// TODO: get closure/destroy?

func (info *BaseInfo) GetType() *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_arg_info_get_type((*C.GIArgInfo)(info.ptr))))
}

/* -- Type Info -- */

type ArrayType C.GIArrayType
const (
	CArray = C.GI_ARRAY_TYPE_C
	GArray = C.GI_ARRAY_TYPE_ARRAY
	PtrArray = C.GI_ARRAY_TYPE_PTR_ARRAY
	ByteArray = C.GI_ARRAY_TYPE_BYTE_ARRAY
)

func TypeTagToString(tag TypeTag) string {
	return GoString(C.g_type_tag_to_string((C.GITypeTag)(tag)))
}

func (info *BaseInfo) IsPointer() bool {
	return GoBool(C.g_type_info_is_pointer((*C.GITypeInfo)(info.ptr)))
}

func (info *BaseInfo) GetTag() TypeTag {
	return (TypeTag)(C.g_type_info_get_tag((*C.GITypeInfo)(info.ptr)))
}

func (info *BaseInfo) GetParamType(n int) *BaseInfo {
	return NewBaseInfo((*C.GIBaseInfo)(C.g_type_info_get_param_type((*C.GITypeInfo)(info.ptr), GlibInt(n))))
}

func (info *BaseInfo) GetTypeInterface() *BaseInfo {
	return NewBaseInfo(C.g_type_info_get_interface((*C.GITypeInfo)(info.ptr)))
}

func (info *BaseInfo) GetArrayLength() int {
	return GoInt(C.g_type_info_get_array_length((*C.GITypeInfo)(info.ptr)))
}

func (info *BaseInfo) GetArrayFixedSize() int {
	return GoInt(C.g_type_info_get_array_fixed_size((*C.GITypeInfo)(info.ptr)))
}

func (info *BaseInfo) IsZeroTerminated() bool {
	return GoBool(C.g_type_info_is_zero_terminated((*C.GITypeInfo)(info.ptr)))
}

func (info *BaseInfo) GetArrayType() ArrayType {
	return (ArrayType)(C.g_type_info_get_array_type((*C.GITypeInfo)(info.ptr)))
}
