package tests

import (
	"bookstore-backend/config"
	"bookstore-backend/database"
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/routes"
	"bookstore-backend/internal/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	router *gin.Engine
	testDB *gorm.DB
)

// TestMain runs ONCE before all tests in this package.
func TestMain(m *testing.M) {

	// Load config
	config.LoadConfig()

	// Connect test database
	testDB = database.ConnectTestDB()

	// Setup router once
	router = gin.Default()
	routes.SetupRoutes(router, testDB)

	// Run all tests
	code := m.Run()

	os.Exit(code)
}

// cleanDatabase runs before each test to reset state
func cleanDatabase() {

	testDB.Exec("DELETE FROM refresh_tokens")
	testDB.Exec("DELETE FROM order_items")
	testDB.Exec("DELETE FROM orders")
	testDB.Exec("DELETE FROM books")
	testDB.Exec("DELETE FROM users")
}

func createTestUser(t *testing.T) models.User {
	hashedPassword, err := utils.HashPassword("password123")
	assert.NoError(t, err)

	user := models.User{
		Name:         "Ahmed",
		Email:        "ahmed@test.com",
		PasswordHash: hashedPassword,
		PhoneNumber:  "01012345678",
		Role:         "user",
	}

	err = testDB.Create(&user).Error
	assert.NoError(t, err)

	return user
}

func createRefreshToken(
	t *testing.T,
	user models.User,
) string {

	tokenString,
		expiresAt,
		err :=
		utils.GenerateRefreshToken(
			user.ID,
		)

	assert.NoError(t, err)

	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}

	err = testDB.Create(&refreshToken).Error
	assert.NoError(t, err)

	return tokenString
}

func authenticatedRequest(
	t *testing.T,
	method string,
	url string,
	token string,
) *http.Request {

	req := httptest.NewRequest(
		method,
		url,
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	return req
}

func createTestOrder(
	t *testing.T,
	userID uint,
	status string,
) models.Order {

	// -------------------------
	// Create a real book
	// -------------------------

	book := models.Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Publisher: "Test Publisher",
		Category:  "Programming",
		Price:     100,
		Stock:     10,
		ImagePath: "",
	}

	err := testDB.Create(&book).Error
	assert.NoError(t, err)

	// -------------------------
	// Create order
	// -------------------------

	order := models.Order{
		UserID:          userID,
		Status:          status,
		ShippingAddress: "Cairo, Egypt",
		TotalPrice:      100,
	}

	err = testDB.Create(&order).Error
	assert.NoError(t, err)

	// -------------------------
	// Create order item
	// -------------------------

	orderItem := models.OrderItem{
		OrderID:  order.ID,
		BookID:   book.ID,
		Quantity: 1,
		Price:    book.Price,
	}

	err = testDB.Create(&orderItem).Error
	assert.NoError(t, err)

	return order
}

func createMultipartBookRequest(
	t *testing.T,
	token string,
	imagePath string,
) *http.Request {

	// Buffer that will contain the multipart body.
	body := &bytes.Buffer{}

	// Multipart writer builds the form-data request.
	writer := multipart.NewWriter(body)

	// -------------------------
	// Add form fields
	// -------------------------

	err := writer.WriteField(
		"title",
		"Clean Code",
	)
	assert.NoError(t, err)

	err = writer.WriteField(
		"author",
		"Robert C. Martin",
	)
	assert.NoError(t, err)

	err = writer.WriteField(
		"publisher",
		"Prentice Hall",
	)
	assert.NoError(t, err)

	err = writer.WriteField(
		"category",
		"Programming",
	)
	assert.NoError(t, err)

	err = writer.WriteField(
		"price",
		"500",
	)
	assert.NoError(t, err)

	err = writer.WriteField(
		"stock",
		"10",
	)
	assert.NoError(t, err)

	// -------------------------
	// Add image
	// -------------------------

	fileWriter, err :=
		writer.CreateFormFile(
			"image",
			filepath.Base(imagePath),
		)

	assert.NoError(t, err)

	file, err :=
		os.Open(imagePath)

	assert.NoError(t, err)

	defer file.Close()

	_, err = io.Copy(
		fileWriter,
		file,
	)

	assert.NoError(t, err)

	// Close the multipart writer.
	err = writer.Close()
	assert.NoError(t, err)

	// -------------------------
	// Build HTTP request
	// -------------------------

	req := httptest.NewRequest(
		http.MethodPost,
		"/books",
		body,
	)

	req.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	if token != "" {

		req.Header.Set(
			"Authorization",
			"Bearer "+token,
		)
	}

	return req
}

func getAdminToken(t *testing.T) string {

	var admin models.User

	err := testDB.
		Where("email = ?", config.AppConfig.AdminEmail).
		First(&admin).Error

	assert.NoError(t, err)

	token, err := utils.GenerateAccessToken(admin.ID, admin.Role)
	assert.NoError(t, err)

	return token
}

