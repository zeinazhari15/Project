package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Ambil DSN dari env TEST_DSN, jika kosong pakai fallback (root tanpa password)
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = "root:@tcp(127.0.0.1:3306)/testdb"
	}

	fmt.Println("Using DSN:", dsn) // debug: tunjukkan DSN yang dipakai (bisa dihapus nanti)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error open DB: ", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Error ping DB: ", err)
	}

	fmt.Println("MySQL Connected!")
}
