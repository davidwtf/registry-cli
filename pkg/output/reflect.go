package output

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const indentStep = 2

func PrintStruct(stdout io.Writer, obj interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(obj))
	return reflectPrintStruct(stdout, "", val, -indentStep)
}

func reflectPrintValue(stdout io.Writer, name string, val reflect.Value, indent int) error {
	kind := val.Kind()
	if kind == reflect.Pointer || kind == reflect.Map || kind == reflect.Slice {
		if val.IsNil() {
			return nil
		}
		val = reflect.Indirect(val)
	}

	kind = val.Kind()
	switch kind {
	case reflect.Bool, reflect.String,
		reflect.Float32, reflect.Float64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Complex64, reflect.Complex128:
		if err := reflectPrintSimpleValue(stdout, name, val, indent); err != nil {
			return err
		}
	case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:

		switch kind {
		case reflect.Array, reflect.Slice:
			if err := reflectPrintArray(stdout, name, val, indent); err != nil {
				return err
			}
		case reflect.Struct:
			if err := reflectPrintStruct(stdout, name, val, indent); err != nil {
				return err
			}
		case reflect.Map:
			if err := reflectPrintMap(stdout, name, val, indent); err != nil {
				return err
			}
		}
	case reflect.Interface:
		return reflectPrintValue(stdout, name, val.Elem(), indent)
	}
	return nil
}

func reflectPrintStruct(stdout io.Writer, name string, val reflect.Value, indent int) error {
	typ := val.Type()
	if name != "" {
		if typ == reflect.TypeOf(time.Time{}) {
			valstr := val.MethodByName("Format").Call([]reflect.Value{
				reflect.ValueOf(time.RFC3339),
			})[0].String()
			if _, err := fmt.Fprintf(stdout, "%s%s: %s\n", strings.Repeat(" ", indent), name, valstr); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(stdout, "%s%s:\n", strings.Repeat(" ", indent), name); err != nil {
				return err
			}
		}
	}
	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
		if !fld.IsExported() {
			continue
		}
		it := val.Field(i)
		if err := reflectPrintValue(stdout, fld.Name, it, indent+indentStep); err != nil {
			return err
		}
	}
	return nil
}

func reflectPrintArray(stdout io.Writer, name string, val reflect.Value, indent int) error {
	if _, err := fmt.Fprintf(stdout, "%s%s:\n", strings.Repeat(" ", indent), name); err != nil {
		return err
	}
	for i := 0; i < val.Len(); i++ {
		it := val.Index(i)
		if err := reflectPrintValue(stdout, strconv.Itoa(i), it, indent+indentStep); err != nil {
			return err
		}
	}
	return nil
}
func reflectPrintMap(stdout io.Writer, name string, val reflect.Value, indent int) error {
	if _, err := fmt.Fprintf(stdout, "%s%s:\n", strings.Repeat(" ", indent), name); err != nil {
		return err
	}
	for _, key := range val.MapKeys() {
		it := val.MapIndex(key)
		if err := reflectPrintValue(stdout, key.String(), it, indent+indentStep); err != nil {
			return err
		}
	}
	return nil
}

func reflectPrintSimpleValue(stdout io.Writer, name string, val reflect.Value, indent int) error {
	var err error
	var valstr string
	switch val.Kind() {
	case reflect.Bool:
		valstr = strconv.FormatBool(val.Bool())
	case reflect.String:
		valstr = val.String()
	case reflect.Float32:
		valstr = strconv.FormatFloat(val.Float(), 'g', 8, 32)
	case reflect.Float64:
		valstr = strconv.FormatFloat(val.Float(), 'g', 8, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		valstr = strconv.FormatUint(val.Uint(), 10)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valstr = strconv.FormatInt(val.Int(), 10)
	case reflect.Complex64:
		valstr = strconv.FormatComplex(val.Complex(), 'g', 8, 64)
	case reflect.Complex128:
		valstr = strconv.FormatComplex(val.Complex(), 'g', 8, 128)
	}

	_, err = fmt.Fprintf(stdout, "%s%s: %s\n", strings.Repeat(" ", indent), name, valstr)

	return err
}
