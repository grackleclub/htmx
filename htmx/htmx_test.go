package htmx

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/htmx" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Make a request to the test server
	resp, err := http.Get(ts.URL + "/htmx")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
		}
	}
}

func TestNew(t *testing.T) {
	// Create a new htmx component
	c, err := New(os.DirFS(path.Join("..")), "static")
	require.NoError(t, err)
	require.NotNil(t, c)
	for _, file := range c.dir {
		if file.IsDir() {
			t.Logf("dir: %s", file.Name())
			continue
		}
		t.Logf("file: %s", file.Name())
	}
}

func TestWriteTemplate(t *testing.T) {
	type Year struct {
		Title string
		Year  string
	}
	err := writeTemplate(
		os.Stdout,
		os.DirFS(path.Join("..")),
		[]string{"static/html/index.html", "static/html/footer.html"},
		Year{Year: "2025", Title: "Hello, World!"},
	)
	require.NoError(t, err)
}
