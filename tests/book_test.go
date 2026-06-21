package tests

import (
	"bookstore-backend/config"
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/seeder"
	"bookstore-backend/internal/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBookSuccess(t *testing.T) {

	// -------------------------
	// ARRANGE
	// -------------------------

	cleanDatabase()

	// Seed admin using real app seeder
	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	// Get admin from DB
	var admin models.User
	err = testDB.
		Where("email = ?", config.AppConfig.AdminEmail).
		First(&admin).Error
	assert.NoError(t, err)

	// Generate admin access token
	token, err := utils.GenerateAccessToken(admin.ID, admin.Role)
	assert.NoError(t, err)

	// Create real test image
	imagePath := "testdata/sample.jpeg"

	// Ensure file exists
	_, err = os.Stat(imagePath)
	assert.NoError(t, err)

	// Build request
	req := createMultipartBookRequest(
		t,
		token,
		imagePath,
	)

	recorder := httptest.NewRecorder()

	// -------------------------
	// ACT
	// -------------------------

	router.ServeHTTP(recorder, req)

	// -------------------------
	// ASSERT RESPONSE
	// -------------------------

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response dto.BookResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Clean Code", response.Title)
	assert.Equal(t, "Robert C. Martin", response.Author)
	assert.Equal(t, "Prentice Hall", response.Publisher)
	assert.Equal(t, "Programming", response.Category)
	assert.Equal(t, float64(500), response.Price)
	assert.Equal(t, 10, response.Stock)

	// Image must be returned
	assert.NotEmpty(t, response.ImageURL)

	assert.NotEmpty(t, response.CreatedAt)

	// -------------------------
	// ASSERT DATABASE
	// -------------------------

	var book models.Book
	err = testDB.
		Where("id = ?", response.ID).
		First(&book).Error

	assert.NoError(t, err)

	assert.Equal(t, response.Title, book.Title)
	assert.Equal(t, response.Author, book.Author)
	assert.Equal(t, response.Category, book.Category)
	assert.Equal(t, response.Price, book.Price)
	assert.Equal(t, response.Stock, book.Stock)
	assert.Equal(t, response.Publisher, book.Publisher)
	assert.NotEmpty(t, book.ImagePath)
	assert.NotEmpty(t, book.CreatedAt)

	// -------------------------
	// ASSERT FILE SYSTEM
	// -------------------------

	_, err = os.Stat(book.ImagePath)
	assert.NoError(t, err)

	// -------------------------
	// CLEANUP
	// -------------------------

	_ = os.Remove(book.ImagePath)
}

func TestCreateBookNoToken(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()

	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	imagePath := "testdata/sample.jpeg"

	req := createMultipartBookRequest(
		t,
		"", // No token
		imagePath,
	)

	recorder := httptest.NewRecorder()

	entriesBefore, _ := os.ReadDir("uploads/books")
	beforeCount := len(entriesBefore)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(recorder, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(
		t,
		http.StatusUnauthorized,
		recorder.Code,
	)

	// -------------------------
	// Assert Response
	// -------------------------

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"authorization header is required",
		response["error"],
	)

	// -------------------------
	// Assert Database
	// -------------------------

	var count int64

	err = testDB.
		Model(&models.Book{}).
		Count(&count).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		int64(0),
		count,
	)

	// -------------------------
	// Assert No Image Saved
	// -------------------------

	entriesAfter, _ := os.ReadDir("uploads/books")
	afterCount := len(entriesAfter)

	assert.Equal(t, beforeCount, afterCount)
}

func TestCreateBookNotAdmin(t *testing.T) {

	cleanDatabase()

	// Create normal user
	user := createTestUser(t)

	token, err := utils.GenerateAccessToken(user.ID, user.Role)
	assert.NoError(t, err)

	imagePath := "testdata/sample.jpeg"

	req := createMultipartBookRequest(
		t,
		token,
		imagePath,
	)

	rec := httptest.NewRecorder()

	entriesBefore, _ := os.ReadDir("uploads/books")
	beforeCount := len(entriesBefore)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "forbidden", resp["error"])

	// -------------------------
	// Assert Database
	// -------------------------

	var count int64

	err = testDB.
		Model(&models.Book{}).
		Count(&count).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		int64(0),
		count,
	)

	// -------------------------
	// Assert No Image Saved
	// -------------------------

	entriesAfter, _ := os.ReadDir("uploads/books")
	afterCount := len(entriesAfter)

	assert.Equal(t, beforeCount, afterCount)
}

