package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestIsClientDisconnectRecognizesWindowsAbortedSend(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/screen.jpg", nil)
	err := errors.New("write tcp 192.168.1.3:4040->192.168.1.112:49587: wsasend: An established connection was aborted by the software in your host machine")

	if !isClientDisconnect(request, err) {
		t.Fatal("expected Windows aborted send to be treated as a client disconnect")
	}
}

func TestIsClientDisconnectRejectsGenericEncodeError(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/screen.jpg", nil)
	err := errors.New("jpeg encoder failed")

	if isClientDisconnect(request, err) {
		t.Fatal("expected generic encode error to remain reportable")
	}
}

func TestSPAFileServerServesRootWithoutRedirect(t *testing.T) {
	handler := spaFileServer(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><title>PPAP</title>")},
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if location := response.Header().Get("Location"); location != "" {
		t.Fatalf("expected no redirect location, got %q", location)
	}
}

func TestSPAFileServerFallsBackWithoutRedirect(t *testing.T) {
	handler := spaFileServer(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><title>PPAP</title>")},
	})

	request := httptest.NewRequest(http.MethodGet, "/missing-route", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if location := response.Header().Get("Location"); location != "" {
		t.Fatalf("expected no redirect location, got %q", location)
	}
}

func TestSPAFileServerServesStaticAsset(t *testing.T) {
	handler := spaFileServer(fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<!doctype html><title>PPAP</title>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
	})

	request := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if response.Body.String() != "console.log('ok')" {
		t.Fatalf("unexpected body %q", response.Body.String())
	}
}
