package client

import (
	"bytes"
	"context"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

func TestFilesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/files" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"uuid":"f1","name":"photo.jpg","type":"image/jpeg","size":1024}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	files, err := c.FilesList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].UUID != "f1" {
		t.Errorf("expected uuid f1, got %q", files[0].UUID)
	}
	if files[0].Name != "photo.jpg" {
		t.Errorf("expected name photo.jpg, got %q", files[0].Name)
	}
}

func TestFileGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/file/f1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"uuid":"f1","name":"doc.pdf","type":"application/pdf","size":2048,"creator_id":1,"created_at":"2024-01-01","data_uri":"/file/f1/data","thumbnail_uri":"/file/f1/thumbnail"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	f, err := c.FileGet(context.Background(), "f1")
	if err != nil {
		t.Fatal(err)
	}
	if f.UUID != "f1" {
		t.Errorf("expected f1, got %q", f.UUID)
	}
	if f.Size != 2048 {
		t.Errorf("expected size 2048, got %d", f.Size)
	}
}

func TestFileUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer-api/v1/file" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		ct := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(ct)
		if err != nil {
			t.Fatalf("failed to parse Content-Type: %v", err)
		}
		if mediaType != "multipart/form-data" {
			t.Errorf("expected multipart/form-data, got %q", mediaType)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatalf("failed to read part: %v", err)
		}
		if part.FormName() != "data" {
			t.Errorf("expected form field 'data', got %q", part.FormName())
		}
		if part.FileName() != "test.txt" {
			t.Errorf("expected filename 'test.txt', got %q", part.FileName())
		}
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(part)
		if buf.String() != "hello world" {
			t.Errorf("expected body 'hello world', got %q", buf.String())
		}

		w.Header().Set("Location", "/customer-api/v1/file/uploaded-uuid")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	uuid, err := c.FileUpload(context.Background(), "test.txt", bytes.NewBufferString("hello world"))
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "uploaded-uuid" {
		t.Errorf("expected uploaded-uuid, got %q", uuid)
	}
}

func TestFileGetData(t *testing.T) {
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/file/f1/data" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify no Accept: application/json header
		if accept := r.Header.Get("Accept"); accept == "application/json" {
			t.Error("GetRaw should not set Accept: application/json")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(binaryData)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	data, err := c.FileGetData(context.Background(), "f1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, binaryData) {
		t.Errorf("expected %v, got %v", binaryData, data)
	}
}

func TestFileGetThumbnail(t *testing.T) {
	thumbData := []byte{0xFF, 0xD8, 0xFF}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer-api/v1/file/f1/thumbnail" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(thumbData)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	data, err := c.FileGetThumbnail(context.Background(), "f1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, thumbData) {
		t.Errorf("expected %v, got %v", thumbData, data)
	}
}

func TestFileGetError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetToken("tok")

	_, err := c.FileGet(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Errorf("expected 404 APIError, got %v", err)
	}
}