func TestCreateBookInvalidDTO(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// missing title intentionally
	writer.WriteField("author", "Author")
	writer.WriteField("publisher", "Pub")
	writer.WriteField("category", "Programming")
	writer.WriteField("price", "100")
	writer.WriteField("stock", "5")

	fileWriter, _ := writer.CreateFormFile("image", "sample.jpeg")

	file, _ := os.Open("testdata/sample.jpeg")
	defer file.Close()

	io.Copy(fileWriter, file)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	entriesBefore, _ := os.ReadDir("uploads/books")
	beforeCount := len(entriesBefore)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// -------------------------
	// Assert Database
	// -------------------------

	var count int64

	err := testDB.
		Model(&models.Book{}).
		Count(&count).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		int64(0),
		count,
	)

	// -------------------------
	// Assert No Image Saved
	// -------------------------

	entriesAfter, _ := os.ReadDir("uploads/books")
	afterCount := len(entriesAfter)

	assert.Equal(t, beforeCount, afterCount)
}

func TestCreateBookInvalidCategory(t *testing.T) {

	cleanDatabase()

	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	token := getAdminToken(t)

	// -------------------------
	// Arrange
	// -------------------------

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	// Valid fields except category.
	err = writer.WriteField("title", "Clean Code")
	assert.NoError(t, err)

	err = writer.WriteField("author", "Robert C. Martin")
	assert.NoError(t, err)

	err = writer.WriteField("publisher", "Prentice Hall")
	assert.NoError(t, err)

	err = writer.WriteField("category", "INVALID_CATEGORY")
	assert.NoError(t, err)

	err = writer.WriteField("price", "500")
	assert.NoError(t, err)

	err = writer.WriteField("stock", "10")
	assert.NoError(t, err)

	// Add real image.
	fileWriter, err := writer.CreateFormFile(
		"image",
		"sample.jpeg",
	)
	assert.NoError(t, err)

	file, err := os.Open("testdata/sample.jpeg")
	assert.NoError(t, err)

	defer file.Close()

	_, err = io.Copy(fileWriter, file)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/books",
		body,
	)

	req.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	// Count uploaded files before request.
	entriesBefore, err := os.ReadDir("uploads/books")
	assert.NoError(t, err)

	beforeCount := len(entriesBefore)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(
		t,
		http.StatusBadRequest,
		rec.Code,
	)

	// -------------------------
	// Assert Response
	// -------------------------

	var response map[string]string

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"invalid category",
		response["error"],
	)

	// -------------------------
	// Assert Database
	// -------------------------

	var count int64

	err = testDB.
		Model(&models.Book{}).
		Count(&count).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		int64(0),
		count,
	)

	// -------------------------
	// Assert Image Cleanup
	// -------------------------

	entriesAfter, err := os.ReadDir("uploads/books")
	assert.NoError(t, err)

	afterCount := len(entriesAfter)

	// The handler should have removed
	// the uploaded image after the
	// service returned an error.
	assert.Equal(
		t,
		beforeCount,
		afterCount,
	)
}

func TestCreateBookInvalidPrice(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("title", "Book")
	writer.WriteField("author", "Author")
	writer.WriteField("publisher", "Pub")
	writer.WriteField("category", "Programming")
	writer.WriteField("price", "0") // invalid
	writer.WriteField("stock", "5")

	fileWriter, _ := writer.CreateFormFile("image", "sample.jpeg")

	file, _ := os.Open("testdata/sample.jpeg")
	defer file.Close()

	io.Copy(fileWriter, file)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	entriesBefore, _ := os.ReadDir("uploads/books")
	beforeCount := len(entriesBefore)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)

	// assert.Equal(t, "price must be greater than 0", resp["error"])

	// -------------------------
	// Assert Database
	// -------------------------

	var count int64

	err := testDB.
		Model(&models.Book{}).
		Count(&count).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		int64(0),
		count,
	)

	// -------------------------
	// Assert No Image Saved
	// -------------------------

	entriesAfter, _ := os.ReadDir("uploads/books")
	afterCount := len(entriesAfter)

	assert.Equal(t, beforeCount, afterCount)
}

