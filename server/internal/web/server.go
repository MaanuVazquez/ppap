package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"image/jpeg"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"ppap/server/internal/input"
	"ppap/server/internal/pen"
	"ppap/server/internal/screen"
)

//go:embed static
var embeddedClientFiles embed.FS

type ServerConfig struct {
	ClientDir    string
	Injector     pen.Injector
	Capturer     screen.Capturer
	LogPenEvents bool
}

type Server struct {
	clientDir string
	injector  pen.Injector
	capturer  screen.Capturer
	upgrader  websocket.Upgrader

	logPenEvents bool
	penEvents    atomic.Uint64
}

func NewServer(config ServerConfig) *Server {
	return &Server{
		clientDir:    config.ClientDir,
		injector:     config.Injector,
		capturer:     config.Capturer,
		logPenEvents: config.LogPenEvents,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(request *http.Request) bool {
				return true
			},
		},
	}
}

func (server *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", server.handleHealth)
	mux.HandleFunc("/api/pen", server.handlePenSocket)
	mux.HandleFunc("/api/screen.jpg", server.handleScreenJPEG)
	mux.Handle("/", server.staticHandler())

	return mux
}

func (server *Server) handleHealth(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]bool{
		"penInjection":  server.injector != nil,
		"screenCapture": server.capturer != nil,
	})
}

func (server *Server) handlePenSocket(writer http.ResponseWriter, request *http.Request) {
	connection, err := server.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer connection.Close()

	for {
		messageType, payload, err := connection.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("websocket read failed: %v", err)
			}
			return
		}
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}

		var event input.PenEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			server.writeSocketError(connection, "invalid-json", err)
			continue
		}
		if err := event.Validate(); err != nil {
			server.writeSocketError(connection, "invalid-event", err)
			continue
		}
		server.logPenEvent(event)
		if server.injector == nil {
			server.writeSocketError(connection, "injector-unavailable", errors.New("pen injector is unavailable"))
			continue
		}
		if err := server.injector.Inject(event); err != nil {
			server.writeSocketError(connection, "inject-failed", err)
			continue
		}
	}
}

func (server *Server) logPenEvent(event input.PenEvent) {
	if !server.logPenEvents {
		return
	}

	count := server.penEvents.Add(1)
	if count <= 20 || count%120 == 0 || event.Type != input.PenEventMove {
		log.Printf(
			"pen event #%d type=%s x=%.4f y=%.4f pressure=%.4f tilt=(%.1f,%.1f) twist=%.1f pointerId=%d",
			count,
			event.Type,
			event.X,
			event.Y,
			event.Pressure,
			event.TiltX,
			event.TiltY,
			event.Twist,
			event.PointerID,
		)
	}
}

func (server *Server) handleScreenJPEG(writer http.ResponseWriter, request *http.Request) {
	if server.capturer == nil {
		http.Error(writer, "screen capture is unavailable", http.StatusServiceUnavailable)
		return
	}

	image, err := server.capturer.Capture()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "image/jpeg")
	writer.Header().Set("Cache-Control", "no-store")
	if err := jpeg.Encode(writer, image, &jpeg.Options{Quality: 55}); err != nil {
		if isClientDisconnect(request, err) {
			return
		}

		log.Printf("jpeg encode failed: %v", err)
	}
}

func isClientDisconnect(request *http.Request, err error) bool {
	if request.Context().Err() != nil {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "broken pipe") ||
		strings.Contains(message, "connection reset by peer") ||
		strings.Contains(message, "connection was aborted") ||
		strings.Contains(message, "forcibly closed") ||
		strings.Contains(message, "wsasend")
}

func (server *Server) staticHandler() http.Handler {
	clientDir := server.clientDir
	if clientDir == "" {
		clientDir = firstExistingDir([]string{
			"../client/dist",
			"client/dist",
			"../../client/dist",
		})
	}
	if clientDir != "" {
		return spaFileServer(os.DirFS(clientDir))
	}

	clientFiles, err := fs.Sub(embeddedClientFiles, "static")
	if err != nil {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			http.Error(writer, "embedded client build is unavailable", http.StatusInternalServerError)
		})
	}

	return spaFileServer(clientFiles)
}

func spaFileServer(fileSystem fs.FS) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		filePath := strings.TrimPrefix(path.Clean("/"+request.URL.Path), "/")
		if filePath == "" {
			filePath = "index.html"
		}

		if info, err := fs.Stat(fileSystem, filePath); err == nil && !info.IsDir() {
			serveFileFromFS(writer, request, fileSystem, filePath, info)
			return
		}

		info, err := fs.Stat(fileSystem, "index.html")
		if err != nil || info.IsDir() {
			http.NotFound(writer, request)
			return
		}
		serveFileFromFS(writer, request, fileSystem, "index.html", info)
	})
}

func serveFileFromFS(writer http.ResponseWriter, request *http.Request, fileSystem fs.FS, filePath string, info fs.FileInfo) {
	content, err := fs.ReadFile(fileSystem, filePath)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	http.ServeContent(writer, request, path.Base(filePath), info.ModTime(), bytes.NewReader(content))
}

func (server *Server) writeSocketError(connection *websocket.Conn, code string, err error) {
	message := map[string]string{
		"type":    "error",
		"code":    code,
		"message": err.Error(),
	}
	_ = connection.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if writeErr := connection.WriteJSON(message); writeErr != nil {
		log.Printf("websocket error write failed: %v", writeErr)
	}
}

func firstExistingDir(paths []string) string {
	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path
		}
	}

	return ""
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		log.Printf("json write failed: %v", err)
	}
}
