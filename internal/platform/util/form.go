package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// DecodeBody parses request body — supports JSON, multipart/form-data, and urlencoded.
// Usage: util.DecodeBody(r, &req)
func DecodeBody(r *http.Request, v any) error {
	ct := r.Header.Get("Content-Type")

	// Form-data or URL-encoded
	if strings.HasPrefix(ct, "multipart/form-data") || strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			return fmt.Errorf("parse form: %w", err)
		}
		return populateStruct(r, v)
	}

	// Default: JSON
	return json.NewDecoder(r.Body).Decode(v)
}

func populateStruct(r *http.Request, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer to struct")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a pointer to struct")
	}

	typ := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		// Get json tag name as form field name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		// Remove omitempty, etc.
		fieldName := strings.Split(jsonTag, ",")[0]

		// Get form value
		formValue := r.FormValue(fieldName)
		if formValue == "" {
			continue
		}

		// Set value based on type
		if err := setFieldValue(field, formValue); err != nil {
			return fmt.Errorf("field %s: %w", fieldName, err)
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer: %s", value)
		}
		field.SetInt(n)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean: %s", value)
		}
		field.SetBool(b)

	case reflect.Ptr:
		// Handle pointer types (e.g., *string, *int, *time.Time)
		if value == "" {
			field.SetZero()
			return nil
		}
		elemType := field.Type().Elem()
		newVal := reflect.New(elemType)

		switch elemType.Kind() {
		case reflect.String:
			newVal.Elem().SetString(value)
		case reflect.Int:
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid integer: %s", value)
			}
			newVal.Elem().SetInt(int64(n))
		case reflect.Struct:
			// Handle *time.Time (date only)
			if elemType == reflect.TypeOf(time.Time{}) {
				t, err := time.Parse("2006-01-02", value)
				if err != nil {
					// Try RFC3339
					t, err = time.Parse(time.RFC3339, value)
					if err != nil {
						return fmt.Errorf("invalid date: %s", value)
					}
				}
				newVal.Elem().Set(reflect.ValueOf(t))
			}
		}
		field.Set(newVal)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}
