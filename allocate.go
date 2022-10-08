/*

The allocate library provides helper functions for allocation/initializing structs.

@author Colorado Reed (colorado at dotdashpay ... com)

This library was seeded by this discussion in the golang-nuts mailing list:
https://groups.google.com/forum/#!topic/golang-nuts/Wd9jiZswwMU
*/

package allocate

import (
	"fmt"
	"reflect"
	"strings"
)

// MustZero will panic instead of return error.
func MustZero[S any](inputIntf S) {
	err := Zero(inputIntf)
	if err != nil {
		panic(err)
	}
}

// MustZeroNested will panic instead of return error.
func MustZeroNested[S any](inputIntf S, fields string) {
	err := ZeroNested(inputIntf, fields)
	if err != nil {
		panic(err)
	}
}

// MustSetNested will panic instead of return error.
func MustSetNested[S, V any](inputIntf S, fields string, value V) {
	err := SetNested(inputIntf, fields, value)
	if err != nil {
		panic(err)
	}
}

// Zero allocates an input structure such that all pointer fields
// are fully allocated, i.e. rather than having a nil value,
// the pointer contains a pointer to an initialized value,
// e.g. an *int field will be a pointer to 0 instead of a nil pointer.
//
// Zero does not allocate private fields.
func Zero[S any](inputIntf S) error {
	indirectVal := reflect.Indirect(reflect.ValueOf(inputIntf))

	if err := structCanSet(indirectVal); err != nil {
		return err
	}

	// allocate each of the structs fields
	for i := 0; i < indirectVal.NumField(); i++ {
		if err := zeroField(indirectVal.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

// ZeroNested is like Zero but only allocates the nested field.
// The fields should be a path split by ".", e.g. "Spec.Template.Resources".
// Returns error if the nested field is not found.
func ZeroNested[S any](inputIntf S, fields string) error {
	field, err := getNested(inputIntf, fields)
	if err != nil {
		return err
	}

	if err := zeroField(field); err != nil {
		return err
	}

	return nil
}

// SetNested is like ZeroNested but can assign a value.
func SetNested[S, V any](inputIntf S, fields string, value V) error {
	field, err := getNested(inputIntf, fields)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(value))

	return nil
}

func getNested(inputIntf any, fields string) (reflect.Value, error) {
	nestedFields := strings.Split(fields, ".")
	if len(nestedFields) > 0 && nestedFields[0] == "" {
		nestedFields = nestedFields[1:]
	}

	input := reflect.ValueOf(inputIntf)

	// find the nested field
	for i, fieldName := range nestedFields {
		indirectVal := reflect.Indirect(input)

		if err := structCanSet(indirectVal); err != nil {
			return reflect.Value{}, err
		}

		input = indirectVal.FieldByName(fieldName)
		if !input.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %s not found", nestedPath(nestedFields[:i+1]))
		}

		if input.Kind() == reflect.Ptr && input.IsNil() {
			if input.CanSet() {
				input.Set(reflect.New(input.Type().Elem()))
			}
		}
	}

	return input, nil
}

func structCanSet(input reflect.Value) error {
	if !input.CanSet() {
		return fmt.Errorf("Input interface is not addressable (can't Set the memory address): %#v",
			input)
	}
	if input.Kind() != reflect.Struct {
		return fmt.Errorf("currently only works with [pointers to] structs, not type %v",
			input.Kind())
	}
	return nil
}

func nestedPath(fields []string) string {
	return "." + strings.Join(fields, ".")
}

func zeroField(field reflect.Value) (err error) {
	// pre-allocate pointer fields
	if field.Kind() == reflect.Ptr && field.IsNil() {
		if field.CanSet() {
			field.Set(reflect.New(field.Type().Elem()))
		}
	}

	indirectField := reflect.Indirect(field)
	switch indirectField.Kind() {
	case reflect.Slice:
		indirectField.Set(reflect.MakeSlice(indirectField.Type(), 0, 0))
	case reflect.Map:
		indirectField.Set(reflect.MakeMap(indirectField.Type()))
	case reflect.Struct:
		// recursively allocate each of the structs embedded fields
		if field.Kind() == reflect.Ptr {
			err = Zero(field.Interface())
		} else {
			// field of Struct can always use field.Addr()
			fieldAddr := field.Addr()
			if fieldAddr.CanInterface() {
				err = Zero(fieldAddr.Interface())
			} else {
				err = fmt.Errorf("struct field can't interface, %#v", fieldAddr)
			}
		}
	}
	return
}

// TODO(cjrd)
// Add an allocate.Random() function that assigns random values rather than nil values
