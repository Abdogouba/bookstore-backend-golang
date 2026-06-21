package mappers

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
)

func ToBookResponse(
	book *models.Book,
) dto.BookResponse {

	return dto.BookResponse{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		Publisher: book.Publisher,
		Category:  book.Category,
		Price:     book.Price,
		Stock:     book.Stock,
		ImageURL: book.ImagePath,
		CreatedAt: book.CreatedAt,
	}
}

// ToBookListItemResponse converts a Book model
// into a response DTO.
func ToBookListItemResponse(
	book *models.Book,
) dto.BookListItemResponse {

	return dto.BookListItemResponse{
		ID: book.ID,

		Title: book.Title,

		Author: book.Author,

		Category: book.Category,

		Price: book.Price,

		IsOutOfStock: book.Stock == 0,

		ImagePath: book.ImagePath,
	}
}
