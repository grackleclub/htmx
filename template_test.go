package template

import (
	"embed"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var testDir = path.Join("static")

// an embed filesystem to test expected production use cases
//
//go:embed static
var static embed.FS

func init() {
	Strict = true
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func TestReadFS(t *testing.T) {
	t.Run("local FS", func(t *testing.T) {
		entries, err := fs.ReadDir(os.DirFS(testDir), ".")
		require.NoError(t, err)
		require.NotNil(t, entries)

		for _, file := range entries {
			if file.IsDir() {
				t.Logf("dir: %s", file.Name())
				continue
			}
			t.Logf("file: %s", file.Name())
		}
	})

	t.Run("embed FS", func(t *testing.T) {
		entries, err := fs.ReadDir(static, "static")
		require.NoError(t, err)
		require.NotNil(t, entries)

		for _, file := range entries {
			if file.IsDir() {
				t.Logf("dir: %s", file.Name())
				continue
			}
			t.Logf("file: %s", file.Name())
		}
	})
}

func TestWriteTemplate(t *testing.T) {
	type Year struct {
		Title string
		Year  string
	}
	s, err := New(static, "static")
	require.NoError(t, err)

	err = s.WriteTemplate(
		os.Stdout,
		[]string{"static/html/index.html", "static/html/footer.html"},
		Year{Year: "2025", Title: "Hello, World!"},
	)
	require.NoError(t, err)
}
