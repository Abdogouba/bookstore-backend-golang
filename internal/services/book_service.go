package services

import (
	"errors"
	"os"

	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/mappers"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/repositories"
	"bookstore-backend/internal/utils"

	"gorm.io/gorm"
)

type BookService struct {
	bookRepo *repositories.BookRepository
}

func NewBookService(
	db *gorm.DB,
) *BookService {

	return &BookService{
		bookRepo: repositories.NewBookRepository(db),
	}
}

// CreateBook creates a new book.
func (s *BookService) CreateBook(
	request dto.CreateBookRequest,
	imagePath string,
) (*dto.BookResponse, error) {

	// -------------------------
	// Validate category
	// -------------------------

	if !utils.IsValidCategory(
		request.Category,
	) {

		return nil,
			errors.New("invalid category")
	}

	// -------------------------
	// Validate price
	// -------------------------

	if request.Price <= 0 {

		return nil,
			errors.New("price must be greater than 0")
	}

	// -------------------------
	// Validate stock
	// -------------------------

	if request.Stock < 0 {

		return nil,
			errors.New("stock cannot be negative")
	}

	// -------------------------
	// Create model
	// -------------------------

	book := models.Book{

		Title: request.Title,

		Author: request.Author,

		Publisher: request.Publisher,

		Category: request.Category,

		Price: request.Price,

		Stock: request.Stock,

		ImagePath: imagePath,
	}

	// -------------------------
	// Save book
	// -------------------------

	err := s.bookRepo.Create(
		&book,
	)

	if err != nil {
		_ = os.Remove(imagePath)
		return nil, err
	}

	// -------------------------
	// Build response
	// -------------------------

	response :=
		mappers.ToBookResponse(
			&book,
		)

	return &response, nil
}

// GetBooks returns a paginated list of books.
func (s *BookService) GetBooks(
	query dto.GetBooksQuery,
) (*dto.BooksListResponse, error) {

	// -------------------------
	// Default pagination
	// -------------------------

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// -------------------------
	// Validate category
	// -------------------------

	if query.Category != "" &&
		!utils.IsValidCategory(query.Category) {

		return nil,
			errors.New("invalid category")
	}

	// -------------------------
	// Query database
	// -------------------------

	books,
		total,
		err :=
		s.bookRepo.GetAll(query)

	if err != nil {
		return nil, err
	}

	// -------------------------
	// Map books
	// -------------------------

	responseBooks :=
		make(
			[]dto.BookListItemResponse,
			0,
			len(books),
		)

	for i := range books {

		responseBooks = append(
			responseBooks,
			mappers.ToBookListItemResponse(
				&books[i],
			),
		)
	}

	// -------------------------
	// Build response
	// -------------------------

	response :=
		dto.BooksListResponse{
			Books: responseBooks,
			Page: query.Page,
			PageSize: query.PageSize,
			Total: total,
		}

	return &response, nil
}

// GetBookByID returns one book.
func (s *BookService) GetBookByID(
	bookID uint,
) (*dto.BookResponse, error) {

	// -------------------------
	// Get book
	// -------------------------

	book, err :=
		s.bookRepo.GetByID(bookID)

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {

			return nil,
				errors.New("book not found")
		}

		return nil, err
	}

	// -------------------------
	// Map to response
	// -------------------------

	response :=
		mappers.ToBookResponse(book)

	return &response, nil
}

// UpdateBook updates an existing book.
func (s *BookService) UpdateBook(
	bookID uint,
	request dto.UpdateBookRequest,
	imagePath string,
) (*dto.BookResponse, string, error) {

	// -------------------------
	// Get existing book
	// -------------------------

	book, err :=
		s.bookRepo.GetByID(bookID)

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {

			return nil,
				"",
				errors.New("book not found")
		}

		return nil,
			"",
			err
	}

	// -------------------------
	// Category validation
	// -------------------------

	if !utils.IsValidCategory(
		request.Category,
	) {

		return nil,
			"",
			errors.New("invalid category")
	}

	// -------------------------
	// Price validation
	// -------------------------

	if request.Price <= 0 {

		return nil,
			"",
			errors.New(
				"price must be greater than 0",
			)
	}

	// -------------------------
	// Stock validation
	// -------------------------

	if request.Stock < 0 {

		return nil,
			"",
			errors.New(
				"stock cannot be negative",
			)
	}

	// -------------------------
	// Preserve old image path
	// -------------------------

	oldImagePath :=
		book.ImagePath

	// -------------------------
	// Update fields
	// -------------------------

	book.Title =
		request.Title

	book.Author =
		request.Author

	book.Publisher =
		request.Publisher

	book.Category =
		request.Category

	book.Price =
		request.Price

	book.Stock =
		request.Stock

	// -------------------------
	// New image uploaded?
	// -------------------------

	if imagePath != "" {

		book.ImagePath =
			imagePath
	}

	// -------------------------
	// Save changes
	// -------------------------

	err =
		s.bookRepo.Save(book)

	if err != nil {

		return nil,
			"",
			err
	}

	// -------------------------
	// Build response
	// -------------------------

	response :=
		mappers.ToBookResponse(book)

	// -------------------------
	// Return old image path
	// -------------------------

	// Handler will decide
	// whether it should be deleted.

	return &response,
		oldImagePath,
		nil
}

func (s *BookService) DeleteBook(
	bookID uint,
) error {

	// -------------------------
	// Get book
	// -------------------------

	book, err :=
		s.bookRepo.GetByID(bookID)

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {

			return errors.New("book not found")
		}

		return err
	}

	// -------------------------
	// Soft delete
	// -------------------------

	err =
		s.bookRepo.Delete(book)

	if err != nil {
		return err
	}

	return nil
}