func createTestBooks(t *testing.T) {

	books := []models.Book{
		{
			Title:    "Clean Code",
			Author:   "Robert Martin",
			Category: "Programming",
			Price:    500,
			Stock:    5,
		},
		{
			Title:    "The Go Way",
			Author:   "Unknown",
			Category: "Programming",
			Price:    300,
			Stock:    0,
		},
		{
			Title:    "History of Egypt",
			Author:   "Ahmed",
			Category: "History",
			Price:    200,
			Stock:    10,
		},
	}

	for _, b := range books {
		err := testDB.Create(&b).Error
		assert.NoError(t, err)
	}
}

func createUpdateBookRequest(
	t *testing.T,
	token string,
	bookID uint,
	body map[string]string,
	imagePath string,
) *http.Request {

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add fields
	for k, v := range body {
		_ = writer.WriteField(k, v)
	}

	// Add image if exists
	if imagePath != "" {
		file, err := os.Open(imagePath)
		assert.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
		assert.NoError(t, err)

		_, err = io.Copy(part, file)
		assert.NoError(t, err)
	}

	writer.Close()

	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/books/%d", bookID),
		&buf,
	)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	return req
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func createTestBook(t *testing.T) models.Book {

	// -------------------------
	// Source image (test fixture)
	// -------------------------
	sourcePath := "testdata/sample.jpeg"

	// -------------------------
	// Destination path (what your API would normally generate)
	// IMPORTANT: in real app this comes from SaveUploadedImage()
	// -------------------------
	destDir := "uploads/books"

	// Ensure folder exists
	err := os.MkdirAll(destDir, os.ModePerm)
	assert.NoError(t, err)

	// Create unique file name to avoid collisions
	fileName := fmt.Sprintf("test_%d.jpeg", time.Now().UnixNano())

	destPath := filepath.Join(destDir, fileName)

	// -------------------------
	// Copy file into uploads folder
	// -------------------------
	srcFile, err := os.Open(sourcePath)
	assert.NoError(t, err)
	defer srcFile.Close()

	dstFile, err := os.Create(destPath)
	assert.NoError(t, err)
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	assert.NoError(t, err)

	// -------------------------
	// Create DB record
	// -------------------------
	book := models.Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Publisher: "Test Publisher",
		Category:  "Programming",
		Price:     100,
		Stock:     10,
		ImagePath: destPath,
	}

	err = testDB.Create(&book).Error
	assert.NoError(t, err)

	return book
}

func createDeleteBookRequest(
	t *testing.T,
	token string,
	bookID uint,
) *http.Request {

	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/books/%d", bookID),
		nil,
	)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

func createOrderRequest(payload dto.CreateOrderRequest, token string) *http.Request {

	// Convert request body to JSON
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

func seedBook(title string, stock int, price float64) models.Book {

	book := models.Book{
		Title:     title,
		Author:    "Author",
		Publisher: "Publisher",
		Category:  "Programming",
		Price:     price,
		Stock:     stock,
	}

	testDB.Create(&book)

	return book
}

func createUserOrder(
	t *testing.T,
	userID uint,
	status string,
	totalPrice float64,
	quantity int,
) models.Order {

	// -------------------------
	// Create Book
	// -------------------------

	book := models.Book{
		Title:     "Clean Code",
		Author:    "Robert Martin",
		Category:  "Programming",
		Price:     100,
		Stock:     100,
		Publisher: "Prentice Hall",
	}

	err := testDB.Create(&book).Error
	assert.NoError(t, err)

	// -------------------------
	// Create Order
	// -------------------------

	order := models.Order{
		UserID:          userID,
		Status:          status,
		ShippingAddress: "Cairo",
		TotalPrice:      totalPrice,
	}

	err = testDB.Create(&order).Error
	assert.NoError(t, err)

	// -------------------------
	// Create Order Item
	// -------------------------

	orderItem := models.OrderItem{
		OrderID:   order.ID,
		BookID:    book.ID,
		Quantity:  quantity,
		Price:     book.Price,
		Title:     book.Title,
		Author:    book.Author,
		Publisher: book.Publisher,
	}

	err = testDB.Create(&orderItem).Error
	assert.NoError(t, err)

	return order
}

func createGetOrderRequest(
	token string,
	orderID uint,
) *http.Request {

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/orders/%d", orderID),
		nil,
	)

	if token != "" {

		req.Header.Set(
			"Authorization",
			"Bearer "+token,
		)
	}

	return req
}

func createAdminViewOrderRequest(
	token string,
	orderID uint,
) *http.Request {

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/admin/orders/%d", orderID),
		nil,
	)

	if token != "" {

		req.Header.Set(
			"Authorization",
			"Bearer "+token,
		)
	}

	return req
}