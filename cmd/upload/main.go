package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	var (
		addr    string
		baseDir string = "/data/uploads"
	)

	// Define the addr flag with a default value of ":3000"
	pflag.StringVar(&addr, "addr", ":3000", "HTTP network address (use :0 for a dynamic port)")
	pflag.StringVar(&baseDir, "dir", baseDir, "Base directory for storage")
	pflag.Parse()

	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatalf("Failed to ensure base dir exists: %v", err)
	}

	// Listen on the specified address (allowing :0 for dynamic port allocation)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	// Extract the actual address (including allocated port)
	actualAddr := listener.Addr().String()

	// Initialize the server
	srv := NewServer(baseDir)
	httpServer := &http.Server{
		Handler: srv.Router(),
	}

	// Log the allocated port
	log.Printf("Listening on %s...\n", actualAddr)

	// Start the server
	if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
