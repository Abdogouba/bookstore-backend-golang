package dto

import "time"

// GetBooksQuery contains the optional query
// parameters for listing books.
type GetBooksQuery struct {
	Search string `form:"search"`

	Category string `form:"category"`

	Page int `form:"page"`

	PageSize int `form:"page_size"`
}

// BookListItemResponse represents a single
// book in the books list.
type BookListItemResponse struct {
	ID uint `json:"id"`

	Title string `json:"title"`

	Author string `json:"author"`

	Category string `json:"category"`

	Price float64 `json:"price"`

	IsOutOfStock bool `json:"is_out_of_stock"`

	ImagePath string `json:"image_path"`
}

// BooksListResponse represents the paginated
// response returned to the client.
type BooksListResponse struct {
	Books []BookListItemResponse `json:"books"`

	Page int `json:"page"`

	PageSize int `json:"page_size"`

	Total int64 `json:"total"`
}

type CreateBookRequest struct {
	Title string `form:"title" binding:"required,min=1,max=255"`

	Author string `form:"author" binding:"required,min=1,max=255"`

	Publisher string `form:"publisher" binding:"max=255"`

	Category string `form:"category" binding:"required"`

	Price float64 `form:"price" binding:"required"`

	Stock int `form:"stock" binding:"required"`
}

type BookResponse struct {
	ID uint `json:"id"`

	Title string `json:"title"`

	Author string `json:"author"`

	Publisher string `json:"publisher"`

	Category string `json:"category"`

	Price float64 `json:"price"`

	Stock int `json:"stock"`

	ImageURL string `json:"image_url"`

	CreatedAt time.Time `json:"created_at"`
}

// UpdateBookRequest contains book data
// for updating an existing book.
type UpdateBookRequest struct {
	Title string `form:"title" binding:"required,max=255"`

	Author string `form:"author" binding:"required,max=255"`

	Publisher string `form:"publisher" binding:"max=255"`

	Category string `form:"category" binding:"required"`

	Price float64 `form:"price" binding:"required"`

	Stock int `form:"stock" binding:"required"`
}