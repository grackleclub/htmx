package trash

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
)

var strictTemplateChecking = true

// writeTemplate executes a set of templates (defined by their path)
// and injects data from any struct, writing the output to the response writer.
//
// When strictTemplateChecking is true, it will re-render the template with data,
// ensuring that all values from the passe struct that can be refleted as strings
// are present in the rendered template.
func writeTemplate(w http.ResponseWriter, templatePaths []string, data interface{}) error {
	tmpl, err := template.ParseFS(os.DirFS("."), templatePaths...)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	// write to buffer first to allow inspection
	// because if a child template is called before a parent template,
	// the output will be empty
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	if buf.Len() == 0 {
		return fmt.Errorf("template output is empty")
	}
	b, err := buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("write template to response: %w", err)
	}
	slog.Debug("template(s) executed",
		"bytes_read", buf.Len(),
		"bytes_written", b,
		"templates", templatePaths,
		"data", data, // TODO exclude data after testing
	)

	// validate
	if strictTemplateChecking {
		// Only because I don't know how to tee a buffer to two outputs,
		// here we parse the content again to validate that the template
		// contains all the values in the data structure.
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, data)
		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		if data == nil {
			slog.Warn("skipping template validation because data is nil")
			return nil
		}
		templateContents := buf.String()
		content, err := parseContent(data)
		if err != nil {
			return fmt.Errorf("failed to parse content: %w", err)
		}
		for k, v := range content {
			exptectedValue, ok := v.(string)
			if !ok {
				slog.Warn("skipping template validation for non-string value", "key", k, "value", v)
				continue
			}
			if !strings.Contains(templateContents, exptectedValue) {
				return fmt.Errorf("template contents do not contain value for key '%s'", k)
			}
			slog.Debug("data value exists in rendered template", "key", k)
		}
	}
	return nil
}

// parseContent parses a nested data structure into a flat map.
// Used for testing template parsing.
func parseContent(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := parseRecursive(data, "", result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// parseRecursive parses a nested data structure into a flat map with only final values.
// Used for testing template parsing.
func parseRecursive(data interface{}, prefix string, result map[string]interface{}) error {
	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Ptr:
		// If it's a pointer, get the element it points to
		return parseRecursive(val.Elem().Interface(), prefix, result)
	case reflect.Struct:
		// If it's a struct, iterate over its fields
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fieldValue := val.Field(i).Interface()
			key := prefix + field.Name
			err := parseRecursive(fieldValue, key+".", result)
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		// If it's a map, iterate over its keys and values
		for _, key := range val.MapKeys() {
			mapValue := val.MapIndex(key).Interface()
			mapKey := fmt.Sprintf("%s%v", prefix, key)
			err := parseRecursive(mapValue, mapKey+".", result)
			if err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		// If it's a slice or array, iterate over its elements
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			key := fmt.Sprintf("%s[%d]", prefix, i)
			err := parseRecursive(elem, key+".", result)
			if err != nil {
				return err
			}
		}
	default:
		// For other types, add the value to the map
		result[prefix] = val.Interface()
	}
	return nil
}
