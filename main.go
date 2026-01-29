package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"
)

type Produk struct {
    ID    int     `json:"id"`
    Nama  string  `json:"nama"`
    Harga float64 `json:"harga"`
    Stok  int     `json:"stok"`
}

var produk = []Produk{
    {ID: 1, Nama: "Indomie Godog", Harga: 3500, Stok: 10},
    {ID: 2, Nama: "Vit 1000ml", Harga: 3000, Stok: 40},
	{ID: 3, Nama: "Kecap", Harga: 12000, Stok: 20},
}

func getProdukByID(w http.ResponseWriter, r *http.Request) {  // ← Hapus kurung tutup
    idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
        return
    }

    for _, p := range produk {
        if p.ID == id {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(p)
            return
        }
    }
    http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
}

func updateProduk(w http.ResponseWriter, r *http.Request) {
    idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
        return
    }

    var updatedProduk Produk
    err = json.NewDecoder(r.Body).Decode(&updatedProduk)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    for i := range produk {
        if produk[i].ID == id {
			updatedProduk.ID = id
            produk[i] = updatedProduk

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(updatedProduk)
            return
        }
    }  // ← Tambahkan kurung tutup loop
    
    http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
}  // tutup kurung

func deleteProduk(w http.ResponseWriter, r *http.Request) {
	// get id
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")
	// ganti id int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
		return
	}
	// loop produk cari ID, dapet index yang mau dihapus
	for i, p := range produk {
		if p.ID == id {
			// bikin slice baru dengan data sebelum dan sesudah index
			produk = append(produk[:i], produk[i+1:]...)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Produk berhasil dihapus",
			})
			
			return
		}
	}

	http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
}

func main() {

	// GET localhost:8080/api/produk/{id}
	// PUT localhost:8080/api/produk/{id}
	// DELETE localhost:8080/api/produk/{id}
    http.HandleFunc("/api/produk/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            getProdukByID(w, r)        
        } else if r.Method == "PUT" {
            updateProduk(w, r)
        } else if r.Method == "DELETE" {
            deleteProduk(w, r)
        }
    })
    
    http.HandleFunc("/api/produk", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(produk)
        } else if r.Method == "POST" {
            var produkBaru Produk
            err := json.NewDecoder(r.Body).Decode(&produkBaru)
            if err != nil {
                http.Error(w, "Invalid request", http.StatusBadRequest) 
                return          
            }

            produkBaru.ID = len(produk) + 1
            produk = append(produk, produkBaru)

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(produkBaru)
        }
    })
    
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "OK",
            "message": "Welcome to API",
        })
    })

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "OK",
            "message": "API is Running",
        })
    })
    
    fmt.Println("Server running di localhost:8080")
    
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("gagal running server")
    }
}
