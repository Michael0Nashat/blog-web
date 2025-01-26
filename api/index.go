package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq" // PostgreSQL driver for NeonDB
)

type Post struct {
	ID      int
	Title   string
	Content string
}

var (
	db       *sql.DB
	tmpl     = template.Must(template.ParseGlob("templates/*.html"))
	dbConfig = "postgres://neondb_owner:npg_2HIor0XPLiMv@ep-muddy-mouse-a5w8aefb-pooler.us-east-2.aws.neon.tech/neondb?sslmode=require"
)

func main() {
	var err error

	// Initialize the database connection
	db, err = sql.Open("postgres", dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure the database is reachable
	if err = db.Ping(); err != nil {
		log.Fatalf("Cannot ping the database: %v", err)
	}

	// Set up routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/post/new", newPostHandler)
	http.HandleFunc("/post/create", createPostHandler)
	http.HandleFunc("/post/view", viewPostHandler)

	// Start the server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, content FROM posts")
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	tmpl.ExecuteTemplate(w, "home.html", posts)
}

func newPostHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "new.html", nil)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	_, err := db.Exec("INSERT INTO posts (title, content) VALUES ($1, $2)", title, content)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func viewPostHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var post Post
	if err := db.QueryRow("SELECT id, title, content FROM posts WHERE id = $1", id).Scan(&post.ID, &post.Title, &post.Content); err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	tmpl.ExecuteTemplate(w, "view.html", post)
}
