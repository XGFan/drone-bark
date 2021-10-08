package main

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var ZERO = reflect.Value{}

type DataSource interface {
	Lookup(prefixes []string) (string, bool)
	FindPrefix(prefixes []string) map[string]string
}

func ParseField(dataSource DataSource,
	prefixes []string,
	fieldType reflect.Type,
	fieldName string) (reflect.Value, error) {
	switch fieldType.Kind() {
	case reflect.Bool:
		s, ok := dataSource.Lookup(append(prefixes, fieldName))
		if ok {
			parsed, err := strconv.ParseBool(s)
			return reflect.ValueOf(parsed), err
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s, ok := dataSource.Lookup(append(prefixes, fieldName))
		if ok {
			parsed, err := strconv.ParseInt(s, 0, 64)
			return reflect.ValueOf(parsed).Convert(fieldType), err
		}
	case reflect.Float32, reflect.Float64:
		s, ok := dataSource.Lookup(append(prefixes, fieldName))
		if ok {
			parsed, err := strconv.ParseFloat(s, 64)
			return reflect.ValueOf(parsed).Convert(fieldType), err
		}
	case reflect.Slice:
		s, ok := dataSource.Lookup(append(prefixes, fieldName))
		if ok {
			split := strings.Split(s, ",")
			switch fieldType.Elem().Kind() {
			case reflect.String:
				slice := make([]string, 0)
				for _, s := range split {
					slice = append(slice, s)
				}
				return reflect.ValueOf(slice), nil
			case reflect.Bool:
				slice := make([]bool, 0)
				for _, s := range split {
					parsed, err := strconv.ParseBool(s)
					if err != nil {
						return ZERO, err
					}
					slice = append(slice, parsed)
				}
				return reflect.ValueOf(slice), nil
			case reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64,
				reflect.Uint,
				reflect.Uint8,
				reflect.Uint16,
				reflect.Uint32,
				reflect.Uint64:
				makeSlice := reflect.MakeSlice(fieldType, 0, len(split))
				for _, s := range split {
					parsed, err := strconv.ParseInt(s, 0, 64)
					if err != nil {
						return ZERO, err
					}
					convert := reflect.ValueOf(parsed).Convert(fieldType.Elem())
					makeSlice = reflect.Append(makeSlice, convert)
				}
				return makeSlice, nil
			case reflect.Float32,
				reflect.Float64:
				makeSlice := reflect.MakeSlice(fieldType, 0, len(split))
				for _, s := range split {
					parsed, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return ZERO, err
					}
					convert := reflect.ValueOf(parsed).Convert(fieldType.Elem())
					makeSlice = reflect.Append(makeSlice, convert)
				}
				return makeSlice, nil
			default:
				return ZERO, errors.New(fmt.Sprintf("unsupport field %s with %s", fieldName, fieldType))
			}
		}
	case reflect.Struct:
		indirectValue := reflect.Indirect(reflect.New(fieldType))
		m := dataSource.FindPrefix(append(prefixes, fieldName))
		if len(m) == 0 {
			break
		}
		for i := 0; i < fieldType.NumField(); i++ {
			indirectValueField := indirectValue.Field(i)
			subField := fieldType.Field(i)
			newValue, err := ParseField(dataSource, append(prefixes, fieldName), subField.Type, subField.Name)
			if err != nil {
				return ZERO, err
			}
			if newValue.IsValid() {
				indirectValueField.Set(newValue)
			}
		}
		return indirectValue, nil
	case reflect.Map:
		m := dataSource.FindPrefix(append(prefixes, fieldName))
		if len(m) == 0 {
			break
		}
		if fieldType.Elem().Kind() != reflect.String {
			return ZERO, errors.New(fmt.Sprintf("unsupport filed %s with %s: key should be string, but got %s",
				fieldName, fieldType.Kind(), fieldType.Elem().Kind()))
		}
		switch fieldType.Key().Kind() {
		case reflect.String:
			return reflect.ValueOf(m), nil
		case reflect.Bool:
			makeMap := reflect.MakeMap(fieldType)
			for k, v := range m {
				parsed, err := strconv.ParseBool(v)
				if err != nil {
					return ZERO, err
				}
				makeMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(parsed).Convert(fieldType.Elem()))
			}
			return makeMap, nil
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			makeMap := reflect.MakeMap(fieldType)
			for k, v := range m {
				parsed, err := strconv.ParseInt(v, 0, 64)
				if err != nil {
					return ZERO, err
				}
				makeMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(parsed).Convert(fieldType.Elem()))
			}
			return makeMap, nil
		case reflect.Float32,
			reflect.Float64:
			makeMap := reflect.MakeMap(fieldType)
			for k, v := range m {
				parsed, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return ZERO, err
				}
				makeMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(parsed).Convert(fieldType.Elem()))
			}
			return makeMap, nil
		default:
			return ZERO, errors.New(fmt.Sprintf("unsupport filed %s with %s", fieldName, fieldType.Kind()))
		}
	case reflect.String:
		s, ok := dataSource.Lookup(append(prefixes, fieldName))
		if ok {
			return reflect.ValueOf(s), nil
		}
	case reflect.Ptr:
		beta, err := ParseField(dataSource, prefixes, fieldType.Elem(), fieldName)
		if err != nil {
			return beta, err
		}
		if beta.IsValid() {
			newValue := reflect.New(fieldType.Elem())
			reflect.Indirect(newValue).Set(beta)
			return newValue, err
		}
	default:
		return ZERO, errors.New(fmt.Sprintf("unsupport filed %s with %s", fieldName, fieldType.Kind()))
	}
	return ZERO, nil
}

func Parse(dataSource DataSource, data interface{}, prefixes []string) error {
	typeOfData := reflect.TypeOf(data)
	if typeOfData.Kind() != reflect.Ptr {
		return errors.New(fmt.Sprintf("type should be Ptr, but got %v", typeOfData.Kind()))
	}
	typeOfData = typeOfData.Elem()
	if typeOfData.Kind() != reflect.Struct {
		return errors.New(fmt.Sprintf("underlying type should be Struct, but got %v", typeOfData.Elem().Kind()))
	}
	value := reflect.Indirect(reflect.ValueOf(data))
	for i := 0; i < typeOfData.NumField(); i++ {
		field := typeOfData.Field(i)
		if !field.Anonymous {
			if field.IsExported() {
				newValue, err := ParseField(dataSource, prefixes, field.Type, field.Name)
				if err != nil {
					return err
				}
				if newValue.IsValid() {
					value.Field(i).Set(newValue)
				}
			}
		} else {
			if field.Type.Kind() == reflect.Ptr {
				newValue := reflect.New(field.Type.Elem())
				err := Parse(dataSource, newValue.Interface(), prefixes)
				if err != nil {
					return err
				} else {
					value.Field(i).Set(newValue)
				}
			} else {
				newValue := reflect.New(field.Type)
				err := Parse(dataSource, newValue.Interface(), prefixes)
				if err != nil {
					return err
				} else {
					value.Field(i).Set(reflect.Indirect(newValue))
				}
			}
		}
	}
	return nil
}
