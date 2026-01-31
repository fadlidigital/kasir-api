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

	"github.com/spf13/viper"
	_ "github.com/lib/pq"
)

type Product struct {
	ID    int    `json:"id"`
	Nama  string `json:"name"`
	Harga int    `json:"price"`
	Stok  int    `json:"stock"`
}

type Config struct {
	Port   string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
}

var db *sql.DB

func initDB(connString string) error {
    fmt.Println("Trying to connect with:", connString) // DEBUG
    
    var err error
    db, err = sql.Open("postgres", connString)
    if err != nil {
        return fmt.Errorf("error opening database: %w", err)
    }

    fmt.Println("Database opened, now pinging...") // DEBUG
    
    if err := db.Ping(); err != nil {
        return fmt.Errorf("error connecting to database: %w", err)
    }

    fmt.Println("✅ Database connection established (Supabase PostgreSQL)")
    return nil
}


func getProdukByID(w http.ResponseWriter, r *http.Request) {
    idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
        return
    }

    var produk Product
    query := "SELECT id, name, price, stock FROM products WHERE id = $1"  // ✅ FIX: Tambahkan kutip penutup
    err = db.QueryRow(query, id).Scan(&produk.ID, &produk.Nama, &produk.Harga, &produk.Stok)
    if err == sql.ErrNoRows {
        http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        log.Println("Error querying database:", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(produk)
}


func getAllProduk(w http.ResponseWriter, r *http.Request) {
    query := "SELECT id, name, price, stock FROM products ORDER BY id"
    rows, err := db.Query(query)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        log.Println("Error querying database:", err)
        return
    }
    defer rows.Close()

    var produkList []Product  // ✅ FIX: Ganti "Products" → "Product"
    for rows.Next() {
        var p Product  // ✅ FIX: Ganti "Products" → "Product"
        if err := rows.Scan(&p.ID, &p.Nama, &p.Harga, &p.Stok); err != nil {
            http.Error(w, "Error scanning data", http.StatusInternalServerError)
            log.Println("Error scanning row:", err)
            return
        }
        produkList = append(produkList, p)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(produkList)
}


func createProduk(w http.ResponseWriter, r *http.Request) {
	var produk Product
	err := json.NewDecoder(r.Body).Decode(&produk)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id"
	err = db.QueryRow(query, produk.Nama, produk.Harga, produk.Stok).Scan(&produk.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("Error inserting data:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(produk)
}

func updateProduk(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
		return
	}

	var produk Product
	err = json.NewDecoder(r.Body).Decode(&produk)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	query := "UPDATE products SET name = $1, price = $2, stock = $3 WHERE id = $4"
	result, err := db.Exec(query, produk.Nama, produk.Harga, produk.Stok, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("Error updating data:", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
		return
	}

	produk.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produk)
}

func deleteProduk(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM products WHERE id = $1"
	result, err := db.Exec(query, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("Error deleting data:", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Produk berhasil dihapus",
	})
}

func main() {
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    if _, err := os.Stat(".env"); err == nil {
        viper.SetConfigFile(".env")
        _ = viper.ReadInConfig()
    }

    // ✅ Ambil dari environment variable (Railway akan inject PORT)
    config := Config{
        Port:   viper.GetString("PORT"),
        DBConn: viper.GetString("DB_CONN"),
    }

    fmt.Println("===== DEBUG =====")
    fmt.Println("PORT:", config.Port)
    fmt.Println("DB_CONN:", config.DBConn)
    fmt.Println("=================")

    // ✅ Default port 8080 jika tidak ada
    if config.Port == "" {
        config.Port = "8080"
    }

    if err := initDB(config.DBConn); err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()

    // ... (handler code tetap sama)

    fmt.Println("🚀 Server running di 0.0.0.0:" + config.Port)

	http.HandleFunc("/api/produk/", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        getProdukByID(w, r)
    } else if r.Method == "PUT" {
        updateProduk(w, r)
    } else if r.Method == "DELETE" {
        deleteProduk(w, r)
    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
})

http.HandleFunc("/api/produk", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        getAllProduk(w, r)
    } else if r.Method == "POST" {
        createProduk(w, r)
    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
})

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "OK",
        "message": "Welcome to Kasir API - Powered by Supabase",
    })
})

http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "OK",
        "message": "API is Running",
    })
})


    // ✅ PENTING: Listen di 0.0.0.0, bukan localhost!
    if err := http.ListenAndServe("0.0.0.0:"+config.Port, nil); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}


