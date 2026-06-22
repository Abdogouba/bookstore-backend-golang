package repositories

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"

	"gorm.io/gorm"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(
	db *gorm.DB,
) *BookRepository {

	return &BookRepository{
		db: db,
	}
}

// Create inserts a new book.
func (r *BookRepository) Create(
	book *models.Book,
) error {

	return r.db.
		Create(book).
		Error
}

// GetAll returns books after applying
// search, filter and pagination.
func (r *BookRepository) GetAll(
	query dto.GetBooksQuery,
) ([]models.Book, int64, error) {

	var books []models.Book

	var total int64

	// -------------------------
	// Start query
	// -------------------------

	db := r.db.Model(&models.Book{})

	// -------------------------
	// Search by title
	// -------------------------

	if query.Search != "" {

		db = db.Where(
			"title ILIKE ?",
			"%"+query.Search+"%",
		)
	}

	// -------------------------
	// Filter by category
	// -------------------------

	if query.Category != "" {

		db = db.Where(
			"category = ?",
			query.Category,
		)
	}

	// -------------------------
	// Count total BEFORE pagination
	// -------------------------

	err := db.Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// -------------------------
	// Calculate offset
	// -------------------------

	offset :=
		(query.Page - 1) * query.PageSize

	// -------------------------
	// Execute final query
	// -------------------------

	err = db.
		Order("price ASC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&books).
		Error

	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

// GetByID returns a book by its ID.
func (r *BookRepository) GetByID(
	id uint,
) (*models.Book, error) {

	var book models.Book

	err := r.db.
		First(&book, id).
		Error

	if err != nil {
		return nil, err
	}

	return &book, nil
}

// Save updates an existing book.
func (r *BookRepository) Save(
	book *models.Book,
) error {

	return r.db.Save(book).Error
}

func (r *BookRepository) Delete(
	book *models.Book,
) error {

	return r.db.Delete(book).Error
}

func (r *BookRepository) GetByIDs(
	bookIDs []uint,
) ([]models.Book, error) {

	var books []models.Book

	err := r.db.
		Where(
			"id IN ?",
			bookIDs,
		).
		Find(&books).
		Error

	return books, err
}

func (r *BookRepository) Update(
	tx *gorm.DB,
	book *models.Book,
) error {

	return tx.
		Save(book).
		Error
}