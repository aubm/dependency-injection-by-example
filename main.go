package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/facebookgo/inject"
	_ "github.com/go-sql-driver/mysql"
)

type Post struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type PostsHandlers struct {
	Manager interface {
		FindPosts() ([]Post, error)
	} `inject:""`
	Encoder interface {
		ToJSON(w http.ResponseWriter, src interface{})
	} `inject:""`
}

func (ph *PostsHandlers) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := ph.Manager.FindPosts()
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	ph.Encoder.ToJSON(w, posts)
}

type PostsManager struct {
	DB *sql.DB `inject:""`
}

func (pm *PostsManager) FindPosts() ([]Post, error) {
	rows, err := pm.DB.Query("SELECT title, content FROM items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := []Post{}
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Title, &p.Content)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

type DefaultEncoder struct{}

func (de *DefaultEncoder) ToJSON(w http.ResponseWriter, src interface{}) {
	b, err := json.Marshal(src)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}
	w.Write(b)
	w.Header().Set("Content-Type", "application/json")
}

func main() {
	db, err := sql.Open("mysql", "root:root@/my_items")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var postsManager PostsManager
	var encoder DefaultEncoder
	var postsHandlers PostsHandlers

	if err := inject.Populate(db, &postsHandlers, &encoder, &postsManager); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	http.HandleFunc("/posts", postsHandlers.GetPosts)

	fmt.Println("Application started on port 8080")
	http.ListenAndServe(":8080", nil)
}
