package handlers

import (
	"net/http"
	"os"
	"strconv"

	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/services"
	"bookstore-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookHandler struct {
	bookService *services.BookService
}

func NewBookHandler(
	db *gorm.DB,
) *BookHandler {

	return &BookHandler{
		bookService: services.NewBookService(db),
	}
}

// CreateBook godoc
//
// @Summary Admin Add Book
// @Description Create a new book
// @Tags Books
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
//
// @Param title formData string true "Title"
// @Param author formData string true "Author"
// @Param publisher formData string false "Publisher"
// @Param category formData string true "Category"
// @Param price formData number true "Price"
// @Param stock formData integer true "Stock"
// @Param image formData file false "Book Image"
//
// @Success 201 {object} dto.BookResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
//
// @Router /books [post]
func (h *BookHandler) CreateBook(
	c *gin.Context,
) {

	// -------------------------
	// Bind form fields
	// -------------------------

	var request dto.CreateBookRequest

	if err :=
		c.ShouldBind(&request); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	// -------------------------
	// Save image
	// -------------------------

	imagePath, err :=
		utils.SaveUploadedImage(
			c,
			"image",
		)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	// -------------------------
	// Create book
	// -------------------------

	response, err :=
		h.bookService.CreateBook(
			request,
			imagePath,
		)

	if err != nil {

		if imagePath != "" {
			_ = os.Remove(imagePath)
		}

		switch err.Error() {

		case "invalid category",
			"price must be greater than 0",
			"stock cannot be negative":

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	c.JSON(
		http.StatusCreated,
		response,
	)
}

// GetBooks godoc
//
// @Summary View All Books
// @Description Returns a paginated list of books sorted by lowest price first
// @Tags Books
// @Produce json
// @Param search query string false "Search by title"
// @Param category query string false "Filter by category"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.BooksListResponse
// @Failure 400 {object} map[string]string
// @Router /books [get]
func (h *BookHandler) GetBooks(
	c *gin.Context,
) {

	// -------------------------
	// Bind query parameters
	// -------------------------

	var query dto.GetBooksQuery

	if err :=
		c.ShouldBindQuery(&query); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	// -------------------------
	// Service
	// -------------------------

	response, err :=
		h.bookService.GetBooks(query)

	if err != nil {

		if err.Error() ==
			"invalid category" {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusOK,
		response,
	)
}

// GetBook godoc
//
// @Summary View One Book
// @Description Returns a single book by its ID
// @Tags Books
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} dto.BookResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books/{id} [get]
func (h *BookHandler) GetBook(
	c *gin.Context,
) {

	// -------------------------
	// Parse ID
	// -------------------------

	bookID, err :=
		strconv.Atoi(
			c.Param("id"),
		)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid book id",
			},
		)

		return
	}

	// -------------------------
	// Service
	// -------------------------

	response, err :=
		h.bookService.GetBookByID(
			uint(bookID),
		)

	if err != nil {

		switch err.Error() {

		case "book not found":

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusOK,
		response,
	)
}

// UpdateBook godoc
//
// @Summary Update Book
// @Description Update an existing book
// @Tags Books
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Book ID"
// @Param title formData string true "Title"
// @Param author formData string true "Author"
// @Param publisher formData string false "Publisher"
// @Param category formData string true "Category"
// @Param price formData number true "Price"
// @Param stock formData int true "Stock"
// @Param image formData file false "Book image"
// @Success 200 {object} dto.BookResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books/{id} [put]
func (h *BookHandler) UpdateBook(
	c *gin.Context,
) {

	// -------------------------
	// Parse Book ID
	// -------------------------

	bookID, err :=
		strconv.Atoi(
			c.Param("id"),
		)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid book id",
			},
		)

		return
	}

	// -------------------------
	// Bind form fields
	// -------------------------

	var request dto.UpdateBookRequest

	if err :=
		c.ShouldBind(&request); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	// -------------------------
	// Optional image upload
	// -------------------------

	imagePath := ""

	_, err =
		c.FormFile("image")

	if err == nil {

		imagePath, err =
			utils.SaveUploadedImage(
				c,
				"image",
			)

		if err != nil {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}
	}

	// -------------------------
	// Service
	// -------------------------

	response,
		oldImagePath,
		err :=
		h.bookService.UpdateBook(
			uint(bookID),
			request,
			imagePath,
		)

	if err != nil {

		// Remove newly uploaded image
		// if database update failed.

		if imagePath != "" {
			_ = os.Remove(imagePath)
		}

		switch err.Error() {

		case "book not found":

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				},
			)

			return

		case "invalid category",
			"price must be greater than 0",
			"stock cannot be negative":

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Delete old image
	// -------------------------

	if imagePath != "" &&
		oldImagePath != "" {

		_ = os.Remove(
			oldImagePath,
		)
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusOK,
		response,
	)
}

// DeleteBook godoc
//	@Summary		Delete book
//	@Description	Soft deletes a book. Admin only.
//	@Tags			Books
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Book ID"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/books/{id} [delete]
func (h *BookHandler) DeleteBook(
	c *gin.Context,
) {

	// -------------------------
	// Parse ID
	// -------------------------

	bookID, err :=
		strconv.ParseUint(
			c.Param("id"),
			10,
			64,
		)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid book id",
			},
		)

		return
	}

	// -------------------------
	// Delete book
	// -------------------------

	err =
		h.bookService.DeleteBook(
			uint(bookID),
		)

	if err != nil {

		switch err.Error() {

		case "book not found":

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusOK,
		dto.MessageResponse{Message: "book deleted successfully"},
	)
}
