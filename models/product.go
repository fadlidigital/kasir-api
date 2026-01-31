package models

type Product struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Price        int    `json:"price"`
	Stock        int    `json:"stock"`
	CategoryID   *int   `json:"category_id,omitempty"`   // Pointer karena bisa NULL
	CategoryName string `json:"category_name,omitempty"` // Untuk JOIN result
}
