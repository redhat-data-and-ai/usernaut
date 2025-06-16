package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// MapToStruct populates a struct with values from a map using json tags
// target must be a pointer to a struct
func MapToStruct(data map[string]interface{}, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer to a struct")
	}

	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return errors.New("target must point to a struct")
	}

	targetType := targetValue.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// Remove json options like omitempty
		if idx := strings.Index(tag, ","); idx != -1 {
			tag = tag[:idx]
		}

		value, ok := data[tag]
		if !ok {
			continue
		}

		if err := setField(fieldValue, value); err != nil {
			return fmt.Errorf("error setting field %s: %w", field.Name, err)
		}
	}

	return nil
}

func setField(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		return setString(field, value)
	case reflect.Bool:
		return setBool(field, value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(field, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setUint(field, value)
	case reflect.Float32, reflect.Float64:
		return setFloat(field, value)
	case reflect.Struct:
		return setStruct(field, value)
	case reflect.Slice:
		return setSlice(field, value)
	case reflect.Map:
		return setMap(field, value)
	}

	return fmt.Errorf("unsupported type: %s", field.Kind())
}

func setString(field reflect.Value, value interface{}) error {
	if str, ok := value.(string); ok {
		field.SetString(str)
		return nil
	}
	field.SetString(fmt.Sprintf("%v", value))
	return nil
}

func setBool(field reflect.Value, value interface{}) error {
	if b, ok := value.(bool); ok {
		field.SetBool(b)
		return nil
	}

	if str, ok := value.(string); ok {
		b, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		field.SetBool(b)
		return nil
	}

	return fmt.Errorf("cannot convert %T to bool", value)
}

func setInt(field reflect.Value, value interface{}) error {
	var intValue int64

	switch v := value.(type) {
	case int:
		intValue = int64(v)
	case int8:
		intValue = int64(v)
	case int16:
		intValue = int64(v)
	case int32:
		intValue = int64(v)
	case int64:
		intValue = v
	case float32:
		intValue = int64(v)
	case float64:
		intValue = int64(v)
	case string:
		var err error
		intValue, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot convert %T to int", value)
	}

	field.SetInt(intValue)
	return nil
}

func setUint(field reflect.Value, value interface{}) error {
	var uintValue uint64

	switch v := value.(type) {
	case uint:
		uintValue = uint64(v)
	case uint8:
		uintValue = uint64(v)
	case uint16:
		uintValue = uint64(v)
	case uint32:
		uintValue = uint64(v)
	case uint64:
		uintValue = v
	case int:
		uintValue = uint64(v)
	case string:
		var err error
		uintValue, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot convert %T to uint", value)
	}

	field.SetUint(uintValue)
	return nil
}

func setFloat(field reflect.Value, value interface{}) error {
	var floatValue float64

	switch v := value.(type) {
	case float32:
		floatValue = float64(v)
	case float64:
		floatValue = v
	case int:
		floatValue = float64(v)
	case string:
		var err error
		floatValue, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot convert %T to float", value)
	}

	field.SetFloat(floatValue)
	return nil
}

func setStruct(field reflect.Value, value interface{}) error {
	mapValue, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot set struct field with %T", value)
	}

	newValue := reflect.New(field.Type())
	if err := MapToStruct(mapValue, newValue.Interface()); err != nil {
		return err
	}

	field.Set(newValue.Elem())
	return nil
}

func setSlice(field reflect.Value, value interface{}) error {
	sliceValue, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("cannot set slice field with %T", value)
	}

	slice := reflect.MakeSlice(field.Type(), len(sliceValue), len(sliceValue))

	for i, item := range sliceValue {
		if err := setField(slice.Index(i), item); err != nil {
			return err
		}
	}

	field.Set(slice)
	return nil
}

func setMap(field reflect.Value, value interface{}) error {
	mapValue, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot set map field with %T", value)
	}

	mapType := field.Type()
	if mapType.Key().Kind() != reflect.String {
		return fmt.Errorf("only string keys are supported for maps")
	}

	resultMap := reflect.MakeMap(mapType)

	for k, v := range mapValue {
		elemValue := reflect.New(mapType.Elem()).Elem()
		if err := setField(elemValue, v); err != nil {
			return err
		}
		resultMap.SetMapIndex(reflect.ValueOf(k), elemValue)
	}

	field.Set(resultMap)
	return nil
}
