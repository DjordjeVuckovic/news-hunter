package reader

import (
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"reflect"
	"time"
)

func SetNestedField(obj reflect.Value, path []string, value string, fieldType string, dateFormat string) error {
	for i := 0; i < len(path)-1; i++ {
		obj = obj.FieldByName(path[i])
		if !obj.IsValid() {
			return fmt.Errorf("invalid field path: %s", path[i])
		}
		if obj.Kind() == reflect.Pointer {
			if obj.IsNil() {
				obj.Set(reflect.New(obj.Type().Elem()))
			}
			obj = obj.Elem()
		}
	}
	field := obj.FieldByName(path[len(path)-1])
	if !field.CanSet() {
		return fmt.Errorf("cannot set field %s", path)
	}

	switch fieldType {
	case "string":
		field.SetString(value)
	case "uuid":
		id, err := uuid.Parse(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(id))
	case "datetime":
		t, err := time.Parse(dateFormat, value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(t))
	case "url":
		u, err := url.Parse(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(*u))
	// Add more types here...
	default:
		return fmt.Errorf("unsupported type: %s", fieldType)
	}
	return nil
}

func SetFlatField(obj reflect.Value, path string, value string, fieldType string, dateFormat string) error {
	obj = obj.FieldByName(path)
	if !obj.IsValid() {
		return fmt.Errorf("invalid field path: %s", path)
	}
	if obj.Kind() == reflect.Pointer {
		if obj.IsNil() {
			obj.Set(reflect.New(obj.Type().Elem()))
		}
		obj = obj.Elem()
	}

	field := obj
	if !field.CanSet() {
		return fmt.Errorf("cannot set field %s", path)
	}

	switch fieldType {
	case "string":
		field.SetString(value)
	case "uuid":
		id, err := uuid.Parse(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(id))
	case "datetime":
		t, err := time.Parse(dateFormat, value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(t))
	case "url":
		u, err := url.Parse(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(*u))
	default:
		return fmt.Errorf("unsupported type: %s", fieldType)
	}
	return nil
}
