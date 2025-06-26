package reader

import (
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// parseDateTime tries multiple datetime formats to handle inconsistent data
func parseDateTime(value string, primaryFormat string) (time.Time, error) {
	// Try primary format first
	if t, err := time.Parse(primaryFormat, value); err == nil {
		return t, nil
	}

	// Common datetime formats to try as fallbacks
	fallbackFormats := []string{
		"2006-01-02 15:04:05.000000",  // with microseconds
		"2006-01-02 15:04:05",         // without microseconds
		"2006-01-02T15:04:05.000000Z", // ISO with microseconds
		"2006-01-02T15:04:05Z",        // ISO without microseconds
		"2006-01-02T15:04:05",         // ISO local
		time.RFC3339,                  // RFC3339
		time.RFC3339Nano,              // RFC3339 with nanoseconds
	}

	for _, format := range fallbackFormats {
		if format == primaryFormat {
			continue // Skip if already tried
		}
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime value '%s' with any known format", value)
}

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
	if !field.IsValid() {
		return fmt.Errorf("invalid field path: %s", path[len(path)-1])
	}
	if !field.CanSet() {
		return fmt.Errorf("cannot set field %s", path)
	}

	switch fieldType {
	case "string":
		field.SetString(value)
	case "int":
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse int value '%s': %w", value, err)
		}
		field.SetInt(intVal)
	case "float":
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float value '%s': %w", value, err)
		}
		field.SetFloat(floatVal)
	case "bool":
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse bool value '%s': %w", value, err)
		}
		field.SetBool(boolVal)
	case "date":
		t, err := time.Parse("2006-01-02", value)
		if err != nil {
			return fmt.Errorf("failed to parse date value '%s': %w", value, err)
		}
		field.Set(reflect.ValueOf(t))
	case "datetime":
		t, err := parseDateTime(value, dateFormat)
		if err != nil {
			return fmt.Errorf("failed to parse datetime value '%s': %w", value, err)
		}
		field.Set(reflect.ValueOf(t))
	case "uuid":
		id, err := uuid.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse uuid value '%s': %w", value, err)
		}
		field.Set(reflect.ValueOf(id))
	case "url":
		u, err := url.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse url value '%s': %w", value, err)
		}
		field.Set(reflect.ValueOf(*u))
	default:
		return fmt.Errorf("unsupported type: %s", fieldType)
	}
	return nil
}

func SetFlatField(obj reflect.Value, path string, value string, fieldType string, dateFormat string) error {
	field := obj.FieldByName(path)

	if !field.IsValid() {
		return fmt.Errorf("invalid field path: %s", path)
	}
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}
	if !field.CanSet() {
		return fmt.Errorf("cannot set field %s", path)
	}

	switch fieldType {
	case "string":
		if field.Kind() != reflect.String {
			return fmt.Errorf("field %s is not a string", path)
		}
		field.SetString(value)

	case "int":
		if field.Kind() != reflect.Int && field.Kind() != reflect.Int64 && field.Kind() != reflect.Int32 {
			return fmt.Errorf("field %s is not an integer type", path)
		}
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse int value '%s': %w", value, err)
		}
		field.SetInt(intVal)

	case "float":
		if field.Kind() != reflect.Float64 && field.Kind() != reflect.Float32 {
			return fmt.Errorf("field %s is not a float type", path)
		}
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float value '%s': %w", value, err)
		}
		field.SetFloat(floatVal)

	case "bool":
		if field.Kind() != reflect.Bool {
			return fmt.Errorf("field %s is not a bool", path)
		}
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse bool value '%s': %w", value, err)
		}
		field.SetBool(boolVal)

	case "date":
		if field.Type() != reflect.TypeOf(time.Time{}) {
			return fmt.Errorf("field %s is not time.Time", path)
		}
		t, err := time.Parse("2006-01-02", value)
		if err != nil {
			return fmt.Errorf("failed to parse date value '%s': %w", value, err)
		}
		field.Set(reflect.ValueOf(t))

	case "datetime":
		if field.Type() != reflect.TypeOf(time.Time{}) {
			return fmt.Errorf("field %s is not time.Time", path)
		}
		t, err := parseDateTime(value, dateFormat)
		if err != nil {
			return fmt.Errorf("failed to parse datetime value '%s': %w", value, err)
		}
		field.Set(reflect.ValueOf(t))

	case "uuid":
		if field.Type() != reflect.TypeOf(uuid.UUID{}) {
			return fmt.Errorf("field %s is not uuid.UUID", path)
		}
		id, err := uuid.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse uuid value '%s': %w", value, err)
		}
		field.Set(reflect.ValueOf(id))

	case "url":
		if field.Type() != reflect.TypeOf(url.URL{}) {
			return fmt.Errorf("field %s is not url.URL", path)
		}
		u, err := url.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse url value '%s': %w", value, err)
		}
		field.Set(reflect.ValueOf(*u))

	default:
		return fmt.Errorf("unsupported field type: %s", fieldType)
	}

	return nil
}