func TestCreateBookInvalidStock(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("title", "Book")
	writer.WriteField("author", "Author")
	writer.WriteField("publisher", "Pub")
	writer.WriteField("category", "Programming")
	writer.WriteField("price", "100")
	writer.WriteField("stock", "-1") // invalid

	fileWriter, _ := writer.CreateFormFile("image", "sample.jpeg")

	file, _ := os.Open("testdata/sample.jpeg")
	defer file.Close()

	io.Copy(fileWriter, file)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	entriesBefore, _ := os.ReadDir("uploads/books")
	beforeCount := len(entriesBefore)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "stock cannot be negative", resp["error"])

	// -------------------------
	// Assert Database
	// -------------------------

	var count int64

	err := testDB.
		Model(&models.Book{}).
		Count(&count).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		int64(0),
		count,
	)

	// -------------------------
	// Assert No Image Saved
	// -------------------------

	entriesAfter, _ := os.ReadDir("uploads/books")
	afterCount := len(entriesAfter)

	assert.Equal(t, beforeCount, afterCount)
}

func TestGetBooksSuccess(t *testing.T) {

	cleanDatabase()
	createTestBooks(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/books",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// -------------------------
	// Status check
	// -------------------------

	assert.Equal(t, http.StatusOK, rec.Code)

	// -------------------------
	// Response parsing
	// -------------------------

	var resp dto.BooksListResponse

	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// -------------------------
	// Basic assertions
	// -------------------------

	assert.Equal(t, 3, len(resp.Books))
	assert.Equal(t, 1, resp.Page)      // default
	assert.Equal(t, 10, resp.PageSize) // default
	assert.Equal(t, 3, int(resp.Total))

	// -------------------------
	// Check sorting (price ASC)
	// -------------------------

	for i := 1; i < len(resp.Books); i++ {
		assert.True(t,
			resp.Books[i-1].Price <= resp.Books[i].Price,
		)
	}
}

func TestGetBooksSearchByTitle(t *testing.T) {

	cleanDatabase()
	createTestBooks(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/books?search=clean",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.BooksListResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	// Only "Clean Code" should match
	assert.Equal(t, 1, len(resp.Books))
	assert.Equal(t, "Clean Code", resp.Books[0].Title)

	assert.Equal(t, 1, resp.Page)      // default
	assert.Equal(t, 10, resp.PageSize) // default
	assert.Equal(t, 1, int(resp.Total))
}

func TestGetBooksFilterByCategory(t *testing.T) {

	cleanDatabase()
	createTestBooks(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/books?category=History",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.BooksListResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, 1, len(resp.Books))
	assert.Equal(t, "History", resp.Books[0].Category)
	assert.Equal(t, 1, resp.Page)      // default
	assert.Equal(t, 10, resp.PageSize) // default
	assert.Equal(t, 1, int(resp.Total))
}

func TestGetBooksPagination(t *testing.T) {

	cleanDatabase()
	createTestBooks(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/books?page=1&page_size=2",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.BooksListResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	// Only 2 books returned
	assert.Equal(t, 2, len(resp.Books))

	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 2, resp.PageSize)
	assert.Equal(t, int64(3), resp.Total)
}

func TestGetBooksInvalidCategory(t *testing.T) {

	cleanDatabase()

	req := httptest.NewRequest(
		http.MethodGet,
		"/books?category=INVALID",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "invalid category", resp["error"])
}

func TestGetBooksInvalidPaginationDefaultsApplied(t *testing.T) {

	cleanDatabase()
	createTestBooks(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/books?page=-1&page_size=-10",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.BooksListResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	// Should fallback to defaults
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
	assert.Equal(t, 3, len(resp.Books))
	assert.Equal(t, int64(3), resp.Total)
}

func TestGetBooksEmptyList(t *testing.T) {

	cleanDatabase()

	req := httptest.NewRequest(
		http.MethodGet,
		"/books",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.BooksListResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, 0, len(resp.Books))
	assert.Equal(t, int64(0), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
}

func TestGetBooksResponseStructure(t *testing.T) {

	cleanDatabase()
	createTestBooks(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/books",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.BooksListResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	for _, b := range resp.Books {

		assert.NotZero(t, b.ID)
		assert.NotEmpty(t, b.Title)
		assert.NotEmpty(t, b.Author)
		assert.NotEmpty(t, b.Category)
		assert.GreaterOrEqual(t, b.Price, 0.0)

		// derived field correctness
		if b.IsOutOfStock {
			assert.Equal(t, true, b.IsOutOfStock)
		} else {
			assert.Equal(t, false, b.IsOutOfStock)
		}
	}

	assert.Equal(t, 3, len(resp.Books))
	assert.Equal(t, int64(3), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
}

func TestGetBookSuccess(t *testing.T) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	book := models.Book{
		Title:     "Clean Code",
		Author:    "Robert C. Martin",
		Publisher: "Prentice Hall",
		Category:  "Programming",
		Price:     500,
		Stock:     8,
		ImagePath: "/uploads/books/test.jpg",
	}

	err := testDB.Create(&book).Error
	assert.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/books/%d", book.ID),
		nil,
	)

	rec := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Status
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	// -------------------------
	// Response
	// -------------------------

	var response dto.BookResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(t, book.ID, response.ID)
	assert.Equal(t, book.Title, response.Title)
	assert.Equal(t, book.Author, response.Author)
	assert.Equal(t, book.Publisher, response.Publisher)
	assert.Equal(t, book.Category, response.Category)
	assert.Equal(t, book.Price, response.Price)
	assert.Equal(t, book.Stock, response.Stock)
	assert.Equal(t, book.ImagePath, response.ImageURL)

	// CreatedAt should be populated.
	assert.False(
		t,
		response.CreatedAt.IsZero(),
	)
}

func TestGetBookNotFound(t *testing.T) {

	cleanDatabase()

	req := httptest.NewRequest(
		http.MethodGet,
		"/books/99999",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusNotFound,
		rec.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"book not found",
		response["error"],
	)
}

func TestGetBookInvalidID(t *testing.T) {

	cleanDatabase()

	req := httptest.NewRequest(
		http.MethodGet,
		"/books/abc",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusBadRequest,
		rec.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"invalid book id",
		response["error"],
	)
}

func TestUpdateBookInvalidID(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	req := createUpdateBookRequest(
		t,
		token,
		999,
		map[string]string{
			"title": "New Title",
			"author": "Author",
			"category": "Programming",
			"price": "100",
			"stock": "10",
		},
		"",
	)

	rec := httptest.NewRecorder()

	// Count files before request
	beforeFiles, _ := os.ReadDir("uploads/books")
	beforeCount := len(beforeFiles)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "book not found", resp["error"])

	// -------------------------
	// Assert NO new image saved
	// -------------------------

	afterFiles, _ := os.ReadDir("uploads/books")
	afterCount := len(afterFiles)

	assert.Equal(t, beforeCount, afterCount)
}

func TestUpdateBookNoToken(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	// Create real book with real image
	book := createTestBook(t)

	oldImagePath := book.ImagePath

	// Prepare update data
	body := map[string]string{
		"title":    "Updated Title",
		"author":   "Updated Author",
		"category": "Programming",
		"price":    "999",
		"stock":    "50",
	}

	// We simulate uploading a new image
	newImage := "testdata/sample.jpeg"

	req := createUpdateBookRequest(
		t,
		"", // ❌ NO TOKEN
		book.ID,
		body,
		newImage,
	)

	rec := httptest.NewRecorder()

	// Count files before request
	beforeFiles, _ := os.ReadDir("uploads/books")
	beforeCount := len(beforeFiles)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// -------------------------
	// Assert DB (book NOT updated)
	// -------------------------

	var dbBook models.Book
	err := testDB.First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	// Ensure all fields are unchanged
	assert.Equal(t, book.Title, dbBook.Title)
	assert.Equal(t, book.Author, dbBook.Author)
	assert.Equal(t, book.Price, dbBook.Price)
	assert.Equal(t, book.Stock, dbBook.Stock)
	assert.Equal(t, book.Category, dbBook.Category)
	assert.Equal(t, book.ImagePath, dbBook.ImagePath)

	// -------------------------
	// Assert old image still exists
	// -------------------------

	assert.True(t, fileExists(oldImagePath))

	// -------------------------
	// Assert NO new image saved
	// -------------------------

	afterFiles, _ := os.ReadDir("uploads/books")
	afterCount := len(afterFiles)

	assert.Equal(t, beforeCount, afterCount)
}

func TestUpdateBookNotAdmin(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	// Create normal user (NOT admin)
	user := createTestUser(t)

	userToken, err :=
		utils.GenerateAccessToken(user.ID, user.Role)
	assert.NoError(t, err)

	// Create real book with image
	book := createTestBook(t)

	oldImagePath := book.ImagePath

	// Update payload
	body := map[string]string{
		"title":    "Updated Title",
		"author":   "Updated Author",
		"category": "Programming",
		"price":    "999",
		"stock":    "50",
	}

	// Try uploading new image
	newImage := "testdata/sample.jpeg"

	req := createUpdateBookRequest(
		t,
		userToken, // ❌ NOT admin token
		book.ID,
		body,
		newImage,
	)

	rec := httptest.NewRecorder()

	// Snapshot file count before request
	beforeFiles, _ := os.ReadDir("uploads/books")
	beforeCount := len(beforeFiles)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusForbidden, rec.Code)

	// -------------------------
	// Assert DB not modified
	// -------------------------

	var dbBook models.Book
	err = testDB.First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.Equal(t, book.Title, dbBook.Title)
	assert.Equal(t, book.Author, dbBook.Author)
	assert.Equal(t, book.Price, dbBook.Price)
	assert.Equal(t, book.Stock, dbBook.Stock)
	assert.Equal(t, book.Category, dbBook.Category)
	assert.Equal(t, book.ImagePath, dbBook.ImagePath)

	// -------------------------
	// Assert old image still exists
	// -------------------------

	assert.True(t, fileExists(oldImagePath))

	// -------------------------
	// Assert no new image saved
	// -------------------------

	afterFiles, _ := os.ReadDir("uploads/books")
	afterCount := len(afterFiles)

	assert.Equal(t, beforeCount, afterCount)
}

func TestUpdateBookValidationFail(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	book := createTestBook(t)

	oldImagePath := book.ImagePath

	body := map[string]string{
		"title":    "", // ❌ invalid
		"author":   "Updated Author",
		"category": "Programming",
		"price":    "100",
		"stock":    "10",
	}

	newImage := "testdata/sample.jpeg"

	req := createUpdateBookRequest(
		t,
		token,
		book.ID,
		body,
		newImage,
	)

	rec := httptest.NewRecorder()

	beforeFiles, _ := os.ReadDir("uploads/books")
	beforeCount := len(beforeFiles)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// -------------------------
	// Assert DB unchanged
	// -------------------------

	var dbBook models.Book
	err := testDB.First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.Equal(t, book.Title, dbBook.Title)
	assert.Equal(t, book.ImagePath, dbBook.ImagePath)

	// -------------------------
	// Assert old image still exists
	// -------------------------

	assert.True(t, fileExists(oldImagePath))

	// -------------------------
	// Assert no new image saved
	// -------------------------

	afterFiles, _ := os.ReadDir("uploads/books")
	afterCount := len(afterFiles)

	assert.Equal(t, beforeCount, afterCount)
}

func TestUpdateBookInvalidCategory(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	book := createTestBook(t)

	oldImagePath := book.ImagePath

	body := map[string]string{
		"title":    "Updated Title",
		"author":   "Updated Author",
		"category": "INVALID_CATEGORY", // ❌
		"price":    "100",
		"stock":    "10",
	}

	newImage := "testdata/sample.jpeg"

	req := createUpdateBookRequest(
		t,
		token,
		book.ID,
		body,
		newImage,
	)

	rec := httptest.NewRecorder()

	beforeFiles, _ := os.ReadDir("uploads/books")
	beforeCount := len(beforeFiles)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// -------------------------
	// Assert DB unchanged
	// -------------------------

	var dbBook models.Book
	err := testDB.First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.Equal(t, book.Category, dbBook.Category)
	assert.Equal(t, book.ImagePath, dbBook.ImagePath)

	// -------------------------
	// Assert old image still exists
	// -------------------------

	assert.True(t, fileExists(oldImagePath))

	// -------------------------
	// Assert no new image saved
	// -------------------------

	afterFiles, _ := os.ReadDir("uploads/books")
	afterCount := len(afterFiles)

	assert.Equal(t, beforeCount, afterCount)
}

func TestUpdateBookInvalidPrice(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	book := createTestBook(t)

	oldImagePath := book.ImagePath

	body := map[string]string{
		"title":    "Updated Title",
		"author":   "Updated Author",
		"category": "Programming",
		"price":    "0", // ❌ invalid
		"stock":    "10",
	}

	newImage := "testdata/sample.jpeg"

	req := createUpdateBookRequest(
		t,
		token,
		book.ID,
		body,
		newImage,
	)

	rec := httptest.NewRecorder()

	beforeFiles, _ := os.ReadDir("uploads/books")
	beforeCount := len(beforeFiles)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// -------------------------
	// Assert DB unchanged
	// -------------------------

	var dbBook models.Book
	err := testDB.First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.Equal(t, book.Price, dbBook.Price)
	assert.Equal(t, book.ImagePath, dbBook.ImagePath)

	// -------------------------
	// Assert old image still exists
	// -------------------------

	assert.True(t, fileExists(oldImagePath))

	// -------------------------
	// Assert no new image saved
	// -------------------------

	afterFiles, _ := os.ReadDir("uploads/books")
	afterCount := len(afterFiles)

	assert.Equal(t, beforeCount, afterCount)
}

func TestUpdateBookInvalidStock(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	book := createTestBook(t)

	oldImagePath := book.ImagePath

	body := map[string]string{
		"title":    "Updated Title",
		"author":   "Updated Author",
		"category": "Programming",
		"price":    "100",
		"stock":    "-1", // ❌ invalid
	}

	newImage := "testdata/sample.jpeg"

	req := createUpdateBookRequest(
		t,
		token,
		book.ID,
		body,
		newImage,
	)

	rec := httptest.NewRecorder()

	beforeFiles, _ := os.ReadDir("uploads/books")
	beforeCount := len(beforeFiles)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// -------------------------
	// Assert DB unchanged
	// -------------------------

	var dbBook models.Book
	err := testDB.First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.Equal(t, book.Stock, dbBook.Stock)
	assert.Equal(t, book.ImagePath, dbBook.ImagePath)

	// -------------------------
	// Assert old image still exists
	// -------------------------

	assert.True(t, fileExists(oldImagePath))

	// -------------------------
	// Assert no new image saved
	// -------------------------

	afterFiles, _ := os.ReadDir("uploads/books")
	afterCount := len(afterFiles)

	assert.Equal(t, beforeCount, afterCount)
}

func TestUpdateBookSuccessWithImage(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	// Create book with real image
	book := createTestBook(t)

	oldImagePath := book.ImagePath

	// Prepare update data
	body := map[string]string{
		"title":    "Updated Title",
		"author":   "Updated Author",
		"category": "Programming",
		"price":    "200",
		"stock":    "25",
	}

	newImage := "testdata/sample.jpeg"

	req := createUpdateBookRequest(
		t,
		token,
		book.ID,
		body,
		newImage,
	)

	rec := httptest.NewRecorder()

	beforeFiles, _ := os.ReadDir("uploads/books")
	beforeCount := len(beforeFiles)

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusOK, rec.Code)

	// -------------------------
	// Assert Response JSON
	// -------------------------

	var resp dto.BookResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "Updated Title", resp.Title)
	assert.Equal(t, "Updated Author", resp.Author)
	assert.Equal(t, "Programming", resp.Category)
	assert.Equal(t, float64(200), resp.Price)
	assert.Equal(t, 25, resp.Stock)
	assert.NotEmpty(t, resp.ImageURL)
	assert.Equal(t, book.ID, resp.ID)

	// -------------------------
	// Assert DB updated
	// -------------------------

	var dbBook models.Book
	err = testDB.First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.Equal(t, "Updated Title", dbBook.Title)
	assert.Equal(t, "Updated Author", dbBook.Author)
	assert.Equal(t, "Programming", dbBook.Category)
	assert.Equal(t, float64(200), dbBook.Price)
	assert.Equal(t, 25, dbBook.Stock)

	// Image should be replaced
	assert.NotEqual(t, oldImagePath, dbBook.ImagePath)

	// -------------------------
	// Assert file system
	// -------------------------

	// Old image must be deleted
	assert.False(t, fileExists(oldImagePath))

	// New image must exist
	assert.True(t, fileExists(dbBook.ImagePath))

	afterFiles, _ := os.ReadDir("uploads/books")
	afterCount := len(afterFiles)

	// File count should remain stable (old removed, new added)
	assert.Equal(t, beforeCount, afterCount)
}

