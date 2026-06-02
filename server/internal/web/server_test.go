package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"ppap/server/internal/input"
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

func TestInjectPressureTestStrokeRampsPressure(t *testing.T) {
	injector := &recordingInjector{}
	server := NewServer(ServerConfig{Injector: injector})

	if err := server.injectPressureTestStroke(); err != nil {
		t.Fatalf("expected pressure test stroke to inject: %v", err)
	}

	if len(injector.events) != 62 {
		t.Fatalf("expected 62 pen events, got %d", len(injector.events))
	}
	if injector.events[0].Type != input.PenEventDown {
		t.Fatalf("expected first event down, got %s", injector.events[0].Type)
	}
	lastEvent := injector.events[len(injector.events)-1]
	if lastEvent.Type != input.PenEventUp {
		t.Fatalf("expected last event up, got %s", lastEvent.Type)
	}
	if injector.events[1].Pressure >= injector.events[len(injector.events)-2].Pressure {
		t.Fatalf("expected pressure ramp, got start %.2f end %.2f", injector.events[1].Pressure, injector.events[len(injector.events)-2].Pressure)
	}
}

type recordingInjector struct {
	events []input.PenEvent
}

func (injector *recordingInjector) Inject(event input.PenEvent) error {
	injector.events = append(injector.events, event)
	return nil
}

func (injector *recordingInjector) Close() error {
	return nil
}
