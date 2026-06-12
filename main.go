package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/diligence-dev/mtg-alternatives/server"
)

//go:embed frontend
var frontendFS embed.FS

func main() {
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "data.db")
	uploadsDir := getEnv("UPLOADS_DIR", "uploads")

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Initialize database
	db, err := server.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	frontend, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		log.Fatalf("Failed to create frontend sub FS: %v", err)
	}

	// Create and start server
	srv := server.NewServer(db, uploadsDir, frontend)
	addr := fmt.Sprintf(":%s", port)
	if err := srv.ListenAndServe(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