func TestDeleteBookNoToken(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	book := createTestBook(t)
	imagePath := book.ImagePath

	req := createDeleteBookRequest(t, "", book.ID)

	rec := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// -------------------------
	// Assert DB not changed
	// -------------------------

	var dbBook models.Book
	err := testDB.Unscoped().First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.False(t, dbBook.DeletedAt.Valid)

	// -------------------------
	// Assert image still exists
	// -------------------------

	assert.True(t, fileExists(imagePath))
}

func TestDeleteBookNotAdmin(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	userToken, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := createTestBook(t)
	imagePath := book.ImagePath

	req := createDeleteBookRequest(t, userToken, book.ID)

	rec := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusForbidden, rec.Code)

	// -------------------------
	// Assert DB unchanged
	// -------------------------

	var dbBook models.Book
	err := testDB.Unscoped().First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.False(t, dbBook.DeletedAt.Valid)

	// -------------------------
	// Assert image still exists
	// -------------------------

	assert.True(t, fileExists(imagePath))
}

func TestDeleteBookNotFound(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	req := createDeleteBookRequest(t, token, 999999)

	rec := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusNotFound, rec.Code)

	// -------------------------
	// Assert DB empty
	// -------------------------

	var count int64
	testDB.Model(&models.Book{}).Count(&count)

	assert.Equal(t, int64(0), count)
}

func TestDeleteBookSuccess(t *testing.T) {

	// -------------------------
	// Arrange
	// -------------------------

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	book := createTestBook(t)
	imagePath := book.ImagePath

	req := createDeleteBookRequest(t, token, book.ID)

	rec := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(rec, req)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(t, http.StatusOK, rec.Code)

	// -------------------------
	// Assert Response DTO
	// -------------------------

	var resp dto.MessageResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "book deleted successfully", resp.Message)

	// -------------------------
	// Assert DB (soft delete)
	// -------------------------

	var dbBook models.Book
	err = testDB.First(&dbBook, book.ID).Error

	// should not be found in normal queries
	assert.Error(t, err)

	// but exists in unscoped
	err = testDB.Unscoped().First(&dbBook, book.ID).Error
	assert.NoError(t, err)

	assert.True(t, dbBook.DeletedAt.Valid)

	// -------------------------
	// Assert image NOT removed
	// -------------------------

	assert.True(t, fileExists(imagePath))
}