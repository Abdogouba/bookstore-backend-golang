package tests

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/seeder"
	"bookstore-backend/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOrder_NoToken_Fail(t *testing.T) {

	cleanDatabase()
	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
		},
	}, "")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateOrder_Admin_Fail(t *testing.T) {

	cleanDatabase()
	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	adminToken := getAdminToken(t)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
		},
	}, adminToken)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreateOrder_EmptyAddress_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := createTestUser(t)
	userToken, _ := utils.GenerateAccessToken(token.ID, token.Role)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
		},
	}, userToken)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_NoItems_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items:   []dto.CreateOrderItemRequest{},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_InvalidQuantity_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 0},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_DuplicateBooks_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
			{BookID: book.ID, Quantity: 2},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_BookNotFound_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: 99999, Quantity: 1},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCreateOrder_InsufficientStock_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := seedBook("Book 1", 2, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 10},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_Success(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book1 := seedBook("Book 1", 10, 100)
	book2 := seedBook("Book 2", 5, 200)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo, Egypt",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book1.ID, Quantity: 2},
			{BookID: book2.ID, Quantity: 1},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// =========================
	// ASSERT RESPONSE
	// =========================
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response dto.CreateOrderResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "pending", response.Status)
	assert.Equal(t, "Cairo, Egypt", response.Address)
	assert.Equal(t, float64(400), response.TotalPrice) // (2*100 + 1*200)

	assert.Len(t, response.Items, 2)

	// check DTO fields
	assert.NotZero(t, response.ID)
	assert.NotZero(t, response.CreatedAt)

	// =========================
	// ASSERT DATABASE (ORDER)
	// =========================
	var order models.Order
	err = testDB.Preload("OrderItems").
		Where("id = ?", response.ID).
		First(&order).Error

	assert.NoError(t, err)
	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, float64(400), order.TotalPrice)

	// =========================
	// ASSERT ORDER ITEMS
	// =========================
	assert.Len(t, order.OrderItems, 2)

	// =========================
	// ASSERT STOCK REDUCTION
	// =========================
	var updatedBook1 models.Book
	testDB.First(&updatedBook1, book1.ID)
	assert.Equal(t, 8, updatedBook1.Stock)

	var updatedBook2 models.Book
	testDB.First(&updatedBook2, book2.ID)
	assert.Equal(t, 4, updatedBook2.Stock)
}

func TestCreateOrder_TransactionRollback(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := seedBook("Book 1", 10, 100)

	// simulate failure by requesting invalid second item
	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
			{BookID: 99999, Quantity: 1}, // invalid book triggers failure
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	// ORDER SHOULD NOT EXIST
	var count int64
	testDB.Model(&models.Order{}).Count(&count)
	assert.Equal(t, int64(0), count)

	// STOCK SHOULD NOT CHANGE
	var updatedBook models.Book
	testDB.First(&updatedBook, book.ID)
	assert.Equal(t, 10, updatedBook.Stock)
}