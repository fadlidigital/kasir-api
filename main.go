package main

import (
	"fmt"
	"kasir-api/database"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port   string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
}

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	// Ambil dari environment variable
	config := Config{
		Port:   viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}

	fmt.Println("===== DEBUG =====")
	fmt.Println("PORT:", config.Port)
	fmt.Println("DB_CONN:", config.DBConn)
	fmt.Println("=================")

	// Default port 8081 jika tidak ada
	if config.Port == "" {
		config.Port = "8081"
	}

	// Initialize database
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize repository, service, and handler
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// Routes
	http.HandleFunc("/api/produk/", productHandler.HandleProductByID)
	http.HandleFunc("/api/produk", productHandler.HandleProducts)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK","message":"Welcome to Kasir API - Powered by Supabase"}`))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK","message":"API is Running"}`))
	})

	// 0.0.0.0 artinya server bisa diakses dari semua network interface
	// localhost hanya bisa diakses dari komputer lokal
	// 0.0.0.0 diperlukan untuk deployment (Railway, Docker, dll)
	fmt.Printf("🚀 Server running on 0.0.0.0:%s (accessible via localhost:%s)\n", config.Port, config.Port)
	
	if err := http.ListenAndServe("0.0.0.0:"+config.Port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
