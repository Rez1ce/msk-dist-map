package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "file:promos.db?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initSchema(db); err != nil {
		log.Fatal(err)
	}

	srv := &Server{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/promos/active", srv.handleGetActivePromos)
	mux.HandleFunc("/api/promos/create", srv.handleCreatePromo)
	mux.Handle("/api/promos/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			srv.handleUpdatePromo(w, r)
		case http.MethodDelete:
			srv.handleDeletePromo(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/", http.FileServer(http.Dir("static")))

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS promos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			station_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL,
			source_message_id TEXT,
			expires_at TIMESTAMP
		);
	`)
	return err
}
