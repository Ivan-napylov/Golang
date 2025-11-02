package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Genre  string `json:"genre"`
	Price  int    `json:"price"`
}

var db *sql.DB

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Cannot open log file:", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "booksdb"),
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/books", getBooksHandler)

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}

func getBooksHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	genre := r.URL.Query().Get("genre")
	sortParam := r.URL.Query().Get("sort")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	query := `SELECT id, title, author, genre, price FROM books`
	var args []interface{}
	var conditions []string

	if genre != "" {
		conditions = append(conditions, fmt.Sprintf("genre = $%d", len(args)+1))
		args = append(args, genre)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	switch sortParam {
	case "price_asc":
		query += " ORDER BY price ASC"
	case "price_desc":
		query += " ORDER BY price DESC"
	}
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			args = append(args, limit)
			query += fmt.Sprintf(" LIMIT $%d", len(args))
		}
	}
	if offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			args = append(args, offset)
			query += fmt.Sprintf(" OFFSET $%d", len(args))
		}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("DB query error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Genre, &b.Price); err != nil {
			log.Printf("Row scan error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	queryTime := time.Since(start)

	log.Printf("%s %s -> %d rows, took %v", r.Method, r.URL.RequestURI(), len(books), queryTime)

	w.Header().Set("X-Query-Time", queryTime.String())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}
