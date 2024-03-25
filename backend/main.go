package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Entry struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}

// main function
func main() {
	// connect to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS entries (id SERIAL PRIMARY KEY, name TEXT, link TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	// create router
	router := mux.NewRouter()
	router.HandleFunc("/api/go/entries", getEntries(db)).Methods("GET")
	router.HandleFunc("/api/go/entries", createEntry(db)).Methods("POST")
	router.HandleFunc("/api/go/entries/{id}", getEntry(db)).Methods("GET")
	router.HandleFunc("/api/go/entries/{id}", updateEntry(db)).Methods("PUT")
	router.HandleFunc("/api/go/entries/{id}", deleteEntry(db)).Methods("DELETE")

	// wrap the router with CORS and JSON content type middlewares
	enhancedRouter := enableCORS(jsonContentTypeMiddleware(router))
	// start server
	log.Fatal(http.ListenAndServe(":8000", enhancedRouter))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Check if the request is for CORS preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass down the request to the next middleware (or final handler)
		next.ServeHTTP(w, r)
	})
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set JSON Content-Type
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// get all entries
func getEntries(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM entries")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		entries := []Entry{} // array of entries
		for rows.Next() {
			var e Entry
			if err := rows.Scan(&e.Id, &e.Name, &e.Link); err != nil {
				log.Fatal(err)
			}
			entries = append(entries, e)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(entries)

	}

}

// get entry by id
func getEntry(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var e Entry
		err := db.QueryRow("SELECT * FROM entries WHERE id = $1", id).Scan(&e.Id, &e.Name, &e.Link)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(e)
	}
}

func createEntry(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e Entry
		json.NewDecoder(r.Body).Decode(&e)

		err := db.QueryRow("INSERT INTO entries (name, link) VALUES ($1, $2) RETURNING id", e.Name, e.Link).Scan(&e.Id)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(e)
	}
}

func updateEntry(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e Entry
		json.NewDecoder(r.Body).Decode(&e)

		vars := mux.Vars(r)
		id := vars["id"]

		// Execute the update query
		_, err := db.Exec("UPDATE entries SET name = $1, link = $2 WHERE id = $3", e.Name, e.Link, id)
		if err != nil {
			log.Fatal(err)
		}

		// Retrieve the updated entry data from the database
		var updatedEntry Entry
		err = db.QueryRow("SELECT id, name, link FROM entries WHERE id = $1", id).Scan(&updatedEntry.Id, &updatedEntry.Name, &updatedEntry.Link)
		if err != nil {
			log.Fatal(err)
		}

		// Send the updated entry data in the response
		json.NewEncoder(w).Encode(updatedEntry)
	}
}

func deleteEntry(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var e Entry
		err := db.QueryRow("SELECT * FROM entries WHERE id = $1", id).Scan(&e.Id, &e.Name, &e.Link)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM entries WHERE id = $1", id)
			if err != nil {
				//todo : fix error handling
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("Entry deleted")
		}
	}
}
