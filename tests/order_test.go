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
	"time"

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

func TestGetMyOrdersNoToken(t *testing.T) {

	cleanDatabase()

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusUnauthorized,
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
		"authorization header is required",
		response["error"],
	)
}

func TestGetMyOrdersAdminForbidden(t *testing.T) {

	cleanDatabase()

	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	token := getAdminToken(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusForbidden,
		rec.Code,
	)

	var response map[string]string

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"forbidden",
		response["error"],
	)
}

func TestGetMyOrdersEmpty(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	var response dto.UserOrdersResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Empty(
		t,
		response.Orders,
	)

	assert.Equal(
		t,
		int64(0),
		response.Total,
	)

	assert.Equal(
		t,
		1,
		response.Page,
	)

	assert.Equal(
		t,
		10,
		response.PageSize,
	)
}

func TestGetMyOrdersSortedLatestFirst(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	oldOrder :=
		createUserOrder(
			t,
			user.ID,
			"pending",
			100,
			1,
		)

	time.Sleep(time.Second)

	newOrder :=
		createUserOrder(
			t,
			user.ID,
			"confirmed",
			200,
			2,
		)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	var response dto.UserOrdersResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Len(
		t,
		response.Orders,
		2,
	)

	assert.Equal(
		t,
		newOrder.ID,
		response.Orders[0].ID,
	)

	assert.Equal(
		t,
		oldOrder.ID,
		response.Orders[1].ID,
	)
}

func TestGetMyOrdersPagination(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	for i := 0; i < 15; i++ {

		createUserOrder(
			t,
			user.ID,
			"pending",
			100,
			1,
		)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders?page=2&page_size=5",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	var response dto.UserOrdersResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Len(
		t,
		response.Orders,
		5,
	)

	assert.Equal(
		t,
		2,
		response.Page,
	)

	assert.Equal(
		t,
		5,
		response.PageSize,
	)

	assert.Equal(
		t,
		int64(15),
		response.Total,
	)
}

func TestGetMyOrdersSuccess(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	order :=
		createUserOrder(
			t,
			user.ID,
			"pending",
			300,
			3,
		)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

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
	// Response DTO
	// -------------------------

	var response dto.UserOrdersResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Len(
		t,
		response.Orders,
		1,
	)

	assert.Equal(
		t,
		int64(1),
		response.Total,
	)

	assert.Equal(
		t,
		1,
		response.Page,
	)

	assert.Equal(
		t,
		10,
		response.PageSize,
	)

	orderResponse :=
		response.Orders[0]

	assert.Equal(
		t,
		order.ID,
		orderResponse.ID,
	)

	assert.Equal(
		t,
		"pending",
		orderResponse.Status,
	)

	assert.Equal(
		t,
		300.0,
		orderResponse.TotalPrice,
	)

	// total quantity
	assert.Equal(
		t,
		3,
		orderResponse.ItemsCount,
	)

	assert.False(
		t,
		orderResponse.CreatedAt.IsZero(),
	)
}

func TestGetMyOrder_NoToken(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	req := createGetOrderRequest(
		"",
		order.ID,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert
	// -------------------------

	assert.Equal(
		t,
		http.StatusUnauthorized,
		rec.Code,
	)
}

func TestGetMyOrder_AdminForbidden(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	seeder.SeedAdmin(testDB)

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	adminToken :=
		getAdminToken(t)

	req := createGetOrderRequest(
		adminToken,
		order.ID,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert
	// -------------------------

	assert.Equal(
		t,
		http.StatusForbidden,
		rec.Code,
	)
}

func TestGetMyOrder_OrderNotFound(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(
		t,
		err,
	)

	req := createGetOrderRequest(
		token,
		999999,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert
	// -------------------------

	assert.Equal(
		t,
		http.StatusNotFound,
		rec.Code,
	)
}

func TestGetMyOrder_OrderBelongsToAnotherUser(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user1 := createTestUser(t)

	user2 := models.User{
		Name:         "Omar",
		Email:        "omar@test.com",
		PasswordHash: user1.PasswordHash,
		PhoneNumber:  "01000000000",
		Role:         "user",
	}

	err := testDB.Create(&user2).Error
	assert.NoError(t, err)

	order := createUserOrder(
		t,
		user2.ID,
		"pending",
		100,
		1,
	)

	token, err := utils.GenerateAccessToken(
		user1.ID,
		user1.Role,
	)

	assert.NoError(t, err)

	req := createGetOrderRequest(
		token,
		order.ID,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert
	// -------------------------

	assert.Equal(
		t,
		http.StatusNotFound,
		rec.Code,
	)
}

func TestGetMyOrder_Success(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	book := models.Book{
		Title:     "Clean Code",
		Author:    "Robert Martin",
		Publisher: "Prentice Hall",
		Category:  "Programming",
		Price:     100,
		Stock:     20,
		ImagePath: "/uploads/books/clean-code.jpg",
	}

	err := testDB.Create(&book).Error
	assert.NoError(t, err)

	order := models.Order{
		UserID:          user.ID,
		Status:          "pending",
		ShippingAddress: "Cairo",
		TotalPrice:      200,
	}

	err = testDB.Create(&order).Error
	assert.NoError(t, err)

	item := models.OrderItem{
		OrderID:   order.ID,
		BookID:    book.ID,
		Quantity:  2,
		Price:     book.Price,
		Title:     book.Title,
		Author:    book.Author,
		Publisher: book.Publisher,
		ImagePath: book.ImagePath,
	}

	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	token, err := utils.GenerateAccessToken(
		user.ID,
		user.Role,
	)

	assert.NoError(t, err)

	req := createGetOrderRequest(
		token,
		order.ID,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	// -------------------------
	// Parse Response
	// -------------------------

	var response dto.UserOrderDetailsResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	// -------------------------
	// Order Fields
	// -------------------------

	assert.Equal(
		t,
		order.ID,
		response.ID,
	)

	assert.Equal(
		t,
		order.Status,
		response.Status,
	)

	assert.Equal(
		t,
		order.ShippingAddress,
		response.Address,
	)

	assert.Equal(
		t,
		order.TotalPrice,
		response.TotalPrice,
	)

	assert.False(
		t,
		response.CreatedAt.IsZero(),
	)

	assert.False(
		t,
		response.UpdatedAt.IsZero(),
	)

	// -------------------------
	// Order Items
	// -------------------------

	assert.Len(
		t,
		response.Items,
		1,
	)

	responseItem := response.Items[0]

	assert.Equal(
		t,
		item.ID,
		responseItem.ID,
	)

	assert.Equal(
		t,
		item.BookID,
		responseItem.BookID,
	)

	assert.Equal(
		t,
		item.Quantity,
		responseItem.Quantity,
	)

	assert.Equal(
		t,
		item.Price,
		responseItem.Price,
	)

	assert.Equal(
		t,
		item.Title,
		responseItem.Title,
	)

	assert.Equal(
		t,
		item.Author,
		responseItem.Author,
	)

	assert.Equal(
		t,
		item.Publisher,
		responseItem.Publisher,
	)

	assert.Equal(
		t,
		item.ImagePath,
		responseItem.ImagePath,
	)
}