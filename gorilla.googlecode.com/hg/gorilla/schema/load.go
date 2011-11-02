// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// ----------------------------------------------------------------------------
// Type conversion
// ----------------------------------------------------------------------------

// Converts values from one type to another.
type TypeConv func(basic reflect.Value) reflect.Value

// Holds all the type converters.
// The key is the name of the target type
// and the value is the function that is responsible
// for converting from the basic type
// to the dynamic type
var conversionMap = make(map[string]TypeConv)

// AddTypeConverter adds a function to perform conversions from the basic type
// to a more complex type.
//
// This can be used to define properties of types that are "alias"
// to basic types.
//
//     type TString string
//
//     type TMyRec struct {
// 	       Prop1 TString
//     }
//
// In that case, the Load function would fail without a converter
// since a simple "String" cannot be set on a TString property
// So we write the function to perform de conversion
// func ConvStringToString(v reflect.Value) reflect.Value {
//		return reflect.ValueOf(TString(v.String()))
// }
// And register that function for later use
// AddTypeConverter(reflect.TypeOf(TString("")), ConvStringToTString)
// Then the Load function can resolve the conversion.
func AddTypeConverter(rt reflect.Type, conv TypeConv) {
	conversionMap[getTypeId(rt)] = conv
}

// getTypeConverter returns a converter for a value, or nil if there are none.
func getTypeConverter(rt reflect.Type) TypeConv {
	if conv, ok := conversionMap[getTypeId(rt)]; ok {
		return conv
	}
	return nil
}

// ----------------------------------------------------------------------------
// ValueError
// ----------------------------------------------------------------------------

// ValueError stores errors for a single value.
//
// The same value can be validated more than once, so it can have multiple
// errors.
type ValueError struct {
	err []os.Error
}

func (e *ValueError) Errors() []os.Error {
	return e.err
}

func (e *ValueError) Add(err os.Error) {
	if e.err == nil {
		e.err = make([]os.Error, 0)
	}
	e.err = append(e.err, err)
}

func (e *ValueError) String() string {
	if e.err == nil {
		return ""
	}
	return fmt.Sprintf("%v", e.err)
}

// ----------------------------------------------------------------------------
// SchemaError
// ----------------------------------------------------------------------------

// SchemaError stores global errors and validation errors for field values.
//
// Global errors are stored using an empty string key.
type SchemaError struct {
	err map[string][]os.Error
}

func (e *SchemaError) Errors() map[string][]os.Error {
	return e.err
}

func (e *SchemaError) Error(key string) []os.Error {
	if e.err != nil {
		if v, ok := e.err[key]; ok {
			return v
		}
	}
	return nil
}

func (e *SchemaError) Add(err os.Error, key string, index int) {
	if e.err == nil {
		e.err = make(map[string][]os.Error)
	}

	v1, ok := e.err[key]
	if !ok || index+1 > cap(v1) {
		newV := make([]os.Error, index+1)
		if ok {
			copy(newV, v1)
		}
		v1 = newV
	}

	v2 := v1[index]
	if v2 == nil {
		v2 = &ValueError{}
		v1[index] = v2
	}

	(v2.(*ValueError)).Add(err)
	e.err[key] = v1
}

func (e *SchemaError) String() string {
	if e.err == nil {
		return ""
	}
	return fmt.Sprintf("%v", e.err)
}

// ----------------------------------------------------------------------------
// Load, Validate and variants
// ----------------------------------------------------------------------------

// Load fills a struct with form values.
//
// The first parameter must be a pointer to a struct. The second is a map,
// typically url.Values, http.Request.Form or http.Request.MultipartForm.
//
// This function is capable of filling nested structs recursivelly using map
// keys as "paths" in dotted notation.
//
// See the package documentation for a full explanation of the mechanics.
func Load(i interface{}, data map[string][]string) os.Error {
	return loadAndValidate(i, data, nil, nil)
}

// not public yet, but will be once filters and validators are implemented.
func loadAndValidate(i interface{}, data map[string][]string,
filters map[string]string, validators map[string]string) os.Error {
	err := &SchemaError{}
	val := reflect.ValueOf(i)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		err.Add(os.NewError("Interface must be a pointer to struct."), "", 0)
	} else {
		rv := val.Elem()
		for path, values := range data {
			parts := strings.Split(path, ".")
			loadValue(rv, values, parts, path, err)
		}
	}
	if err.String() == "" {
		return nil
	}
	return err
}

// ----------------------------------------------------------------------------
// Internals
// ----------------------------------------------------------------------------

