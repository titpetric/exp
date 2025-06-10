package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Metadata struct {
	Filename    string `json:"filename"` // Original, unhashed path
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	SHA256      string `json:"sha256"`
}

type Server struct {
	BaseDir string
}

// NewServer creates a new Server instance.
func NewServer(baseDir string) *Server {
	return &Server{BaseDir: baseDir}
}

// LoggingMiddleware wraps the next handler in CombinedLoggingHandler.
func LoggingMiddleware(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, next)
}

// Router returns the HTTP handler.
func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	// Use logging middleware first, so all requests are logged.
	r.Use(LoggingMiddleware)

	// Define handlers
	r.HandleFunc("/", s.ListHandler).Methods(http.MethodGet)
	r.HandleFunc("/{path:.*}", s.PutHandler).Methods(http.MethodPut)
	r.HandleFunc("/{path:.*}", s.GetHandler).Methods(http.MethodGet)
	r.HandleFunc("/{path:.*}", s.DeleteHandler).Methods(http.MethodDelete)

	return r
}

// getLocalFile hashes reqPath and returns the hash as a hex string.
func (s *Server) getLocalFile(reqPath string) string {
	h := sha256.Sum256([]byte(reqPath))
	return hex.EncodeToString(h[:])
}

// getFilePaths returns the full paths for the data and json files given a hash.
func (s *Server) getFilePaths(hash string) (dataFile, jsonFile string) {
	dataFile = filepath.Join(s.BaseDir, hash+".data")
	jsonFile = filepath.Join(s.BaseDir, hash+".json")
	return dataFile, jsonFile
}

// writeFile handles writing the file and metadata atomically via temporary files and rename.
// Returns the hash and the metadata on success.
func (s *Server) writeFile(reqPath string, r *http.Request) (string, Metadata, error) {
	hash := s.getLocalFile(reqPath)
	dataFile, jsonFile := s.getFilePaths(hash)

	tempData := dataFile + ".tmp"
	tempJSON := jsonFile + ".tmp"

	// Write data to a temporary file
	f, err := os.Create(tempData)
	if err != nil {
		return hash, Metadata{}, fmt.Errorf("failed to create temp data file: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(f, hasher), r.Body)
	if err != nil {
		return hash, Metadata{}, fmt.Errorf("failed to write data: %w", err)
	}

	// Flush and sync data file
	if err := f.Sync(); err != nil {
		return hash, Metadata{}, fmt.Errorf("failed to sync temp data file: %w", err)
	}

	dataHash := hex.EncodeToString(hasher.Sum(nil))
	meta := Metadata{
		Filename:    reqPath,
		ContentType: r.Header.Get("Content-Type"),
		Size:        size,
		SHA256:      dataHash,
	}

	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return hash, Metadata{}, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write metadata to a temporary file
	if err := os.WriteFile(tempJSON, metaBytes, 0644); err != nil {
		return hash, Metadata{}, fmt.Errorf("failed to write temp metadata file: %w", err)
	}

	// Sync the metadata file
	metadataF, err := os.OpenFile(tempJSON, os.O_RDWR, 0644)
	if err != nil {
		return hash, Metadata{}, fmt.Errorf("failed to reopen temp metadata file: %w", err)
	}
	if err := metadataF.Sync(); err != nil {
		metadataF.Close()
		return hash, Metadata{}, fmt.Errorf("failed to sync temp metadata file: %w", err)
	}
	metadataF.Close()

	// Rename temp files to final files atomically
	if err := os.Rename(tempData, dataFile); err != nil {
		return hash, Metadata{}, fmt.Errorf("failed to finalize data file: %w", err)
	}
	if err := os.Rename(tempJSON, jsonFile); err != nil {
		return hash, Metadata{}, fmt.Errorf("failed to finalize metadata file: %w", err)
	}

	return hash, meta, nil
}

// ListHandler handles GET /
// It reads all *.json metadata files in the BaseDir and returns them as a JSON array.
func (s *Server) ListHandler(w http.ResponseWriter, r *http.Request) {
	files, err := filepath.Glob(filepath.Join(s.BaseDir, "*.json"))
	if err != nil {
		http.Error(w, "Failed to list metadata files", http.StatusInternalServerError)
		return
	}

	var allMeta []Metadata
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			// If unreadable, skip
			continue
		}
		var meta Metadata
		if err := json.Unmarshal(data, &meta); err != nil {
			// If invalid JSON, skip
			continue
		}
		allMeta = append(allMeta, meta)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(allMeta); err != nil {
		http.Error(w, "Failed to encode metadata", http.StatusInternalServerError)
	}
}

// PutHandler handles PUT /{path:.*}
func (s *Server) PutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqPath := strings.TrimPrefix(vars["path"], "/")

	_, meta, err := s.writeFile(reqPath, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(meta)
}

// GetHandler handles GET /{path:.*}
func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqPath := strings.TrimPrefix(vars["path"], "/")

	hash := s.getLocalFile(reqPath)
	dataFile, jsonFile := s.getFilePaths(hash)

	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		http.Error(w, "Metadata Not Found", http.StatusNotFound)
		return
	}

	metaBytes, err := os.ReadFile(jsonFile)
	if err != nil {
		http.Error(w, "Failed to read metadata", http.StatusInternalServerError)
		return
	}
	var meta Metadata
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		http.Error(w, "Failed to parse metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", meta.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))
	http.ServeFile(w, r, dataFile)
}

// DeleteHandler handles DELETE /{path:.*}
func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqPath := strings.TrimPrefix(vars["path"], "/")

	hash := s.getLocalFile(reqPath)
	dataFile, jsonFile := s.getFilePaths(hash)

	if err := os.Remove(dataFile); err != nil && !os.IsNotExist(err) {
		http.Error(w, "Failed to delete data file", http.StatusInternalServerError)
		return
	}

	if err := os.Remove(jsonFile); err != nil && !os.IsNotExist(err) {
		http.Error(w, "Failed to delete metadata file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
