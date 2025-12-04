package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var tmpl *template.Template

func main() {
	// Parse semua template HTML di folder web/template
	tmpl = template.Must(template.ParseGlob("web/template/*.html"))

	// gunakan ServeMux agar bisa mount static dan route terpisah
	mux := http.NewServeMux()

	// --- Static files (important) ---
	// Pastikan folder web/static berisi cv.pdf
	fileServer := http.FileServer(http.Dir("web/static"))
	// URL /static/cv.pdf -> file web/static/cv.pdf
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// --- Routes ---
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// render home.html
		if err := tmpl.ExecuteTemplate(w, "home.html", nil); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("home template error:", err)
		}
	})

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.ExecuteTemplate(w, "dashboard.html", nil); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("dashboard template error:", err)
		}
	})

	mux.HandleFunc("/news", func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.ExecuteTemplate(w, "news.html", nil); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("news template error:", err)
		}
	})

	// optional: logging middleware
	handler := loggingMiddleware(mux)

	// baca PORT dari env (Render memberi PORT), fallback 8080
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

	log.Printf("Server running on http://0.0.0.0%s\n", addr)
    log.Printf("Server running on http://localhost:8080/", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
