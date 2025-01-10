package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to the database
	db, err := sql.Open("postgres", "user=postgres password=yourpassword dbname=test sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		// Extract 'page' and 'limit' query parameters
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page < 1 {
			page = 1 // Default to page 1
		}
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit < 1 {
			limit = 10 // Default to 10 items per page
		}

		// Calculate the OFFSET
		offset := (page - 1) * limit

		// Query the database
		rows, err := db.Query("SELECT id, name, created_at FROM items LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Process the rows
		items := []map[string]interface{}{}
		for rows.Next() {
			var id int
			var name string
			var createdAt string
			if err := rows.Scan(&id, &name, &createdAt); err != nil {
				http.Error(w, "Failed to scan items", http.StatusInternalServerError)
				return
			}
			items = append(items, map[string]interface{}{
				"id":         id,
				"name":       name,
				"created_at": createdAt,
			})
		}

		// Respond with JSON
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"page":%d,"limit":%d,"items":%v}`, page, limit, items)
	})

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
