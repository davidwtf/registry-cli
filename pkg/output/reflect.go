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
	return reflectPrintStruct(stdout, "", val, -indentStep, 0)
}

func reflectPrintValue(stdout io.Writer, name string, val reflect.Value, indent, nameWidth int) error {
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
		if err := reflectPrintSimpleValue(stdout, name, val, indent, nameWidth); err != nil {
			return err
		}
	case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:

		switch kind {
		case reflect.Array, reflect.Slice:
			if err := reflectPrintArray(stdout, name, val, indent); err != nil {
				return err
			}
		case reflect.Struct:
			if err := reflectPrintStruct(stdout, name, val, indent, nameWidth); err != nil {
				return err
			}
		case reflect.Map:
			if err := reflectPrintMap(stdout, name, val, indent); err != nil {
				return err
			}
		}
	case reflect.Interface:
		return reflectPrintValue(stdout, name, val.Elem(), indent, nameWidth)
	}
	return nil
}

func getPadding(name string, nameWidth int) string {
	if nameWidth > 0 {
		n := nameWidth - len(name)
		if n > 0 {
			return strings.Repeat(" ", n)
		}
	}
	return ""
}

func reflectPrintStruct(stdout io.Writer, name string, val reflect.Value, indent, nameWidth int) error {
	typ := val.Type()
	if name != "" {
		if typ == reflect.TypeOf(time.Time{}) {
			valstr := val.MethodByName("Format").Call([]reflect.Value{
				reflect.ValueOf(time.RFC3339),
			})[0].String()
			if _, err := fmt.Fprintf(stdout, "%s%s: %s%s\n", strings.Repeat(" ", indent), name, getPadding(name, nameWidth), valstr); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(stdout, "%s%s:\n", strings.Repeat(" ", indent), name); err != nil {
				return err
			}
		}
	}
	subNameWidth := 0
	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
		if !fld.IsExported() {
			continue
		}
		width := len(fld.Name)
		if width > nameWidth {
			subNameWidth = width
		}
	}

	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
		if !fld.IsExported() {
			continue
		}
		it := val.Field(i)
		if err := reflectPrintValue(stdout, fld.Name, it, indent+indentStep, subNameWidth); err != nil {
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
		if err := reflectPrintValue(stdout, strconv.Itoa(i), it, indent+indentStep, 0); err != nil {
			return err
		}
	}
	return nil
}
func reflectPrintMap(stdout io.Writer, name string, val reflect.Value, indent int) error {
	if _, err := fmt.Fprintf(stdout, "%s%s:\n", strings.Repeat(" ", indent), name); err != nil {
		return err
	}
	nameWidth := 0
	for _, key := range val.MapKeys() {
		name := val.MapIndex(key).String()
		if len(name) > nameWidth {
			nameWidth = len(name)
		}
	}
	for _, key := range val.MapKeys() {
		it := val.MapIndex(key)
		if err := reflectPrintValue(stdout, key.String(), it, indent+indentStep, nameWidth); err != nil {
			return err
		}
	}
	return nil
}

func reflectPrintSimpleValue(stdout io.Writer, name string, val reflect.Value, indent, nameWidth int) error {
	var err error
	var valstr string
	isSizeField := strings.HasSuffix(strings.ToLower(name), "size")
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
		if isSizeField {
			valstr = customSize("%.4g%s", float64(val.Uint()), 1024.0, binaryAbbrs)
		} else {
			valstr = strconv.FormatUint(val.Uint(), 10)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if isSizeField {
			valstr = customSize("%.4g%s", float64(val.Int()), 1024.0, binaryAbbrs)
		} else {
			valstr = strconv.FormatInt(val.Int(), 10)
		}
	case reflect.Complex64:
		valstr = strconv.FormatComplex(val.Complex(), 'g', 8, 64)
	case reflect.Complex128:
		valstr = strconv.FormatComplex(val.Complex(), 'g', 8, 128)
	}

	_, err = fmt.Fprintf(stdout, "%s%s: %s%s\n", strings.Repeat(" ", indent), name, getPadding(name, nameWidth), valstr)

	return err
}
