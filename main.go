package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to the database
	db, err := sql.Open("postgres", "user=postgres password=Siddhesh dbname=test sslmode=disable")
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

		sort := r.URL.Query().Get("sort")
		order := r.URL.Query().Get("order")
		nameFilter := r.URL.Query().Get("name")
		createdAfter := r.URL.Query().Get("created_after")

		// Validate the sort column
		validSortColumns := map[string]bool{"name": true, "created_at": true}
		if !validSortColumns[sort] {
			sort = "created_at" // Default sort column
		}

		// Validate the sort order
		if order != "asc" && order != "desc" {
			order = "asc" // Default sort order
		}
		fmt.Println("sort:", sort, "order:", order)

		// Calculate the OFFSET
		offset := (page - 1) * limit

		whereClauses := []string{}
		args := []interface{}{} // Slice to store the query arguments
		argIndex := 1

		if nameFilter != "" {
			whereClauses = append(whereClauses, "name ILIKE $1")
			args = append(args, "%"+nameFilter+"%")
			argIndex++
		}

		if createdAfter != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("created_at > $%d", argIndex))
			args = append(args, createdAfter)
			argIndex++
		}

		args = append(args, limit, offset)

		// Combine WHERE clauses
		whereSQL := ""
		if len(whereClauses) > 0 {
			whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
		}

		// Query the database
		query := fmt.Sprintf("SELECT id, name, created_at FROM items %s ORDER BY %s %s LIMIT $%d OFFSET $%d", whereSQL, sort, order, argIndex, argIndex+1)
		rows, err := db.Query(query, args...)
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
