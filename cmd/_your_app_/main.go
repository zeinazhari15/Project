package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var tmpl *template.Template
var db *sql.DB

// Submission merepresentasikan row di table news_submissions
type Submission struct {
	ID        int
	Username  string
	Email     string
	CreatedAt time.Time
}

func main() {
	// parse semua template di web/template
	tmpl = template.Must(template.ParseGlob("web/template/*.html"))

	// baca env DB
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser != "" && dbHost != "" && dbPort != "" && dbName != "" {
		dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"
		var err error
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("failed to open db: %v", err)
		}
		// ping untuk memastikan koneksi ok
		if err := db.Ping(); err != nil {
			log.Fatalf("failed to ping db: %v", err)
		}
		// tutup DB saat proses exit
		defer db.Close()
		log.Println("Connected to DB")
	} else {
		log.Println("DB environment variables not fully set, DB features will be disabled")
	}

	mux := http.NewServeMux()

	// static files
	fileServer := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// routes
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/dashboard", dashboardHandler)
	mux.HandleFunc("/news", newsHandler)

	// api endpoints
	mux.HandleFunc("/submit-news", submitNewsHandler) // POST JSON
	mux.HandleFunc("/history", historyHandler)        // GET -> render history.html

	// start server
	handler := loggingMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server running on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// handlers

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("home template error:", err)
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "dashboard.html", nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("dashboard template error:", err)
	}
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "news.html", nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("news template error:", err)
	}
}

// submitNewsHandler expects JSON {username, email}, both required
func submitNewsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// server-side validation: both must be provided
	if payload.Username == "" || payload.Email == "" {
		http.Error(w, "username and email must be filled", http.StatusBadRequest)
		return
	}

	if db == nil {
		http.Error(w, "database not configured", http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("INSERT INTO news_submissions (username, email) VALUES (?, ?)")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		log.Println("prepare error:", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(payload.Username, payload.Email)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		log.Println("insert error:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// historyHandler: query all news_submissions and render history.html
func historyHandler(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "database not configured", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, username, email, created_at FROM news_submissions ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		log.Println("query error:", err)
		return
	}
	defer rows.Close()

	var list []Submission
	for rows.Next() {
		var s Submission
		if err := rows.Scan(&s.ID, &s.Username, &s.Email, &s.CreatedAt); err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			log.Println("scan error:", err)
			return
		}
		list = append(list, s)
	}
	if err := rows.Err(); err != nil {
		log.Println("rows error:", err)
	}

	// data untuk template
	data := map[string]interface{}{
		"Submissions": list,
	}

	if err := tmpl.ExecuteTemplate(w, "history.html", data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		log.Println("history template error:", err)
		return
	}
}

// middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
