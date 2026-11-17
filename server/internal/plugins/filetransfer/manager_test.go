package filetransfer

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/internal/session"
	"github.com/m1k1o/neko/server/pkg/auth"
	"github.com/m1k1o/neko/server/pkg/types"
)

type uploadFile struct {
	name    string
	content string
}

func newTestManager(t *testing.T) (*Manager, types.Session) {
	t.Helper()

	sessions := session.New(&config.Session{})
	sess, _, err := sessions.Create("admin", types.MemberProfile{IsAdmin: true})
	if err != nil {
		t.Fatalf("could not create session: %v", err)
	}

	return NewManager(sessions, &Config{Enabled: true, RootDir: t.TempDir()}), sess
}

func upload(t *testing.T, m *Manager, sess types.Session, files ...uploadFile) {
	t.Helper()

	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)
	for _, file := range files {
		part, err := form.CreateFormFile("files", file.name)
		if err != nil {
			t.Fatalf("could not create form file: %v", err)
		}
		if _, err := part.Write([]byte(file.content)); err != nil {
			t.Fatalf("could not write form file: %v", err)
		}
	}
	if err := form.Close(); err != nil {
		t.Fatalf("could not close form: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", form.FormDataContentType())
	r = r.WithContext(auth.SetSession(r, sess))

	if err := m.uploadFileHandler(httptest.NewRecorder(), r); err != nil {
		t.Fatalf("upload failed: %v", err)
	}
}

func listedSize(t *testing.T, m *Manager, name string) int64 {
	t.Helper()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.fileList {
		if item.Name == name {
			return item.Size
		}
	}

	t.Fatalf("file %q is not listed", name)
	return 0
}

func assertFile(t *testing.T, m *Manager, name, want string) {
	t.Helper()

	got, err := os.ReadFile(filepath.Join(m.config.RootDir, name))
	if err != nil {
		t.Fatalf("could not read %q: %v", name, err)
	}
	if string(got) != want {
		t.Errorf("file %q contains %q, want %q", name, got, want)
	}
	if size := listedSize(t, m, name); size != int64(len(want)) {
		t.Errorf("file %q is listed with size %d, want %d", name, size, len(want))
	}
}

func TestUploadStoresFiles(t *testing.T) {
	m, sess := newTestManager(t)

	upload(t, m, sess, uploadFile{"a.txt", "first"}, uploadFile{"b.txt", "second file"})

	assertFile(t, m, "a.txt", "first")
	assertFile(t, m, "b.txt", "second file")
}

func TestUploadOverwritesExistingFile(t *testing.T) {
	tests := []struct {
		name  string
		first string
		next  string
	}{
		{"shorter", "a much longer original file", "short"},
		{"same size", "aaaaa", "bbbbb"},
		{"longer", "short", "a much longer replacement file"},
		{"empty", "not empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, sess := newTestManager(t)

			upload(t, m, sess, uploadFile{"file.txt", tt.first})
			upload(t, m, sess, uploadFile{"file.txt", tt.next})

			assertFile(t, m, "file.txt", tt.next)
		})
	}
}
