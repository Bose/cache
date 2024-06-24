package utils

import (
	"fmt"
	"reflect"
)

func IsStruct(v interface{}) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		return true
	}
	return false
}

func StructHasField(v interface{}, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	return rv.FieldByName(name).IsValid()
}

func StructSetField(s interface{}, field string, value interface{}) error {
	sv := reflect.ValueOf(s)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	} else {
		return fmt.Errorf("StructSetField: not a pointer")
	}
	if sv.Kind() != reflect.Struct {
		return fmt.Errorf("StructSetField: not a struct")
	}

	f := sv.FieldByName(field)
	if f.IsValid() {
		// can only change if addressable and is an exported field
		if f.CanSet() {
			if f.Kind() == reflect.ValueOf(value).Kind() {

				f.Set(reflect.ValueOf(value))

			} else {
				return fmt.Errorf("StructSetField: field and value types dont match")
			}
		} else {
			return fmt.Errorf("StructSetField: cant change field: %s", field)
		}
	} else {
		return fmt.Errorf("StructSetField: field not found in struct %s", field)
	}
	return nil
}

// StructGetFieldPtr requires a pointer to a struct, and a string of a field name inside that struct type.
// If the field exists, can be addressed, and is exported a pointer will be returned for the field INSIDE the struct.
// Otherwise an error is returned.
func StructGetFieldPtr(s interface{}, field string) (interface{}, error) {
	sv := reflect.ValueOf(s)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	} else {
		return nil, fmt.Errorf("StructGetFieldPtr: not a pointer")
	}
	if sv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("StructGetFieldPtr: not a struct")
	}

	f := sv.FieldByName(field)
	if f.IsValid() {
		if f.CanAddr() && f.CanSet(){
			return f.Addr().Interface(), nil
		} else {
			return nil, fmt.Errorf("StructGetFieldPtr: field cant be addrressed or is unexported")
		}
	} else {
		return nil, fmt.Errorf("StructGetFieldPtr: field doesn't exist")
	}
	return nil, fmt.Errorf("StructGetFieldPtr: unknown error")
}

// StructToSerializedArgs requires a pointer to a struct. it returns a list of field names immediately followed by their
// Serialized value.
func StructToSerializedArgs(s interface{}) ([]interface{}, error) {
	sv := reflect.ValueOf(s)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	} else {
		return nil, fmt.Errorf("StructToSerializedArgs: not a pointer")
	}
	if sv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("StructToSerializedArgs: not a struct")
	}
	serializedArgs := []interface{}{}
	st := sv.Type()
	for i := 0; i < st.NumField(); i++ {
		f := sv.Field(i)
		// We only want exported fields, since unexported will panic. However, we want to allow unexported in the struct
		// so the callers can have unexported fields for their own use cases.
		if f.CanInterface() {
			tf := st.Field(i)
			name := tf.Name
			srlzdValue, err := Serialize(f.Interface())
			if err != nil {
				return nil, fmt.Errorf("StructToSerializedArgs: failed to serialize: %s", err.Error())
			}
			serializedArgs = append(serializedArgs, name, srlzdValue)
		}
	}

	return serializedArgs, nil
}