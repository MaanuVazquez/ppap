package main

import (
	"flag"
	"log"
	"net/http"

	"ppap/server/internal/pen"
	"ppap/server/internal/platform"
	"ppap/server/internal/screen"
	"ppap/server/internal/web"
)

func main() {
	addr := flag.String("addr", ":4040", "HTTP listen address")
	clientDir := flag.String("client-dir", "", "built client directory to serve")
	logPenEvents := flag.Bool("log-pen-events", false, "log sampled pen events received from clients")
	flag.Parse()

	if err := platform.EnableDPIAwareness(); err != nil {
		log.Printf("DPI awareness setup failed: %v", err)
	}

	injector, err := pen.NewInjector()
	if err != nil {
		log.Printf("pen injection unavailable: %v", err)
	}
	if injector != nil {
		defer injector.Close()
	}

	capturer, err := screen.NewCapturer()
	if err != nil {
		log.Printf("screen capture unavailable: %v", err)
	}
	if capturer != nil {
		defer capturer.Close()
	}

	server := web.NewServer(web.ServerConfig{
		ClientDir:    *clientDir,
		Injector:     injector,
		Capturer:     capturer,
		LogPenEvents: *logPenEvents,
	})

	log.Printf("ppap server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, server.Routes()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
