package template

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"text/template"
)

var (
	Strict         bool // Strict template checking
	ErrValidation  = fmt.Errorf("template validation failed")
	ErrMissingData = fmt.Errorf("source data value not present")
)

// Assets represents a static directory which contains assets and templates
type Assets struct {
	dir []fs.DirEntry
	fs  fs.FS
}

// New provides a new Assets object when given a filesystem and a directory name.
// Methods are provided to serve htmx components and write templates.
func New(filesystem fs.FS, staticDir string) (*Assets, error) {
	dir, err := fs.ReadDir(filesystem, staticDir)
	if err != nil {
		return nil, fmt.Errorf("open static dir: %w", err)
	}
	return &Assets{dir: dir, fs: filesystem}, nil
}

// serveHtmx dynamically serves htmx components based on the path
func serveHtmx(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// WriteTemplate executes a set of templates (defined by their path)
// and injects data from any struct, writing the output to the provided writer.
//
// When Strict is true, buffer will be checked for presence of all data values,
// and if any are missing, return ErrMissingDatato allow for errors.Is() comparison.
//
// Strict checking may incur performance penalties.
func (h *Assets) WriteTemplate(w io.Writer, templatePaths []string, data interface{}) error {
	tmpl, err := template.ParseFS(h.fs, templatePaths...) // todo can we not keep the filesystem obj since it's already parsed in New?
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	// write to buffer first to allow inspection
	// because if a child template is called before a parent template,
	// the output will be empty
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	len := buf.Len()
	if len == 0 {
		return fmt.Errorf("template output is empty")
	}

	if Strict {
		content, err := parseContent(data)
		if err != nil {
			return fmt.Errorf("strict check, parse content: %w", err)
		}
		for k, v := range content {
			exptectedValue, ok := v.(string)
			if !ok {
				slog.Warn("skipping template validation for non-string value",
					"key", k,
					"value", v,
				)
				continue
			}
			if !strings.Contains(buf.String(), exptectedValue) {
				return fmt.Errorf("key %q: %w: %w", k, ErrMissingData, ErrValidation)
			}
			slog.Debug("data value exists in rendered template", "key", k)
		}
	}

	b, err := buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("write template to response: %w", err)
	}
	slog.Debug("template(s) executed",
		"bytes_read", len,
		"bytes_written", b,
		"templates", templatePaths,
		"data", data, // TODO exclude data after testing
	)

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
