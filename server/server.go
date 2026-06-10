package server

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	db         *sql.DB
	uploadsDir string
	frontend   fs.FS
	mux        *http.ServeMux
}

func NewServer(db *sql.DB, uploadsDir string, frontend fs.FS) *Server {
	s := &Server{
		db:         db,
		uploadsDir: uploadsDir,
		frontend:   frontend,
		mux:        http.NewServeMux(),
	}

	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/search", s.handleSearch)
	s.mux.HandleFunc("/api/alternatives", s.handleAlternatives)

	// Serve uploaded files
	uploadsFS := http.Dir(s.uploadsDir)
	s.mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(uploadsFS)))

	// Serve frontend
	if s.frontend != nil {
		s.mux.Handle("/", http.FileServer(http.FS(s.frontend)))
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) ListenAndServe(addr string) error {
	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, s)
}