// loadValue sets the value for a path in a struct.
//
// - rv is the current struct being walked.
//
// - values are the ummodified values to be set.
//
// - parts are the remaining path parts to be walked.
//
// - key is the unmodified data key.
//
// - se is the SchemaError instance to save errors.
func loadValue(rv reflect.Value, values, parts []string, key string,
se *SchemaError) {
	spec, err := defaultStructMap.getOrLoad(rv.Type())
	if err != nil {
		// Struct spec could not be loaded.
		se.Add(err, "", 0)
		return
	}

	fieldSpec, ok := spec.fields[parts[0]]
	if !ok {
		// Field doesn't exist.
		return
	}

	parts = parts[1:]
	field := setIndirect(rv.FieldByName(fieldSpec.realName))
	kind := field.Kind()
	if (kind == reflect.Struct || (kind == reflect.Slice && len(parts) > 0) || kind == reflect.Map) == (len(parts) == 0) {
		// Last part can't be a struct or map. Others must be a struct or map.
		return
	}

	var idx string
	if kind == reflect.Map {
		// Get map index.
		idx = parts[0]
		parts = parts[1:]
		if len(parts) > 0 {
			// Last part must be the map index.
			return
		}
	}

	if len(parts) > 0 {
		if kind == reflect.Slice {
			if field.IsNil() {
				slice := reflect.MakeSlice(field.Type(), len(values), len(values))
				field.Set(slice)
			}
			for i := 0; i < len(values); i++ {
				sv := field.Index(i)
				loadValue(sv, values[i:i+1], parts, key, se)
			}
		} else {
			// A struct. Move to next part.
			loadValue(field, values, parts, key, se)
		}
		return
	}

	// Last part: set the value.
	var value reflect.Value
	switch kind {
	case reflect.Bool,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.String,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		value, err = coerce(kind, values[0])
		if err != nil {
			se.Add(err, key, 0)
		} else {
			if conv := getTypeConverter(field.Type()); conv != nil {
				value = conv(value)
			}
			field.Set(value)
		}
	case reflect.Map:
		elem := field.Type().Elem()
		ekind := elem.Kind()
		if field.IsNil() {
			field.Set(reflect.MakeMap(field.Type()))
		}
		value, err = coerce(ekind, values[0])
		if err != nil {
			// Create a zero value to not miss an index.
			value = reflect.New(elem)
			se.Add(err, key, 0)
		}
		if conv := getTypeConverter(elem); conv != nil {
			value = conv(value)
		}
		field.SetMapIndex(reflect.ValueOf(idx), value)
	case reflect.Slice:
		elem := field.Type().Elem()
		ekind := elem.Kind()
		slice := reflect.MakeSlice(field.Type(), 0, 0)
		conv := getTypeConverter(elem)
		for k, v := range values {
			value, err = coerce(ekind, v)
			if err != nil {
				// Create a zero value to not miss an index.
				value = reflect.New(elem)
				se.Add(err, key, k)
			}
			if conv != nil {
				value = conv(value)
			}
			slice = reflect.Append(slice, value)
		}
		field.Set(slice)
	}
	return
}

// coerce coerces basic types from a string to a reflect.Value of a given kind.
func coerce(kind reflect.Kind, value string) (rv reflect.Value, err os.Error) {
	switch kind {
	case reflect.Bool:
		var v bool
		v, err = strconv.Atob(value)
		rv = reflect.ValueOf(v)
	case reflect.Float32:
		var v float32
		v, err = strconv.Atof32(value)
		rv = reflect.ValueOf(v)
	case reflect.Float64:
		var v float64
		v, err = strconv.Atof64(value)
		rv = reflect.ValueOf(v)
	case reflect.Int:
		var v int
		v, err = strconv.Atoi(value)
		rv = reflect.ValueOf(v)
	case reflect.Int8:
		var v int
		v, err = strconv.Atoi(value)
		rv = reflect.ValueOf(int8(v))
	case reflect.Int16:
		var v int
		v, err = strconv.Atoi(value)
		rv = reflect.ValueOf(int16(v))
	case reflect.Int32:
		var v int
		v, err = strconv.Atoi(value)
		rv = reflect.ValueOf(int32(v))
	case reflect.Int64:
		var v int64
		v, err = strconv.Atoi64(value)
		rv = reflect.ValueOf(v)
	case reflect.String:
		rv = reflect.ValueOf(value)
	case reflect.Uint:
		var v uint
		v, err = strconv.Atoui(value)
		rv = reflect.ValueOf(v)
	case reflect.Uint8:
		var v uint
		v, err = strconv.Atoui(value)
		rv = reflect.ValueOf(uint8(v))
	case reflect.Uint16:
		var v uint
		v, err = strconv.Atoui(value)
		rv = reflect.ValueOf(uint16(v))
	case reflect.Uint32:
		var v uint
		v, err = strconv.Atoui(value)
		rv = reflect.ValueOf(uint32(v))
	case reflect.Uint64:
		var v uint64
		v, err = strconv.Atoui64(value)
		rv = reflect.ValueOf(v)
	default:
		err = os.NewError("Unsupported type.")
	}
	return
}

// ----------------------------------------------------------------------------
// structMap
// ----------------------------------------------------------------------------

// Internal map of cached struct specs.
var defaultStructMap = newStructMap()

