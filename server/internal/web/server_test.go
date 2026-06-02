package web

import (
	"errors"
	"net/http/httptest"
	"testing"
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
