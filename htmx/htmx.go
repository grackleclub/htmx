package htmx

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"text/template"
)

var strictTemplateChecking = true

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

type x struct {
	// filesystem fs.FS
	// staticDir  string
	dir []fs.DirEntry
}

func New(filesystem fs.FS, staticDir string) (*x, error) {
	dir, err := fs.ReadDir(filesystem, staticDir)
	if err != nil {
		return nil, fmt.Errorf("open static dir: %w", err)
	}
	return &x{dir: dir}, nil
}

// serveHtmx dynamically serves htmx components based on the path
func serveHtmx(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// writeTemplate executes a set of templates (defined by their path)
// and injects data from any struct, writing the output to the response writer.
//
// When strictTemplateChecking is true, it will re-render the template with data,
// ensuring that all values from the passe struct that can be refleted as strings
// are present in the rendered template.
func writeTemplate(w io.Writer, filesystem fs.FS, templatePaths []string, data interface{}) error {
	tmpl, err := template.ParseFS(filesystem, templatePaths...)
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
	len := buf.Len()
	if len == 0 {
		return fmt.Errorf("template output is empty")
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
