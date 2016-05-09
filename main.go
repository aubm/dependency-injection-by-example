package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type Post struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("mysql", "root:root@/my_items")
		defer db.Close()
		rows, err := db.Query("SELECT title, content FROM items")
		if err != nil {
			http.Error(w, "Internal server error", 500)
			return
		}
		defer rows.Close()
		posts := []Post{}
		for rows.Next() {
			p := Post{}
			err := rows.Scan(&p.Title, &p.Content)
			if err != nil {
				http.Error(w, "Internal server error", 500)
				return
			}
			posts = append(posts, p)
		}
		b, err := json.Marshal(posts)
		if err != nil {
			http.Error(w, "Internal server error", 500)
			return
		}
		w.Write(b)
		w.Header().Set("Content-Type", "application/json")
	})
	fmt.Println("Application started on port 8080")
	http.ListenAndServe(":8080", nil)
}
