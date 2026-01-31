package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"kasir-api/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetAll() ([]models.Product, error) {
	query := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.stock,
			p.category_id,
			c.name AS category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.id
	`
	
	fmt.Println("DEBUG: Executing GetAll query...")
	rows, err := r.db.Query(query)
	if err != nil {
		fmt.Printf("DEBUG: Query error: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		var categoryID sql.NullInt64      // PENTING: Handle NULL
		var categoryName sql.NullString   // PENTING: Handle NULL
		
		// Scan 6 kolom (bukan 4!)
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Price,
			&p.Stock,
			&categoryID,      // Scan ke sql.NullInt64
			&categoryName,    // Scan ke sql.NullString
		); err != nil {
			fmt.Printf("DEBUG: Scan error: %v\n", err)
			return nil, err
		}
		
		// Convert sql.NullInt64 ke *int
		if categoryID.Valid {
			catID := int(categoryID.Int64)
			p.CategoryID = &catID
			fmt.Printf("DEBUG: Product %d has category_id: %d\n", p.ID, catID)
		} else {
			fmt.Printf("DEBUG: Product %d has NULL category_id\n", p.ID)
		}
		
		// Convert sql.NullString ke string
		if categoryName.Valid {
			p.CategoryName = categoryName.String
			fmt.Printf("DEBUG: Product %d has category_name: %s\n", p.ID, p.CategoryName)
		} else {
			fmt.Printf("DEBUG: Product %d has NULL category_name\n", p.ID)
		}
		
		products = append(products, p)
	}

	fmt.Printf("DEBUG: Total products fetched: %d\n", len(products))
	return products, nil
}


// GetByID - ambil produk by ID
func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	query := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.stock,
			p.category_id,
			c.name AS category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`
	
	fmt.Printf("DEBUG: Executing GetByID query for id: %d\n", id)
	var p models.Product
	var categoryID sql.NullInt64
	var categoryName sql.NullString
	
	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Price,
		&p.Stock,
		&categoryID,
		&categoryName,
	)
	
	if err != nil {
		fmt.Printf("DEBUG: GetByID error: %v\n", err)
		return nil, err
	}
	
	if categoryID.Valid {
		catID := int(categoryID.Int64)
		p.CategoryID = &catID
		fmt.Printf("DEBUG: Product %d has category_id: %d\n", p.ID, catID)
	} else {
		fmt.Printf("DEBUG: Product %d has NULL category_id\n", p.ID)
	}
	
	if categoryName.Valid {
		p.CategoryName = categoryName.String
		fmt.Printf("DEBUG: Product %d has category_name: %s\n", p.ID, p.CategoryName)
	} else {
		fmt.Printf("DEBUG: Product %d has NULL category_name\n", p.ID)
	}
	
	return &p, nil
}


// Create - tambah produk baru
func (r *ProductRepository) Create(product *models.Product) error {
	query := `
		INSERT INTO products (name, price, stock, category_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query, 
		product.Name, 
		product.Price, 
		product.Stock, 
		product.CategoryID,
	).Scan(&product.ID)
	
	if err != nil {
		fmt.Printf("DEBUG: Create error: %v\n", err)
		return err
	}
	
	fmt.Printf("DEBUG: Product created with ID: %d\n", product.ID)
	return nil
}

// Update - update produk
func (r *ProductRepository) Update(product *models.Product) error {
	query := `
		UPDATE products 
		SET name = $1, price = $2, stock = $3, category_id = $4 
		WHERE id = $5
	`
	
	result, err := r.db.Exec(
		query, 
		product.Name, 
		product.Price, 
		product.Stock, 
		product.CategoryID, 
		product.ID,
	)
	
	if err != nil {
		fmt.Printf("DEBUG: Update error: %v\n", err)
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return errors.New("produk tidak ditemukan")
	}
	
	fmt.Printf("DEBUG: Product %d updated successfully\n", product.ID)
	return nil
}

// Delete - hapus produk
func (r *ProductRepository) Delete(id int) error {
	query := "DELETE FROM products WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("produk tidak ditemukan")
	}

	return err
}