// structMap caches parsed structSpec's keyed by package+name.
type structMap struct {
	specs map[string]*structSpec
	mutex sync.RWMutex
}

// newStructMap returns a new structMap instance.
func newStructMap() *structMap {
	return &structMap{
		specs: make(map[string]*structSpec),
	}
}

// getByType returns a cached structSpec given a struct type.
//
// It returns nil if the type argument is not a reflect.Struct.
func (m *structMap) getByType(t reflect.Type) (spec *structSpec) {
	if m.specs != nil && t.Kind() == reflect.Struct {
		m.mutex.RLock()
		spec = m.specs[getTypeId(t)]
		m.mutex.RUnlock()
	}
	return
}

// getOrLoad returns a cached structSpec, loading and caching it if needed.
//
// It returns nil if the passed type is not a struct.
func (m *structMap) getOrLoad(t reflect.Type) (spec *structSpec,
err os.Error) {
	if spec = m.getByType(t); spec != nil {
		return spec, nil
	}

	// Lock it for writes until the new type is loaded.
	m.mutex.Lock()
	loaded := make([]string, 0)
	if spec, err = m.load(t, &loaded); err != nil {
		// Roll back loaded structs.
		for _, v := range loaded {
			m.specs[v] = nil, false
		}
		return
	}
	m.mutex.Unlock()

	return
}

// load caches parsed struct metadata.
//
// It is an internal function used by getOrLoad and can't be called directly
// because a write lock is required.
//
// The loaded argument is the list of keys to roll back in case of error.
func (m *structMap) load(t reflect.Type, loaded *[]string) (spec *structSpec,
err os.Error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(os.Error)
		}
	}()

	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, os.NewError("Not a struct.")
	}

	structId := getTypeId(t)
	spec = &structSpec{fields: make(map[string]*structFieldSpec)}
	m.specs[structId] = spec
	*loaded = append(*loaded, structId)

	var toLoad reflect.Type
	uniqueNames := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !isSupportedType(field.Type) {
			continue
		}

		toLoad = nil
		switch field.Type.Kind() {
		case reflect.Map, reflect.Slice:
			et := field.Type.Elem()
			if et.Kind() == reflect.Struct {
				toLoad = et
			}
		case reflect.Struct:
			toLoad = field.Type
		}

		if toLoad != nil {
			// Load nested struct.
			structId = getTypeId(toLoad)
			if m.specs[structId] == nil {
				if _, err = m.load(toLoad, loaded); err != nil {
					return nil, err
				}
			}
		}

		// Use the name defined in the tag, if available.
		name := field.Tag.Get("schema-name")
		if name == "" {
			name = field.Name
		}

		// The name must be unique for the struct.
		for _, uniqueName := range uniqueNames {
			if name == uniqueName {
				return nil, os.NewError("Field names and name tags in a " +
					"struct must be unique.")
			}
		}
		uniqueNames[i] = name

		// Set tags.
		tags := make([]string, 0)
		tagStrings := field.Tag.Get("schema-tags")
		for _, tag := range strings.Split(tagStrings, " ") {
			tag := strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}

		// Finally, set the field.
		spec.fields[name] = &structFieldSpec{
			name:     name,
			realName: field.Name,
			tags:     tags,
		}
	}
	return
}

// ----------------------------------------------------------------------------
// structSpec
// ----------------------------------------------------------------------------

// structSpec stores information from a parsed struct.
//
// It is used to fill a struct with values from a multi-map, checking if keys
// in dotted notation can be translated to a struct field and executing
// filters and conversions.
type structSpec struct {
	fields map[string]*structFieldSpec
}

// ----------------------------------------------------------------------------
// structFieldSpec
// ----------------------------------------------------------------------------

// structFieldSpec stores information from a parsed struct field.
type structFieldSpec struct {
	// Name defined in the field tag, or the real field name.
	name string
	// Real field name as defined in the struct.
	realName string
	// Tags, used to identify filters and validators.
	tags []string
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// getTypeId returns an ID for a struct: package name + "." + struct name.
func getTypeId(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

// isSupportedType returns true for supported field types.
func isSupportedType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if isSupportedBasicType(t) {
		return true
	} else {
		switch t.Kind() {
		case reflect.Slice:
			return true
		case reflect.Struct:
			return true
		case reflect.Map:
			// Only map[string]anyOfTheBaseTypes.
			stringKey := t.Key().Kind() == reflect.String
			if stringKey && isSupportedBasicType(t.Elem()) {
				return true
			}
		}
	}
	return false
}

// isSupportedBasicType returns true for supported basic field types.
//
// Only basic types can be used in maps/slices values.
func isSupportedBasicType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64,
		reflect.String,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return true
	}
	return false
}

// setIndirect resolves a pointer to value, setting it recursivelly if needed.
func setIndirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			ptr := reflect.New(v.Type().Elem())
			v.Set(ptr)
			v = ptr
		}
		v = reflect.Indirect(v)
	}
	return v
}
