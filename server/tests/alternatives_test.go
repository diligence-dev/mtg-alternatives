package server_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diligence-dev/mtg-alternatives/server"
)

func TestGetAlternatives_MissingName(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/alternatives", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body server.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error != "name parameter is required" {
		t.Fatalf("unexpected error: %q", body.Error)
	}
}

func TestGetAlternatives_ReturnsEmptyList(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/alternatives?name=Lightning+Bolt", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Alternatives []server.AlternativeResponse `json:"alternatives"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Alternatives) != 0 {
		t.Fatalf("expected 0 alternatives, got %d", len(body.Alternatives))
	}
}

func TestGetAlternatives_ReturnsStoredAlternatives(t *testing.T) {
	srv := newTestServer(t)

	server.InsertAlternative(srv.DB(), "Lightning Bolt", "image1.png")
	server.InsertAlternative(srv.DB(), "Lightning Bolt", "image2.png")

	req := httptest.NewRequest("GET", "/api/alternatives?name=Lightning+Bolt", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Alternatives []server.AlternativeResponse `json:"alternatives"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Alternatives) != 2 {
		t.Fatalf("expected 2 alternatives, got %d", len(body.Alternatives))
	}
	if !strings.HasPrefix(body.Alternatives[0].URL, "/uploads/") {
		t.Errorf("unexpected URL: %q", body.Alternatives[0].URL)
	}
}

func TestGetAlternatives_DeleteMethod(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("DELETE", "/api/alternatives?name=test", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestUpload_MissingName(t *testing.T) {
	srv := newTestServer(t)

	req := newUploadRequest(t, "", "test.png", "image/png", []byte("data"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body server.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error != "name is required" {
		t.Fatalf("unexpected error: %q", body.Error)
	}
}

func TestUpload_MissingFile(t *testing.T) {
	srv := newTestServer(t)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("name", "Lightning Bolt")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/alternatives", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body server.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error != "image file is required" {
		t.Fatalf("unexpected error: %q", body.Error)
	}
}

func TestUpload_WrongFileType(t *testing.T) {
	srv := newTestServer(t)

	req := newUploadRequest(t, "Card Name", "test.gif", "image/gif", []byte("GIF89a fake"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var body server.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if !strings.Contains(body.Error, "Invalid file type") {
		t.Fatalf("unexpected error: %q", body.Error)
	}
}

func TestUpload_Success(t *testing.T) {
	srv := newTestServer(t)

	req := newUploadRequest(t, "Lightning Bolt", "test.png", "image/png", []byte("fake png data"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Alternative server.AlternativeResponse `json:"alternative"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Alternative.ID == 0 {
		t.Fatal("expected non-zero id")
	}
	if !strings.HasPrefix(body.Alternative.URL, "/uploads/") {
		t.Errorf("unexpected URL: %q", body.Alternative.URL)
	}
	if body.Alternative.UploadedAt == "" {
		t.Error("expected uploaded_at to be set")
	}

	filename := strings.TrimPrefix(body.Alternative.URL, "/uploads/")
	filePath := filepath.Join(srv.UploadsDir(), filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("uploaded file was not saved to disk")
	}

	results, _ := server.GetAlternatives(srv.DB(), "Lightning Bolt")
	if len(results) != 1 {
		t.Fatalf("expected 1 DB record, got %d", len(results))
	}
	if results[0].Name != "Lightning Bolt" {
		t.Errorf("expected name 'Lightning Bolt', got %q", results[0].Name)
	}
}

func TestUpload_FilenameIsUUID(t *testing.T) {
	srv := newTestServer(t)

	req := newUploadRequest(t, "Lightning Bolt", "my-art.png", "image/png", []byte("data"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	var body struct {
		Alternative server.AlternativeResponse `json:"alternative"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)

	filename := strings.TrimPrefix(body.Alternative.URL, "/uploads/")
	if filename == "my-art.png" {
		t.Fatal("filename should be UUID-based, not original filename")
	}
	if !strings.HasSuffix(filename, ".png") {
		t.Errorf("expected .png extension, got %q", filename)
	}
}

func TestUpload_JpegAccepted(t *testing.T) {
	srv := newTestServer(t)

	req := newUploadRequest(t, "Lightning Bolt", "test.jpg", "image/jpeg", []byte("fake jpeg data"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpload_WebpAccepted(t *testing.T) {
	srv := newTestServer(t)

	req := newUploadRequest(t, "Lightning Bolt", "test.webp", "image/webp", []byte("fake webp data"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCardsWithAlternativesEndpoint(t *testing.T) {
	srv := newTestServer(t)

	server.InsertAlternative(srv.DB(), "Card A", "file1.png")
	server.InsertAlternative(srv.DB(), "Card B", "file2.png")

	req := httptest.NewRequest("GET", "/api/cards", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Cards []server.CardEntry `json:"cards"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(body.Cards))
	}

	names := map[string]bool{}
	for _, c := range body.Cards {
		names[c.Name] = true
	}
	if !names["Card A"] {
		t.Error("expected card 'Card A'")
	}
	if !names["Card B"] {
		t.Error("expected card 'Card B'")
	}
}

func newUploadRequest(t *testing.T, name, filename, contentType string, data []byte) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if name != "" {
		writer.WriteField("name", name)
	}

	if filename != "" {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="image"; filename="`+filename+`"`)
		h.Set("Content-Type", contentType)
		part, _ := writer.CreatePart(h)
		part.Write(data)
	}

	writer.Close()

	req := httptest.NewRequest("POST", "/api/alternatives", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
